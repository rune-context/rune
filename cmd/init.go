// Package cmd implements the CLI commands for Rune.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rune-context/rune/internal/config"
)

// Init initializes a .rune directory in the current repository.
func Init(root string) error {
	runePath := config.RunePath(root)

	// Create .rune directory and subdirectories (idempotent)
	dirs := []string{
		runePath,
		config.SubPath(root, config.FilesDir),
		config.SubPath(root, config.FeaturesDir),
		config.SubPath(root, config.OwnershipDir),
		config.SubPath(root, config.SessionsDir),
		config.SubPath(root, config.CacheDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}

	// Create config.json if it doesn't exist
	cfgPath := config.SubPath(root, "config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfg := &config.Config{
			Version: "0.1",
			Root:    root,
			Ignore:  config.DefaultIgnore(),
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(cfgPath, data, 0644); err != nil {
			return err
		}
	}

	// Create spec.md if it doesn't exist
	specPath := config.SubPath(root, config.SpecFile)
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		spec := "# Project Spec\n\nDescribe your project here.\n"
		if err := os.WriteFile(specPath, []byte(spec), 0644); err != nil {
			return err
		}
	}

	// Create conventions.md if it doesn't exist
	convPath := config.SubPath(root, config.ConventionsFile)
	if _, err := os.Stat(convPath); os.IsNotExist(err) {
		conv := "# Conventions\n\nAdd your coding conventions here.\n"
		if err := os.WriteFile(convPath, []byte(conv), 0644); err != nil {
			return err
		}
	}

	fmt.Println("✓ Rune initialized in", runePath)
	return nil
}
