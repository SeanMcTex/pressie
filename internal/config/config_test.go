package config

import (
	"testing"
)

func TestParseArgs_AllStringFlags(t *testing.T) {
	cfg, err := ParseArgs([]string{
		"--for", "Kris", "--to", "Blair", "--from", "Sam",
		"--item", "Letterpress print", "--occasion", "christmas",
		"--date", "2025-12-25", "--url", "https://example.com",
		"--notes", "Saw it at the market",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.For != "Kris" {
		t.Errorf("For = %q, want %q", cfg.For, "Kris")
	}
	if cfg.To != "Blair" {
		t.Errorf("To = %q, want %q", cfg.To, "Blair")
	}
	if cfg.From != "Sam" {
		t.Errorf("From = %q, want %q", cfg.From, "Sam")
	}
	if cfg.Item != "Letterpress print" {
		t.Errorf("Item = %q, want %q", cfg.Item, "Letterpress print")
	}
	if cfg.Occasion != "christmas" {
		t.Errorf("Occasion = %q, want %q", cfg.Occasion, "christmas")
	}
	if cfg.Date != "2025-12-25" {
		t.Errorf("Date = %q, want %q", cfg.Date, "2025-12-25")
	}
	if cfg.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", cfg.URL, "https://example.com")
	}
	if cfg.Notes != "Saw it at the market" {
		t.Errorf("Notes = %q, want %q", cfg.Notes, "Saw it at the market")
	}
}

func TestParseTags_CommaSeparated(t *testing.T) {
	cfg, err := ParseArgs([]string{"--tags", "art, irish, kitchen"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tags) != 3 {
		t.Fatalf("Tags len = %d, want 3", len(cfg.Tags))
	}
	want := []string{"art", "irish", "kitchen"}
	for i, w := range want {
		if cfg.Tags[i] != w {
			t.Errorf("Tags[%d] = %q, want %q", i, cfg.Tags[i], w)
		}
	}
}

func TestParseTags_EmptyString(t *testing.T) {
	cfg, err := ParseArgs([]string{"--tags", ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Tags != nil {
		t.Errorf("Tags = %v, want nil", cfg.Tags)
	}
}

func TestParseTags_TrimsWhitespace(t *testing.T) {
	cfg, _ := ParseArgs([]string{"--tags", "  art ,  irish  "})
	if len(cfg.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(cfg.Tags))
	}
	if cfg.Tags[0] != "art" {
		t.Errorf("Tags[0] = %q, want %q", cfg.Tags[0], "art")
	}
	if cfg.Tags[1] != "irish" {
		t.Errorf("Tags[1] = %q, want %q", cfg.Tags[1], "irish")
	}
}

func TestParseArgs_Price(t *testing.T) {
	cfg, err := ParseArgs([]string{"--price", "42.50"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Price != 42.50 {
		t.Errorf("Price = %v, want 42.50", cfg.Price)
	}
	if !cfg.HasPrice {
		t.Error("HasPrice = false, want true")
	}
}

func TestParseArgs_InvalidPrice(t *testing.T) {
	_, err := ParseArgs([]string{"--price", "notanumber"})
	if err == nil {
		t.Fatal("expected error for invalid price, got nil")
	}
}

func TestParseArgs_EqualsForm(t *testing.T) {
	cfg, err := ParseArgs([]string{"--for=Blair", "--item=pour-over"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.For != "Blair" {
		t.Errorf("For = %q, want %q", cfg.For, "Blair")
	}
	if cfg.Item != "pour-over" {
		t.Errorf("Item = %q, want %q", cfg.Item, "pour-over")
	}
}

func TestParseArgs_PrivateDefault(t *testing.T) {
	cfg, err := ParseArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Visibility != "private" {
		t.Errorf("Visibility = %q, want %q", cfg.Visibility, "private")
	}
}

func TestParseArgs_SharedFlag(t *testing.T) {
	cfg, err := ParseArgs([]string{"--shared"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Visibility != "shared" {
		t.Errorf("Visibility = %q, want %q", cfg.Visibility, "shared")
	}
}

func TestParseArgs_PrivateFlag(t *testing.T) {
	cfg, err := ParseArgs([]string{"--private"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Visibility != "private" {
		t.Errorf("Visibility = %q, want %q", cfg.Visibility, "private")
	}
}

func TestParseArgs_StatusDirectionYear(t *testing.T) {
	cfg, err := ParseArgs([]string{"--status", "purchased", "--direction", "given", "--year", "2025"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Status != "purchased" {
		t.Errorf("Status = %q, want %q", cfg.Status, "purchased")
	}
	if cfg.Direction != "given" {
		t.Errorf("Direction = %q, want %q", cfg.Direction, "given")
	}
	if cfg.Year != "2025" {
		t.Errorf("Year = %q, want %q", cfg.Year, "2025")
	}
}

func TestParseArgs_PositionalGiftsDir(t *testing.T) {
	cfg, err := ParseArgs([]string{"/tmp/my-gifts", "--for", "Kris"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GiftsDir != "/tmp/my-gifts" {
		t.Errorf("GiftsDir = %q, want %q", cfg.GiftsDir, "/tmp/my-gifts")
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	_, err := ParseArgs([]string{"--bogus", "value"})
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestParseArgs_MissingValue(t *testing.T) {
	_, err := ParseArgs([]string{"--for"})
	if err == nil {
		t.Fatal("expected error for missing value, got nil")
	}
}

func TestParseArgs_MultiplePositionals(t *testing.T) {
	_, err := ParseArgs([]string{"/tmp/a", "/tmp/b"})
	if err == nil {
		t.Fatal("expected error for multiple positionals, got nil")
	}
}

func TestParseArgs_IdeaFlag(t *testing.T) {
	cfg, err := ParseArgs([]string{"--idea", "abc-123-def"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Idea != "abc-123-def" {
		t.Errorf("Idea = %q, want %q", cfg.Idea, "abc-123-def")
	}
}