// Package graph manages the dependency graph stored in .rune/graph.json.
package graph

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

// Graph represents the dependency graph: file -> [dependencies].
type Graph struct {
	mu   sync.RWMutex
	Deps map[string][]string `json:"deps"`
}

// New creates a new empty graph.
func New() *Graph {
	return &Graph{
		Deps: make(map[string][]string),
	}
}

// Load reads a graph from a JSON file.
func Load(path string) (*Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, err
	}
	g := New()
	if err := json.Unmarshal(data, &g.Deps); err != nil {
		return nil, err
	}
	return g, nil
}

// Save writes the graph to a JSON file.
func (g *Graph) Save(path string) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Sort keys for deterministic output
	keys := make([]string, 0, len(g.Deps))
	for k := range g.Deps {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make(map[string][]string, len(g.Deps))
	for _, k := range keys {
		deps := make([]string, len(g.Deps[k]))
		copy(deps, g.Deps[k])
		sort.Strings(deps)
		sorted[k] = deps
	}

	data, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Set sets the dependencies for a file.
func (g *Graph) Set(file string, deps []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Deps[file] = deps
}

// Get returns the dependencies for a file.
func (g *Graph) Get(file string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Deps[file]
}

// Remove removes a file from the graph.
func (g *Graph) Remove(file string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Deps, file)
}

// Dependents returns all files that depend on the given file.
func (g *Graph) Dependents(file string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []string
	for k, deps := range g.Deps {
		for _, d := range deps {
			if d == file {
				result = append(result, k)
				break
			}
		}
	}
	sort.Strings(result)
	return result
}

// Files returns all files in the graph, sorted.
func (g *Graph) Files() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	keys := make([]string, 0, len(g.Deps))
	for k := range g.Deps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Related returns files related to the given file (deps + dependents), sorted.
func (g *Graph) Related(file string) []string {
	seen := make(map[string]bool)
	for _, d := range g.Get(file) {
		seen[d] = true
	}
	for _, d := range g.Dependents(file) {
		seen[d] = true
	}
	result := make([]string, 0, len(seen))
	for f := range seen {
		result = append(result, f)
	}
	sort.Strings(result)
	return result
}
