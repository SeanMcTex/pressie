package config

// Package config handles argument parsing and configuration.
// Stub — implementation to follow.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// ParseArgs parses a flag slice (with the command name already stripped) into a Config.
// All flags use --flag value or --flag=value form. A bare non-flag argument is treated
// as the gifts directory (positional, used by init).
func ParseArgs(args []string) (*Config, error) {
	cfg := &Config{Visibility: "private"}

	i := 0
	for i < len(args) {
		arg := args[i]
		i++

		if !strings.HasPrefix(arg, "--") {
			// Positional argument: treats it as the gifts directory.
			if cfg.GiftsDir == "" {
				cfg.GiftsDir = arg
			} else {
				return nil, fmt.Errorf("unexpected argument: %s", arg)
			}
			continue
		}

		// Split --flag=value form.
		name := arg
		value := ""
		hasEqValue := false
		if eq := strings.IndexByte(arg, '='); eq != -1 {
			name = arg[:eq]
			value = arg[eq+1:]
			hasEqValue = true
		}

		// Boolean flags take no value.
		switch name {
		case "--private":
			cfg.Visibility = "private"
			continue
		case "--shared":
			cfg.Visibility = "shared"
			continue
		}

		// Value flags: consume next arg unless --flag=value form.
		if !hasEqValue {
			if i >= len(args) {
				return nil, fmt.Errorf("flag %s requires a value", name)
			}
			value = args[i]
			i++
		}

		switch name {
		case "--for":
			cfg.For = value
		case "--to":
			cfg.To = value
		case "--from":
			cfg.From = value
		case "--item":
			cfg.Item = value
		case "--occasion":
			cfg.Occasion = value
		case "--date":
			cfg.Date = value
		case "--url":
			cfg.URL = value
		case "--notes":
			cfg.Notes = value
		case "--tags":
			cfg.Tags = parseTags(value)
		case "--price":
			p, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid --price value %q: %w", value, err)
			}
			cfg.Price = p
			cfg.HasPrice = true
		case "--status":
			cfg.Status = value
		case "--direction":
			cfg.Direction = value
		case "--year":
			cfg.Year = value
		default:
			return nil, fmt.Errorf("unknown flag: %s", name)
		}
	}

	if cfg.GiftsDir == "" {
		cfg.GiftsDir = DefaultGiftsDir()
	}

	return cfg, nil
}

// parseTags splits a comma-separated tag string into a trimmed slice.
// An empty input yields nil.
func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}