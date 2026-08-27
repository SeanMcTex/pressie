package contacts

// Package contacts implements the plugin interface for resolving names to contacts.
// Plugins are external executables that take a query string and return JSON.

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Contact represents a resolved contact from a plugin.
type Contact struct {
	Key      string            `json:"key"`
	Name     string            `json:"name"`
	Email    string            `json:"email,omitempty"`
	Birthday string            `json:"birthday,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Plugin runs a contacts plugin executable and returns matches.
type Plugin struct {
	Name      string
	Command   string
	TimeoutMs int
}

// Resolve queries the plugin for contacts matching the query string.
func (p *Plugin) Resolve(ctx context.Context, query string) ([]Contact, error) {
	timeout := time.Duration(p.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.Command, query)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("plugin %s failed: %w", p.Name, err)
	}

	var contacts []Contact
	if err := json.Unmarshal(output, &contacts); err != nil {
		return nil, fmt.Errorf("plugin %s returned invalid JSON: %w", p.Name, err)
	}

	// Namespace keys with plugin name
	for i := range contacts {
		contacts[i].Key = fmt.Sprintf("%s:%s", p.Name, contacts[i].Key)
	}

	return contacts, nil
}