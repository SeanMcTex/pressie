package config

// Package config handles argument parsing and configuration.
// Stub — implementation to follow.

import (
	"os"
	"path/filepath"
)

// Config holds parsed CLI arguments and resolved paths.
type Config struct {
	GiftsDir    string
	For         string
	To          string
	From        string
	Item        string
	Occasion    string
	Date        string
	URL         string
	Tags        []string
	Price       float64
	HasPrice    bool
	Notes       string
	Visibility  string // "private" or "shared"
	Status      string
	Direction   string
	Year        string
}

// DefaultGiftsDir returns the default gifts directory path.
func DefaultGiftsDir() string {
	if env := os.Getenv("PRESSIE_DIR"); env != "" {
		return env
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pressie")
}

// ParseArgs parses a flag slice into a Config.
// TODO: implement proper flag parsing
func ParseArgs(args []string) (*Config, error) {
	return &Config{
		Visibility: "private",
	}, nil
}