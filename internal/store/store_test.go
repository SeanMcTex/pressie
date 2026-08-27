package store

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestSaveLoadIndex_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	idx := &IndexFile{
		Version: 1,
		Plugins: []PluginConfig{
			{Name: "manual", Command: "plugins/manual.sh", TimeoutMs: 1000},
		},
		Contacts: map[string]ContactMapping{
			"manual:kris": {
				File:       "_private/kris.json",
				Visibility: "private",
				Tags:       []string{"art", "irish"},
			},
		},
	}

	if err := SaveIndex(dir, idx); err != nil {
		t.Fatalf("SaveIndex: %v", err)
	}

	loaded, err := LoadIndex(dir)
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if loaded.Version != 1 {
		t.Errorf("Version = %d, want 1", loaded.Version)
	}
	if len(loaded.Plugins) != 1 {
		t.Fatalf("Plugins len = %d, want 1", len(loaded.Plugins))
	}
	if loaded.Plugins[0].Name != "manual" {
		t.Errorf("Plugin name = %q, want %q", loaded.Plugins[0].Name, "manual")
	}
	m, ok := loaded.Contacts["manual:kris"]
	if !ok {
		t.Fatal("contact manual:kris not found")
	}
	if m.File != "_private/kris.json" {
		t.Errorf("File = %q, want %q", m.File, "_private/kris.json")
	}
	if m.Visibility != "private" {
		t.Errorf("Visibility = %q, want %q", m.Visibility, "private")
	}
}

func TestLoadIndex_NotExist(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadIndex(dir)
	if err == nil {
		t.Fatal("expected error for missing index, got nil")
	}
}

func TestSaveLoadContactFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kris.json")

	price := 85.0
	cf := &ContactFile{
		ContactKey: "manual:kris",
		Name:       "Kris McMains",
		Tags:       []string{"art", "irish"},
		GiftsGiven: []Gift{
			{
				ID:       "test-id",
				Date:     "2025-12-25",
				Occasion: "christmas",
				Item:     "Custom map",
				Price:    &price,
				Currency: "USD",
				Notes:    "Framed",
				Source:   "manual",
				Added:    "2025-12-25T10:30:00Z",
			},
		},
		GiftsReceived: []Gift{},
		Ideas: []Idea{
			{
				ID:     "idea-1",
				Item:   "Letterpress print",
				Tags:   []string{"art"},
				Status: "open",
				Added:  "2026-08-27",
			},
		},
	}

	if err := SaveContactFile(path, cf); err != nil {
		t.Fatalf("SaveContactFile: %v", err)
	}

	loaded, err := LoadContactFile(path)
	if err != nil {
		t.Fatalf("LoadContactFile: %v", err)
	}
	if loaded.ContactKey != "manual:kris" {
		t.Errorf("ContactKey = %q", loaded.ContactKey)
	}
	if loaded.Name != "Kris McMains" {
		t.Errorf("Name = %q", loaded.Name)
	}
	if len(loaded.GiftsGiven) != 1 {
		t.Fatalf("GiftsGiven len = %d", len(loaded.GiftsGiven))
	}
	if loaded.GiftsGiven[0].Item != "Custom map" {
		t.Errorf("Gift item = %q", loaded.GiftsGiven[0].Item)
	}
	if loaded.GiftsGiven[0].Price == nil || *loaded.GiftsGiven[0].Price != 85 {
		t.Errorf("Price = %v, want 85", loaded.GiftsGiven[0].Price)
	}
	if len(loaded.Ideas) != 1 {
		t.Fatalf("Ideas len = %d", len(loaded.Ideas))
	}
	if loaded.Ideas[0].Status != "open" {
		t.Errorf("Idea status = %q", loaded.Ideas[0].Status)
	}
}

func TestSaveLoadGeneralIdeas_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	ideas := []Idea{
		{ID: "g1", Item: "Ceramic pour-over", Tags: []string{"kitchen"}, Status: "open", Added: "2026-08-27"},
		{ID: "g2", Item: "Wool scarf", Tags: []string{"warm"}, Status: "purchased", Added: "2026-08-20"},
	}

	if err := SaveGeneralIdeas(dir, ideas); err != nil {
		t.Fatalf("SaveGeneralIdeas: %v", err)
	}

	loaded, err := LoadGeneralIdeas(dir)
	if err != nil {
		t.Fatalf("LoadGeneralIdeas: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("len = %d, want 2", len(loaded))
	}
	if loaded[0].Item != "Ceramic pour-over" {
		t.Errorf("Item[0] = %q", loaded[0].Item)
	}
	if loaded[1].Status != "purchased" {
		t.Errorf("Status[1] = %q", loaded[1].Status)
	}
}

func TestLoadGeneralIdeas_NotExist(t *testing.T) {
	dir := t.TempDir()
	ideas, err := LoadGeneralIdeas(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ideas) != 0 {
		t.Errorf("len = %d, want 0", len(ideas))
	}
}

func TestNewUUID_Format(t *testing.T) {
	uuidRe := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for range 100 {
		u := NewUUID()
		if !uuidRe.MatchString(u) {
			t.Errorf("UUID %q does not match v4 format", u)
		}
	}
}

func TestNewUUID_Unique(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for range 1000 {
		u := NewUUID()
		if seen[u] {
			t.Fatalf("duplicate UUID generated: %s", u)
		}
		seen[u] = true
	}
}

func TestSaveIndex_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	idx := &IndexFile{Version: 1, Plugins: []PluginConfig{}}
	if err := SaveIndex(dir, idx); err != nil {
		t.Fatalf("SaveIndex: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "_index.json")); err != nil {
		t.Fatalf("_index.json not created: %v", err)
	}
}