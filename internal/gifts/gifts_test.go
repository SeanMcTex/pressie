package gifts

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/SeanMcTex/pressie/internal/store"
)

// setupGiftsDir creates a temp directory with _index.json and optional
// pre-existing contact files. Returns the dir path and the index.
func setupGiftsDir(t *testing.T, contacts map[string]store.ContactMapping) (string, *store.IndexFile) {
	t.Helper()
	dir := t.TempDir()
	idx := &store.IndexFile{
		Version:  1,
		Plugins:  []store.PluginConfig{},
		Contacts: contacts,
	}
	if idx.Contacts == nil {
		idx.Contacts = make(map[string]store.ContactMapping)
	}
	if err := store.SaveIndex(dir, idx); err != nil {
		t.Fatalf("SaveIndex: %v", err)
	}
	// Create _private and _shared dirs.
	os.MkdirAll(filepath.Join(dir, "_private"), 0o755)
	os.MkdirAll(filepath.Join(dir, "_shared"), 0o755)
	return dir, idx
}

// writeContactFile writes a contact file at the given relative path.
func writeContactFile(t *testing.T, giftsDir, relPath, key, name string, tags []string) {
	t.Helper()
	cf := &store.ContactFile{
		ContactKey:    key,
		Name:          name,
		Tags:          tags,
		GiftsGiven:    []store.Gift{},
		GiftsReceived: []store.Gift{},
		Ideas:         []store.Idea{},
	}
	absPath := filepath.Join(giftsDir, relPath)
	os.MkdirAll(filepath.Dir(absPath), 0o755)
	if err := store.SaveContactFile(absPath, cf); err != nil {
		t.Fatalf("SaveContactFile: %v", err)
	}
}

func TestResolveContact_ManualFallback_NewContact(t *testing.T) {
	dir, idx := setupGiftsDir(t, nil)
	key, name, relPath, err := ResolveContact(context.Background(), dir, idx, "Kris", "private")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "manual:kris" {
		t.Errorf("key = %q, want %q", key, "manual:kris")
	}
	if name != "Kris" {
		t.Errorf("name = %q, want %q", name, "Kris")
	}
	if relPath != "_private/kris.json" {
		t.Errorf("relPath = %q, want %q", relPath, "_private/kris.json")
	}
}

func TestResolveContact_ManualFallback_Shared(t *testing.T) {
	dir, idx := setupGiftsDir(t, nil)
	_, _, relPath, err := ResolveContact(context.Background(), dir, idx, "Blair", "shared")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if relPath != "_shared/blair.json" {
		t.Errorf("relPath = %q, want %q", relPath, "_shared/blair.json")
	}
}

func TestResolveContact_ExactSlugMatch(t *testing.T) {
	dir, idx := setupGiftsDir(t, map[string]store.ContactMapping{
		"manual:kris-mcmains": {
			File:       "_private/kris-mcmains.json",
			Visibility: "private",
			Tags:       []string{"art"},
		},
	})
	writeContactFile(t, dir, "_private/kris-mcmains.json", "manual:kris-mcmains", "Kris McMains", []string{"art"})

	key, name, _, err := ResolveContact(context.Background(), dir, idx, "Kris McMains", "private")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "manual:kris-mcmains" {
		t.Errorf("key = %q, want %q", key, "manual:kris-mcmains")
	}
	if name != "Kris McMains" {
		t.Errorf("name = %q, want %q", name, "Kris McMains")
	}
}

func TestResolveContact_ExactSlugCaseInsensitive(t *testing.T) {
	dir, idx := setupGiftsDir(t, map[string]store.ContactMapping{
		"manual:kris-mcmains": {
			File:       "_private/kris-mcmains.json",
			Visibility: "private",
		},
	})
	writeContactFile(t, dir, "_private/kris-mcmains.json", "manual:kris-mcmains", "Kris McMains", nil)

	_, name, _, err := ResolveContact(context.Background(), dir, idx, "kris mcmains", "private")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Kris McMains" {
		t.Errorf("name = %q, want %q", name, "Kris McMains")
	}
}

func TestResolveContact_FuzzySingleMatch(t *testing.T) {
	dir, idx := setupGiftsDir(t, map[string]store.ContactMapping{
		"manual:kris-mcmains": {
			File:       "_private/kris-mcmains.json",
			Visibility: "private",
		},
	})
	writeContactFile(t, dir, "_private/kris-mcmains.json", "manual:kris-mcmains", "Kris McMains", nil)

	// "Kris" is contained in "Kris McMains" — single match.
	key, name, _, err := ResolveContact(context.Background(), dir, idx, "Kris", "private")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "manual:kris-mcmains" {
		t.Errorf("key = %q, want %q", key, "manual:kris-mcmains")
	}
	if name != "Kris McMains" {
		t.Errorf("name = %q, want %q", name, "Kris McMains")
	}
}

