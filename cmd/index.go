package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rune-context/rune/internal/config"
	"github.com/rune-context/rune/internal/scanner"
	"github.com/rune-context/rune/internal/session"
	"github.com/rune-context/rune/internal/summary"
)

// Index performs a full repository scan and generates RUNE.md.
func Index(root string) error {
	cfg, err := config.Load(root)
	if err != nil {
		return fmt.Errorf("loading config (run 'rune init' first): %w", err)
	}

	fmt.Println("⟳ Scanning repository...")
	start := time.Now()

	s := scanner.New(root, cfg.Ignore)
	result, err := s.Scan()
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	fmt.Printf("  Found %d files in %s\n", len(result.Files), result.Duration.Round(time.Millisecond))

	// Save graph to .rune/
	graphPath := config.SubPath(root, config.GraphFile)
	if err := result.Graph.Save(graphPath); err != nil {
		return fmt.Errorf("saving graph: %w", err)
	}

	// Write individual file summaries to .rune/files/
	if err := summary.WriteFileSummaries(root, result.Files, result.Graph); err != nil {
		return fmt.Errorf("writing summaries: %w", err)
	}

	// Write architecture to .rune/
	if err := summary.WriteArchitecture(root, result.Files); err != nil {
		return fmt.Errorf("writing architecture: %w", err)
	}

	// Generate RUNE.md at repo root (the primary output)
	if err := summary.WriteRuneMD(root, result.Files, result.Graph); err != nil {
		return fmt.Errorf("writing RUNE.md: %w", err)
	}
	fmt.Printf("  ✓ RUNE.md generated\n")

	// Save cache for incremental updates
	cache := s.GetCache()
	cacheData, _ := json.MarshalIndent(cache, "", "  ")
	cachePath := config.SubPath(root, config.CacheDir+"/hashes.json")
	os.WriteFile(cachePath, cacheData, 0644)

	// Create session
	sess := session.New()
	for _, f := range result.Files {
		sess.AddFiles(f.RelPath)
	}
	sess.Save(root)

	runemdPath := filepath.Join(root, "RUNE.md")
	fmt.Printf("\n✓ Done in %s → %s\n", time.Since(start).Round(time.Millisecond), runemdPath)
	return nil
}
