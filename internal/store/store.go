package store

// Package store handles reading and writing JSON files in the gifts directory.
// This is a stub — implementation to follow.

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Gift represents a single gift given or received.
type Gift struct {
	ID       string   `json:"id"`
	Date     string   `json:"date"`
	Occasion string   `json:"occasion"`
	Item     string   `json:"item"`
	Price    *float64 `json:"price,omitempty"`
	Currency string   `json:"currency,omitempty"`
	Notes    string   `json:"notes,omitempty"`
	Images   []string `json:"images,omitempty"`
	Source   string   `json:"source,omitempty"`
	Added    string   `json:"added,omitempty"`
}

// Idea represents a gift idea (assigned to a person or general).
type Idea struct {
	ID            string   `json:"id"`
	Item          string   `json:"item"`
	URL           string   `json:"url,omitempty"`
	PriceEstimate *float64 `json:"price_estimate,omitempty"`
	Currency      string   `json:"currency,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Status        string   `json:"status"`
	Added         string   `json:"added"`
	Notes         string   `json:"notes,omitempty"`
	Images        []string `json:"images,omitempty"`
	AssignedTo    *string  `json:"assigned_to,omitempty"`
}

// ContactFile is the per-contact JSON structure.
type ContactFile struct {
	ContactKey    string `json:"contact_key"`
	Name          string `json:"name"`
	Tags          []string `json:"tags,omitempty"`
	GiftsGiven    []Gift `json:"gifts_given"`
	GiftsReceived []Gift `json:"gifts_received"`
	Ideas         []Idea `json:"ideas"`
	Preferences   string `json:"preferences,omitempty"`
}

// IndexFile is the _index.json structure.
type IndexFile struct {
	Version  int `json:"version"`
	Plugins  []PluginConfig `json:"plugins"`
	Sync     *SyncConfig `json:"sync,omitempty"`
	Contacts map[string]ContactMapping `json:"contacts,omitempty"`
}

type PluginConfig struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	TimeoutMs int    `json:"timeout_ms,omitempty"`
}

type SyncConfig struct {
	Type   string `json:"type"`
	Remote string `json:"remote,omitempty"`
	Branch string `json:"branch,omitempty"`
}

type ContactMapping struct {
	File       string   `json:"file"`
	Visibility string   `json:"visibility"`
	Tags       []string `json:"tags,omitempty"`
}

// LoadIndex reads _index.json from the gifts directory.
func LoadIndex(giftsDir string) (*IndexFile, error) {
	path := filepath.Join(giftsDir, "_index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx IndexFile
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// SaveIndex writes _index.json to the gifts directory.
func SaveIndex(giftsDir string, idx *IndexFile) error {
	path := filepath.Join(giftsDir, "_index.json")
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, data)
}

// LoadContactFile reads a per-contact JSON file.
func LoadContactFile(path string) (*ContactFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cf ContactFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}
	return &cf, nil
}

// SaveContactFile writes a per-contact JSON file.
func SaveContactFile(path string, cf *ContactFile) error {
	data, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, data)
}

// LoadGeneralIdeas reads the general ideas file.
func LoadGeneralIdeas(giftsDir string) ([]Idea, error) {
	path := filepath.Join(giftsDir, "_ideas-general.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Idea{}, nil
		}
		return nil, err
	}
	var ideas []Idea
	if err := json.Unmarshal(data, &ideas); err != nil {
		return nil, err
	}
	return ideas, nil
}

// SaveGeneralIdeas writes the general ideas file.
func SaveGeneralIdeas(giftsDir string, ideas []Idea) error {
	path := filepath.Join(giftsDir, "_ideas-general.json")
	data, err := json.MarshalIndent(ideas, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, data)
}

// writeAtomic writes data to path atomically by first writing to a temp
// file in the same directory, then renaming. This prevents partial writes
// from crashes, interrupted processes, or concurrent access.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".pressie-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Clean up temp file on any error path.
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}