func TestResolveContact_FuzzyQueryContainsStored(t *testing.T) {
	dir, idx := setupGiftsDir(t, map[string]store.ContactMapping{
		"manual:al": {
			File:       "_private/al.json",
			Visibility: "private",
		},
	})
	writeContactFile(t, dir, "_private/al.json", "manual:al", "Al", nil)

	// "Al Brown" contains "Al" — single match.
	_, name, _, err := ResolveContact(context.Background(), dir, idx, "Al Brown", "private")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Al" {
		t.Errorf("name = %q, want %q", name, "Al")
	}
}

func TestResolveContact_AmbiguousFuzzy(t *testing.T) {
	dir, idx := setupGiftsDir(t, map[string]store.ContactMapping{
		"manual:kris-mcmains": {
			File:       "_private/kris-mcmains.json",
			Visibility: "private",
		},
		"manual:kris-smith": {
			File:       "_private/kris-smith.json",
			Visibility: "private",
		},
	})
	writeContactFile(t, dir, "_private/kris-mcmains.json", "manual:kris-mcmains", "Kris McMains", nil)
	writeContactFile(t, dir, "_private/kris-smith.json", "manual:kris-smith", "Kris Smith", nil)

	_, _, _, err := ResolveContact(context.Background(), dir, idx, "Kris", "private")
	if err == nil {
		t.Fatal("expected ambiguity error, got nil")
	}
}

func TestResolveContact_ManualFallbackWhenNoFuzzyMatch(t *testing.T) {
	dir, idx := setupGiftsDir(t, map[string]store.ContactMapping{
		"manual:blair": {
			File:       "_private/blair.json",
			Visibility: "private",
		},
	})
	writeContactFile(t, dir, "_private/blair.json", "manual:blair", "Blair", nil)

	// "Kris" doesn't match any existing contact — manual fallback creates new.
	key, _, _, err := ResolveContact(context.Background(), dir, idx, "Kris", "private")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "manual:kris" {
		t.Errorf("key = %q, want %q", key, "manual:kris")
	}
}

func TestEnsureContactFile_NewFile(t *testing.T) {
	dir := t.TempDir()
	relPath := "_private/kris.json"
	os.MkdirAll(filepath.Join(dir, "_private"), 0o755)

	cf, err := EnsureContactFile(dir, relPath, "manual:kris", "Kris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.ContactKey != "manual:kris" {
		t.Errorf("ContactKey = %q", cf.ContactKey)
	}
	if cf.Name != "Kris" {
		t.Errorf("Name = %q", cf.Name)
	}
	if len(cf.Ideas) != 0 || len(cf.GiftsGiven) != 0 {
		t.Errorf("new file should have empty slices")
	}
}

func TestEnsureContactFile_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	relPath := "_private/kris.json"
	os.MkdirAll(filepath.Join(dir, "_private"), 0o755)

	// Write existing file with data.
	writeContactFile(t, dir, relPath, "manual:kris", "Kris", []string{"art"})

	cf, err := EnsureContactFile(dir, relPath, "manual:kris", "Kris")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.Name != "Kris" {
		t.Errorf("Name = %q", cf.Name)
	}
	if len(cf.Tags) != 1 || cf.Tags[0] != "art" {
		t.Errorf("Tags = %v, want [art]", cf.Tags)
	}
}

func TestRegisterContact_NewContact(t *testing.T) {
	idx := &store.IndexFile{Version: 1}
	RegisterContact(idx, "manual:kris", "_private/kris.json", "private", []string{"art"})
	m, ok := idx.Contacts["manual:kris"]
	if !ok {
		t.Fatal("contact not registered")
	}
	if m.File != "_private/kris.json" {
		t.Errorf("File = %q", m.File)
	}
	if m.Visibility != "private" {
		t.Errorf("Visibility = %q", m.Visibility)
	}
}

func TestRegisterContact_DoesNotOverwrite(t *testing.T) {
	idx := &store.IndexFile{
		Version: 1,
		Contacts: map[string]store.ContactMapping{
			"manual:kris": {File: "_private/kris.json", Visibility: "private", Tags: []string{"art"}},
		},
	}
	// Try to register again with different data — should be ignored.
	RegisterContact(idx, "manual:kris", "_shared/kris.json", "shared", []string{"music"})
	m := idx.Contacts["manual:kris"]
	if m.File != "_private/kris.json" {
		t.Errorf("File = %q, want %q (should not overwrite)", m.File, "_private/kris.json")
	}
	if m.Visibility != "private" {
		t.Errorf("Visibility = %q, want %q", m.Visibility, "private")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Kris McMains", "kris-mcmains"},
		{"Kris", "kris"},
		{"  Blair Raker  ", "blair-raker"},
		{"O'Brien", "o-brien"},
		{"Mary-Jane", "mary-jane"},
		{"", ""},
	}
	for _, tc := range tests {
		got := slugify(tc.input)
		if got != tc.want {
			t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}