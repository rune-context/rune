// Package session manages .rune/sessions/ tracking.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rune-context/rune/internal/config"
)

// Session tracks recently touched files.
type Session struct {
	ID        string   `json:"id"`
	Files     []string `json:"files"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// New creates a new session.
func New() *Session {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Session{
		ID:        fmt.Sprintf("%d", time.Now().UnixMilli()),
		Files:     []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddFiles adds files to the session.
func (s *Session) AddFiles(files ...string) {
	seen := make(map[string]bool)
	for _, f := range s.Files {
		seen[f] = true
	}
	for _, f := range files {
		if !seen[f] {
			s.Files = append(s.Files, f)
			seen[f] = true
		}
	}
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
}

// Save writes the session to disk.
func (s *Session) Save(root string) error {
	dir := config.SubPath(root, config.SessionsDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadLatest loads the most recent session.
func LoadLatest(root string) (*Session, error) {
	dir := config.SubPath(root, config.SessionsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var latest string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			if e.Name() > latest {
				latest = e.Name()
			}
		}
	}

	if latest == "" {
		return nil, fmt.Errorf("no sessions found")
	}

	data, err := os.ReadFile(filepath.Join(dir, latest))
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
