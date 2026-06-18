package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rune-context/rune/internal/config"
	"github.com/rune-context/rune/internal/graph"
	"github.com/rune-context/rune/internal/scanner"
	"github.com/rune-context/rune/internal/session"
	"github.com/rune-context/rune/internal/summary"
)

// Update incrementally refreshes changed files and regenerates RUNE.md.
func Update(root string) error {
	cfg, err := config.Load(root)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Println("⟳ Checking for changes...")
	start := time.Now()

	// Load previous cache
	s := scanner.New(root, cfg.Ignore)
	cachePath := config.SubPath(root, config.CacheDir+"/hashes.json")
	if data, err := os.ReadFile(cachePath); err == nil {
		var cache map[string]string
		if json.Unmarshal(data, &cache) == nil {
			s.SetCache(cache)
		}
	}

	result, err := s.ScanChanged()
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	if len(result.Files) == 0 {
		fmt.Println("✓ No changes detected.")
		return nil
	}

	fmt.Printf("  %d files changed\n", len(result.Files))

	// Update graph entries (merge with existing)
	graphPath := config.SubPath(root, config.GraphFile)
	existingGraph, _ := graph.Load(graphPath)
	for _, f := range result.Files {
		if len(f.Dependencies) > 0 {
			existingGraph.Set(f.RelPath, f.Dependencies)
		}
	}
	if err := existingGraph.Save(graphPath); err != nil {
		return fmt.Errorf("saving graph: %w", err)
	}

	// Update file summaries
	if err := summary.WriteFileSummaries(root, result.Files, existingGraph); err != nil {
		return fmt.Errorf("writing summaries: %w", err)
	}

	// Re-scan all files for a full RUNE.md regeneration
	allResult, err := scanner.New(root, cfg.Ignore).Scan()
	if err == nil {
		if err := summary.WriteRuneMD(root, allResult.Files, existingGraph); err != nil {
			return fmt.Errorf("writing RUNE.md: %w", err)
		}
		fmt.Printf("  ✓ RUNE.md regenerated\n")
	}

	// Save updated cache
	cache := s.GetCache()
	cacheData, _ := json.MarshalIndent(cache, "", "  ")
	os.WriteFile(cachePath, cacheData, 0644)

	// Update session
	sess := session.New()
	for _, f := range result.Files {
		sess.AddFiles(f.RelPath)
	}
	sess.Save(root)

	fmt.Printf("✓ Update complete in %s\n", time.Since(start).Round(time.Millisecond))
	return nil
}
