// Package summary generates the RUNE.md skill file and internal artifacts.
package summary

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rune-context/rune/internal/config"
	"github.com/rune-context/rune/internal/graph"
	"github.com/rune-context/rune/internal/scanner"
)

// GenerateRuneMD produces the complete RUNE.md skill file content.
func GenerateRuneMD(root string, files []*scanner.FileInfo, g *graph.Graph) string {
	var sb strings.Builder

	// --- Header ---
	sb.WriteString("# RUNE.md\n\n")
	sb.WriteString("> Auto-generated repository context. Do not edit manually.\n")
	sb.WriteString("> Regenerate with: `rune index`\n\n")

	// --- Project description (from .rune/spec.md if exists) ---
	specPath := config.SubPath(root, config.SpecFile)
	if data, err := os.ReadFile(specPath); err == nil {
		content := strings.TrimSpace(string(data))
		if content != "" && content != "# Project Spec\n\nDescribe your project here." {
			sb.WriteString("---\n\n")
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}

	// --- Architecture ---
	sb.WriteString("---\n\n")
	sb.WriteString(generateArchitecture(files))

	// --- Conventions (from .rune/conventions.md if exists) ---
	convPath := config.SubPath(root, config.ConventionsFile)
	if data, err := os.ReadFile(convPath); err == nil {
		content := strings.TrimSpace(string(data))
		if content != "" && content != "# Conventions\n\nAdd your coding conventions here." {
			sb.WriteString("---\n\n")
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}

	// --- Directory tree ---
	sb.WriteString("---\n\n")
	sb.WriteString(generateTree(files))

	// --- File Map (compact: one line per file with purpose) ---
	sb.WriteString("---\n\n")
	sb.WriteString(generateFileMap(files, g))

	// --- Dependency Graph (only files with deps) ---
	if len(g.Files()) > 0 {
		sb.WriteString("---\n\n")
		sb.WriteString(generateGraphSection(g))
	}

	// --- Key Exports ---
	sb.WriteString("---\n\n")
	sb.WriteString(generateExportsSection(files))

	return sb.String()
}

// WriteRuneMD writes the RUNE.md file to the repository root.
func WriteRuneMD(root string, files []*scanner.FileInfo, g *graph.Graph) error {
	content := GenerateRuneMD(root, files, g)
	path := filepath.Join(root, "RUNE.md")
	return os.WriteFile(path, []byte(content), 0644)
}

// WriteFileSummaries writes individual file summaries to .rune/files/.
func WriteFileSummaries(root string, files []*scanner.FileInfo, g *graph.Graph) error {
	dir := config.SubPath(root, config.FilesDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for _, f := range files {
		sum := generateFileSummary(f, g)
		safeName := strings.ReplaceAll(f.RelPath, "/", "__")
		safeName = strings.TrimSuffix(safeName, filepath.Ext(safeName)) + ".md"
		path := filepath.Join(dir, safeName)
		if err := os.WriteFile(path, []byte(sum), 0644); err != nil {
			return fmt.Errorf("writing summary for %s: %w", f.RelPath, err)
		}
	}
	return nil
}

// WriteArchitecture writes the architecture summary to .rune/architecture.md.
func WriteArchitecture(root string, files []*scanner.FileInfo) error {
	content := generateArchitecture(files)
	path := config.SubPath(root, config.ArchitectureFile)
	return os.WriteFile(path, []byte(content), 0644)
}

// --- Internal generators ---

func generateArchitecture(files []*scanner.FileInfo) string {
	var sb strings.Builder
	sb.WriteString("## Architecture\n\n")

	// Languages with counts, sorted by count desc
	langCount := make(map[string]int)
	totalLines := 0
	for _, f := range files {
		if f.Language != "unknown" {
			langCount[f.Language]++
		}
		totalLines += f.LineCount
	}

	type langInfo struct {
		name  string
		count int
	}
	var langs []langInfo
	for l, c := range langCount {
		langs = append(langs, langInfo{l, c})
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].count > langs[j].count })

	sb.WriteString(fmt.Sprintf("**%d files** · **%d lines**\n\n", len(files), totalLines))

	sb.WriteString("Languages:\n")
	for _, l := range langs {
		sb.WriteString(fmt.Sprintf("- %s (%d files)\n", l.name, l.count))
	}
	sb.WriteString("\n")

	// Top-level directory structure
	dirs := make(map[string]int)
	for _, f := range files {
		parts := strings.Split(f.RelPath, "/")
		if len(parts) > 1 {
			dirs[parts[0]]++
		} else {
			dirs["."]++
		}
	}

	type dirInfo struct {
		name  string
		count int
	}
	var dirList []dirInfo
	for d, c := range dirs {
		dirList = append(dirList, dirInfo{d, c})
	}
	sort.Slice(dirList, func(i, j int) bool { return dirList[i].count > dirList[j].count })

	sb.WriteString("Structure:\n")
	for _, d := range dirList {
		if d.name == "." {
			sb.WriteString(fmt.Sprintf("- (root): %d files\n", d.count))
		} else {
			sb.WriteString(fmt.Sprintf("- %s/: %d files\n", d.name, d.count))
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

func generateTree(files []*scanner.FileInfo) string {
	var sb strings.Builder
	sb.WriteString("## File Tree\n\n")
	sb.WriteString("```\n")

	// Build tree from paths
	type treeNode struct {
		children map[string]*treeNode
		isFile   bool
		lang     string
	}

	root := &treeNode{children: make(map[string]*treeNode)}

	for _, f := range files {
		parts := strings.Split(f.RelPath, "/")
		node := root
		for i, part := range parts {
			if node.children[part] == nil {
				node.children[part] = &treeNode{children: make(map[string]*treeNode)}
			}
			node = node.children[part]
			if i == len(parts)-1 {
				node.isFile = true
				node.lang = f.Language
			}
		}
	}

	// Render tree
	var renderTree func(node *treeNode, prefix string, isLast bool, name string)
	renderTree = func(node *treeNode, prefix string, isLast bool, name string) {
		if name != "" {
			connector := "├── "
			if isLast {
				connector = "└── "
			}
			sb.WriteString(prefix + connector + name)
			if !node.isFile && len(node.children) > 0 {
				sb.WriteString("/")
			}
			sb.WriteString("\n")
		}

		// Sort children: directories first, then files
		var childDirs, childFiles []string
		for k, v := range node.children {
			if !v.isFile || len(v.children) > 0 {
				childDirs = append(childDirs, k)
			} else {
				childFiles = append(childFiles, k)
			}
		}
		sort.Strings(childDirs)
		sort.Strings(childFiles)
		allChildren := append(childDirs, childFiles...)

		for i, child := range allChildren {
			newPrefix := prefix
			if name != "" {
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
			}
			renderTree(node.children[child], newPrefix, i == len(allChildren)-1, child)
		}
	}

	renderTree(root, "", true, "")
	sb.WriteString("```\n\n")
	return sb.String()
}

func generateFileMap(files []*scanner.FileInfo, g *graph.Graph) string {
	var sb strings.Builder
	sb.WriteString("## File Map\n\n")

	// Group by top-level directory
	groups := make(map[string][]*scanner.FileInfo)
	for _, f := range files {
		parts := strings.Split(f.RelPath, "/")
		group := "."
		if len(parts) > 1 {
			group = parts[0]
		}
		groups[group] = append(groups[group], f)
	}

	// Sort group keys
	var keys []string
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, group := range keys {
		groupFiles := groups[group]
		if group == "." {
			sb.WriteString("### Root\n\n")
		} else {
			sb.WriteString(fmt.Sprintf("### %s/\n\n", group))
		}

		for _, f := range groupFiles {
			// Compact one-liner: path | language | exports
			exports := ""
			if len(f.Exports) > 0 {
				maxExports := 8
				if len(f.Exports) < maxExports {
					maxExports = len(f.Exports)
				}
				exports = " → " + strings.Join(f.Exports[:maxExports], ", ")
				if len(f.Exports) > maxExports {
					exports += fmt.Sprintf(" (+%d more)", len(f.Exports)-maxExports)
				}
			}

			deps := ""
			if len(f.Dependencies) > 0 {
				deps = fmt.Sprintf(" ← [%s]", strings.Join(f.Dependencies, ", "))
			}

			sb.WriteString(fmt.Sprintf("- `%s` (%s, %d lines)%s%s\n",
				f.RelPath, f.Language, f.LineCount, exports, deps))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func generateGraphSection(g *graph.Graph) string {
	var sb strings.Builder
	sb.WriteString("## Dependencies\n\n")

	files := g.Files()
	for _, f := range files {
		deps := g.Get(f)
		if len(deps) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("- `%s`\n", f))
		for _, d := range deps {
			sb.WriteString(fmt.Sprintf("  - → `%s`\n", d))
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

func generateExportsSection(files []*scanner.FileInfo) string {
	var sb strings.Builder
	sb.WriteString("## Key Exports\n\n")

	// Only show files with exports, grouped by language
	langFiles := make(map[string][]*scanner.FileInfo)
	for _, f := range files {
		if len(f.Exports) > 0 {
			langFiles[f.Language] = append(langFiles[f.Language], f)
		}
	}

	if len(langFiles) == 0 {
		sb.WriteString("No public exports detected.\n\n")
		return sb.String()
	}

	var langs []string
	for l := range langFiles {
		langs = append(langs, l)
	}
	sort.Strings(langs)

	for _, lang := range langs {
		lf := langFiles[lang]
		for _, f := range lf {
			sb.WriteString(fmt.Sprintf("**%s**\n", f.RelPath))
			for _, e := range f.Exports {
				sb.WriteString(fmt.Sprintf("- `%s`\n", e))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func generateFileSummary(info *scanner.FileInfo, g *graph.Graph) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", info.RelPath))
	sb.WriteString(fmt.Sprintf("Language: %s\n\n", info.Language))
	sb.WriteString(fmt.Sprintf("Lines: %d\n\n", info.LineCount))

	if len(info.Exports) > 0 {
		sb.WriteString("Exports:\n")
		for _, e := range info.Exports {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteString("\n")
	}

	if len(info.Dependencies) > 0 {
		sb.WriteString("Dependencies:\n")
		for _, d := range info.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", d))
		}
		sb.WriteString("\n")
	}

	dependents := g.Dependents(info.RelPath)
	if len(dependents) > 0 {
		sb.WriteString("Used by:\n")
		for _, d := range dependents {
			sb.WriteString(fmt.Sprintf("- %s\n", d))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
