package gifts

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SeanMcTex/pressie/internal/contacts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// slugify converts a name into a filesystem-safe slug.
// Lowercase, non-alphanumeric runs become single hyphens, no leading/trailing hyphens.
var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(name)
	s = nonAlphaNum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ResolveContact resolves a name query to a contact key and file path.
// Resolution order:
//   1. Exact slug match in the index (case-insensitive).
//   2. Fuzzy match: stored contact name contains the query, or query contains the stored name (case-insensitive).
//   3. Configured plugins (first match wins).
//   4. Manual fallback: slugify the name, no external service needed.
//
// If the fuzzy pass finds multiple candidates, an error is returned listing
// the matched names so the caller can disambiguate.
//
// Returns the contact key, the contact's display name, and the relative
// file path within the gifts directory (e.g. "_private/kris-mcmains.json").
func ResolveContact(ctx context.Context, giftsDir string, idx *store.IndexFile, name, visibility string) (key, displayName, relPath string, err error) {
	querySlug := slugify(name)
	queryLower := strings.ToLower(name)

	// Pass 1: exact slug match in the index.
	if idx.Contacts != nil {
		for k, m := range idx.Contacts {
			if strings.EqualFold(filepath.Base(m.File), querySlug+".json") {
				storedName := lookupContactName(giftsDir, m.File)
				displayName = storedName
				if displayName == "" {
					displayName = name
				}
				return k, displayName, m.File, nil
			}
		}
	}

	// Pass 2: fuzzy match against stored contact names.
	if idx.Contacts != nil {
		type candidate struct {
			key, name, file string
		}
		var matches []candidate
		for k, m := range idx.Contacts {
			storedName := lookupContactName(giftsDir, m.File)
			if storedName == "" {
				continue
			}
			storedLower := strings.ToLower(storedName)
			if strings.Contains(storedLower, queryLower) || strings.Contains(queryLower, storedLower) {
				matches = append(matches, candidate{k, storedName, m.File})
			}
		}
		switch len(matches) {
		case 0:
			// No fuzzy match, fall through to plugins.
		case 1:
			return matches[0].key, matches[0].name, matches[0].file, nil
		default:
			names := make([]string, len(matches))
			for i, c := range matches {
				names[i] = c.name
			}
			return "", "", "", fmt.Errorf("ambiguous contact %q matches: %s — use the full name", name, strings.Join(names, ", "))
		}
	}

	// Pass 3: try configured plugins.
	for _, pc := range idx.Plugins {
		plugin := &contacts.Plugin{
			Name:      pc.Name,
			Command:   pc.Command,
			TimeoutMs: pc.TimeoutMs,
		}
		pluginMatches, perr := plugin.Resolve(ctx, name)
		if perr != nil {
			// Skip failing plugins, don't abort.
			continue
		}
		if len(pluginMatches) > 1 {
			names := make([]string, len(pluginMatches))
			for i, c := range pluginMatches {
				names[i] = c.Name
			}
			return "", "", "", fmt.Errorf("ambiguous contact %q matches: %s — use the full name", name, strings.Join(names, ", "))
		}
		if len(pluginMatches) == 1 {
			c := pluginMatches[0]
			relPath = filepath.Join(visibilityDir(visibility), slugify(c.Name)+".json")
			return c.Key, c.Name, relPath, nil
		}
	}

	// Pass 4: manual fallback, no plugin needed.
	key = fmt.Sprintf("manual:%s", querySlug)
	relPath = filepath.Join(visibilityDir(visibility), querySlug+".json")
	return key, name, relPath, nil
}

// lookupContactName loads a contact file and returns its Name field.
// Returns empty string if the file can't be read or parsed.
func lookupContactName(giftsDir, relPath string) string {
	cf, err := store.LoadContactFile(filepath.Join(giftsDir, relPath))
	if err != nil {
		return ""
	}
	return cf.Name
}

// EnsureContactFile loads the contact file at relPath, creating an empty
// one if it doesn't exist yet. It populates ContactKey and Name on first
// creation.
func EnsureContactFile(giftsDir, relPath, key, name string) (*store.ContactFile, error) {
	absPath := filepath.Join(giftsDir, relPath)
	cf, err := store.LoadContactFile(absPath)
	if err == nil {
		return cf, nil
	}
	// File doesn't exist or is invalid — create fresh.
	cf = &store.ContactFile{
		ContactKey:    key,
		Name:          name,
		GiftsGiven:    []store.Gift{},
		GiftsReceived: []store.Gift{},
		Ideas:         []store.Idea{},
	}
	return cf, nil
}

// RegisterContact records a contact mapping in the index if not already present.
func RegisterContact(idx *store.IndexFile, key, relPath, visibility string, tags []string) {
	if idx.Contacts == nil {
		idx.Contacts = make(map[string]store.ContactMapping)
	}
	if _, exists := idx.Contacts[key]; exists {
		return
	}
	idx.Contacts[key] = store.ContactMapping{
		File:       relPath,
		Visibility: visibility,
		Tags:       tags,
	}
}

// visibilityDir returns the subdirectory for the given visibility level.
func visibilityDir(visibility string) string {
	if visibility == "shared" {
		return "_shared"
	}
	return "_private"
}

// IsPluginConfigured returns true if the index has at least one plugin entry.
func IsPluginConfigured(idx *store.IndexFile) bool {
	return len(idx.Plugins) > 0
}

// PluginAvailable checks whether the first configured plugin's command
// is executable. Returns false if no plugin is configured or the command
// is not found.
func PluginAvailable(idx *store.IndexFile) bool {
	if len(idx.Plugins) == 0 {
		return false
	}
	if _, err := exec.LookPath(idx.Plugins[0].Command); err == nil {
		return true
	}
	return false
}