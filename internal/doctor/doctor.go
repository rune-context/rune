// Package doctor validates .rune health.
package doctor

import (
	"fmt"
	"os"
	"strings"

	"github.com/rune-context/rune/internal/config"
	"github.com/rune-context/rune/internal/graph"
)

// Issue represents a health check finding.
type Issue struct {
	Level   string `json:"level"` // "error", "warning", "info"
	Message string `json:"message"`
}

// Check runs all health checks on a repository.
func Check(root string) []Issue {
	var issues []Issue

	// Check .rune directory exists
	runePath := config.RunePath(root)
	if _, err := os.Stat(runePath); os.IsNotExist(err) {
		issues = append(issues, Issue{"error", ".rune directory not found. Run 'rune init' first."})
		return issues
	}

	// Check graph
	graphPath := config.SubPath(root, config.GraphFile)
	if _, err := os.Stat(graphPath); os.IsNotExist(err) {
		issues = append(issues, Issue{"warning", "graph.json not found. Run 'rune index' to generate."})
	} else {
		g, err := graph.Load(graphPath)
		if err != nil {
			issues = append(issues, Issue{"error", fmt.Sprintf("graph.json is corrupted: %v", err)})
		} else {
			if len(g.Files()) == 0 {
				issues = append(issues, Issue{"warning", "graph.json is empty."})
			} else {
				issues = append(issues, Issue{"info", fmt.Sprintf("graph.json: %d files tracked.", len(g.Files()))})
			}
		}
	}

	// Check files directory
	filesDir := config.SubPath(root, config.FilesDir)
	if entries, err := os.ReadDir(filesDir); err != nil {
		issues = append(issues, Issue{"warning", "files/ directory missing. Run 'rune index'."})
	} else {
		count := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".md") {
				count++
			}
		}
		issues = append(issues, Issue{"info", fmt.Sprintf("%d file summaries found.", count)})
	}

	// Check architecture
	archPath := config.SubPath(root, config.ArchitectureFile)
	if _, err := os.Stat(archPath); os.IsNotExist(err) {
		issues = append(issues, Issue{"warning", "architecture.md not found."})
	} else {
		issues = append(issues, Issue{"info", "architecture.md exists."})
	}

	// Check config
	cfgPath := config.SubPath(root, "config.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		issues = append(issues, Issue{"warning", "config.json missing."})
	}

	return issues
}
