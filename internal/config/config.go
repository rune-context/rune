// Package config defines the Rune configuration structure and defaults.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	// RuneDir is the directory name used by Rune inside repositories.
	RuneDir = ".rune"

	// GraphFile is the dependency graph filename.
	GraphFile = "graph.json"

	// ArchitectureFile is the architecture summary filename.
	ArchitectureFile = "architecture.md"

	// ConventionsFile is the conventions summary filename.
	ConventionsFile = "conventions.md"

	// SpecFile is the project spec filename.
	SpecFile = "spec.md"

	// FilesDir holds per-file summaries.
	FilesDir = "files"

	// FeaturesDir holds per-feature summaries.
	FeaturesDir = "features"

	// OwnershipDir holds ownership information.
	OwnershipDir = "ownership"

	// SessionsDir holds session data.
	SessionsDir = "sessions"

	// CacheDir holds internal cache data.
	CacheDir = "cache"
)

// Config holds Rune configuration for a repository.
type Config struct {
	Version string   `json:"version"`
	Root    string   `json:"root"`
	Ignore  []string `json:"ignore"`
}

// DefaultIgnore returns the default set of ignored patterns.
func DefaultIgnore() []string {
	return []string{
		".git",
		".rune",
		"node_modules",
		"vendor",
		"__pycache__",
		".venv",
		"venv",
		"dist",
		"build",
		".next",
		".nuxt",
		"target",
		"bin",
		"obj",
		".idea",
		".vscode",
		"*.pyc",
		"*.o",
		"*.so",
		"*.dylib",
		"*.exe",
		"*.dll",
		"*.class",
		"*.jar",
		"*.wasm",
		"*.min.js",
		"*.min.css",
		"*.map",
		"*.lock",
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"go.sum",
		"Cargo.lock",
	}
}

// RunePath returns the full path to the .rune directory for a given repo root.
func RunePath(root string) string {
	return filepath.Join(root, RuneDir)
}

// SubPath returns the path to a file/dir inside .rune.
func SubPath(root, name string) string {
	return filepath.Join(root, RuneDir, name)
}

// Load reads the Rune config from a repository root.
func Load(root string) (*Config, error) {
	path := filepath.Join(RunePath(root), "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the Rune config to a repository root.
func Save(root string, cfg *Config) error {
	path := filepath.Join(RunePath(root), "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
