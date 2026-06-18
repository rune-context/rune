// Package context provides query-based context retrieval from .rune data.
package context

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rune-context/rune/internal/config"
	"github.com/rune-context/rune/internal/graph"
)

// Result holds the context produced for a query.
type Result struct {
	Query        string   `json:"query"`
	Files        []string `json:"files"`
	Summaries    []string `json:"summaries"`
	Architecture string   `json:"architecture,omitempty"`
	Conventions  string   `json:"conventions,omitempty"`
}

// Query finds relevant files and summaries for a given query string.
func Query(root, query string) (*Result, error) {
	g, err := graph.Load(config.SubPath(root, config.GraphFile))
	if err != nil {
		return nil, fmt.Errorf("loading graph: %w", err)
	}

	keywords := tokenize(query)
	scored := scoreFiles(g, keywords, root)

	// Take top results
	maxResults := 20
	if len(scored) < maxResults {
		maxResults = len(scored)
	}
	topFiles := scored[:maxResults]

	// Load summaries
	var summaries []string
	for _, sf := range topFiles {
		sum := loadFileSummary(root, sf.file)
		if sum != "" {
			summaries = append(summaries, sum)
		}
	}

	// Load architecture
	arch := ""
	archPath := config.SubPath(root, config.ArchitectureFile)
	if data, err := os.ReadFile(archPath); err == nil {
		arch = string(data)
	}

	// Load conventions
	conv := ""
	convPath := config.SubPath(root, config.ConventionsFile)
	if data, err := os.ReadFile(convPath); err == nil {
		conv = string(data)
	}

	files := make([]string, len(topFiles))
	for i, sf := range topFiles {
		files[i] = sf.file
	}

	return &Result{
		Query:        query,
		Files:        files,
		Summaries:    summaries,
		Architecture: arch,
		Conventions:  conv,
	}, nil
}

type scoredFile struct {
	file  string
	score int
}

func scoreFiles(g *graph.Graph, keywords []string, root string) []scoredFile {
	allFiles := g.Files()
	var scored []scoredFile

	for _, f := range allFiles {
		score := 0
		fLower := strings.ToLower(f)
		for _, kw := range keywords {
			if strings.Contains(fLower, kw) {
				score += 10
			}
		}

		// Check summary content
		sum := loadFileSummary(root, f)
		sumLower := strings.ToLower(sum)
		for _, kw := range keywords {
			if strings.Contains(sumLower, kw) {
				score += 5
			}
		}

		// Bonus for files with more connections
		deps := g.Get(f)
		dependents := g.Dependents(f)
		score += len(deps) + len(dependents)

		if score > 0 {
			scored = append(scored, scoredFile{file: f, score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	return scored
}

func tokenize(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	var tokens []string
	stopwords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true,
		"or": true, "is": true, "in": true, "to": true,
		"for": true, "of": true, "with": true, "add": true,
		"implement": true, "create": true, "make": true,
	}
	for _, w := range words {
		w = strings.Trim(w, ".,!?\"'")
		if len(w) > 1 && !stopwords[w] {
			tokens = append(tokens, w)
		}
	}
	return tokens
}

func loadFileSummary(root, file string) string {
	safeName := strings.ReplaceAll(file, "/", "__")
	safeName = strings.TrimSuffix(safeName, filepath.Ext(safeName)) + ".md"
	path := filepath.Join(config.SubPath(root, config.FilesDir), safeName)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// Format produces a formatted string for the context result.
func (r *Result) Format() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Context for: %s\n\n", r.Query))

	if r.Architecture != "" {
		sb.WriteString("---\n\n")
		sb.WriteString(r.Architecture)
		sb.WriteString("\n")
	}

	if r.Conventions != "" {
		sb.WriteString("---\n\n")
		sb.WriteString(r.Conventions)
		sb.WriteString("\n")
	}

	if len(r.Files) > 0 {
		sb.WriteString("---\n\n## Relevant Files\n\n")
		for _, f := range r.Files {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	for _, sum := range r.Summaries {
		sb.WriteString("---\n\n")
		sb.WriteString(sum)
		sb.WriteString("\n")
	}

	return sb.String()
}
