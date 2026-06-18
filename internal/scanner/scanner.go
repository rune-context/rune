// Package scanner provides the file scanning and analysis engine.
package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rune-context/rune/internal/graph"
)

// FileInfo holds scanned metadata about a source file.
type FileInfo struct {
	Path         string   `json:"path"`
	RelPath      string   `json:"rel_path"`
	Language     string   `json:"language"`
	Size         int64    `json:"size"`
	Hash         string   `json:"hash"`
	ModTime      string   `json:"mod_time"`
	Exports      []string `json:"exports,omitempty"`
	Imports      []string `json:"imports,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	LineCount    int      `json:"line_count"`
}

// ScanResult holds the results of scanning a repository.
type ScanResult struct {
	Files    []*FileInfo `json:"files"`
	Graph    *graph.Graph
	Duration time.Duration `json:"duration"`
}

// Scanner scans a repository and extracts file metadata.
type Scanner struct {
	root    string
	ignore  []string
	cache   map[string]string // path -> hash for incremental updates
	mu      sync.Mutex
}

// New creates a new Scanner for the given repository root.
func New(root string, ignore []string) *Scanner {
	return &Scanner{
		root:   root,
		ignore: ignore,
		cache:  make(map[string]string),
	}
}

// SetCache loads a previous scan's hashes for incremental scanning.
func (s *Scanner) SetCache(cache map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = cache
}

// Scan performs a full repository scan.
func (s *Scanner) Scan() (*ScanResult, error) {
	start := time.Now()

	files, err := s.collectFiles()
	if err != nil {
		return nil, fmt.Errorf("collecting files: %w", err)
	}

	// Process files concurrently
	var wg sync.WaitGroup
	results := make([]*FileInfo, len(files))
	semaphore := make(chan struct{}, 32) // limit concurrency

	for i, path := range files {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			info, err := s.analyzeFile(p)
			if err != nil {
				return
			}
			results[idx] = info
		}(i, path)
	}
	wg.Wait()

	// Filter nil results
	var scanned []*FileInfo
	for _, r := range results {
		if r != nil {
			scanned = append(scanned, r)
		}
	}

	// Build graph
	g := graph.New()
	for _, f := range scanned {
		if len(f.Dependencies) > 0 {
			g.Set(f.RelPath, f.Dependencies)
		}
	}

	return &ScanResult{
		Files:    scanned,
		Graph:    g,
		Duration: time.Since(start),
	}, nil
}

// ScanChanged scans only files that have changed since last scan.
func (s *Scanner) ScanChanged() (*ScanResult, error) {
	start := time.Now()

	files, err := s.collectFiles()
	if err != nil {
		return nil, fmt.Errorf("collecting files: %w", err)
	}

	var changed []*FileInfo
	for _, path := range files {
		relPath, _ := filepath.Rel(s.root, path)
		relPath = filepath.ToSlash(relPath)

		hash, err := hashFile(path)
		if err != nil {
			continue
		}

		s.mu.Lock()
		prevHash, exists := s.cache[relPath]
		s.mu.Unlock()

		if exists && prevHash == hash {
			continue
		}

		info, err := s.analyzeFile(path)
		if err != nil {
			continue
		}
		changed = append(changed, info)
	}

	g := graph.New()
	for _, f := range changed {
		if len(f.Dependencies) > 0 {
			g.Set(f.RelPath, f.Dependencies)
		}
	}

	return &ScanResult{
		Files:    changed,
		Graph:    g,
		Duration: time.Since(start),
	}, nil
}

func (s *Scanner) collectFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(s.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}

		relPath, _ := filepath.Rel(s.root, path)
		relPath = filepath.ToSlash(relPath)

		if info.IsDir() {
			if s.shouldIgnoreDir(relPath, info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if s.shouldIgnoreFile(relPath, info.Name()) {
			return nil
		}

		if !isSourceFile(info.Name()) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

func (s *Scanner) shouldIgnoreDir(relPath, name string) bool {
	if name == "." {
		return false
	}
	for _, pattern := range s.ignore {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}
	// Always ignore hidden directories (except root)
	if strings.HasPrefix(name, ".") && relPath != "." {
		return true
	}
	return false
}

func (s *Scanner) shouldIgnoreFile(relPath, name string) bool {
	for _, pattern := range s.ignore {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}
	return false
}

func (s *Scanner) analyzeFile(path string) (*FileInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	relPath, _ := filepath.Rel(s.root, path)
	relPath = filepath.ToSlash(relPath)

	hash := hashBytes(content)

	// Update cache
	s.mu.Lock()
	s.cache[relPath] = hash
	s.mu.Unlock()

	lang := detectLanguage(filepath.Ext(path))
	text := string(content)
	lines := strings.Count(text, "\n") + 1

	exports := extractExports(text, lang)
	imports := extractImports(text, lang)
	deps := resolveLocalDeps(imports, relPath, s.root)

	return &FileInfo{
		Path:         path,
		RelPath:      relPath,
		Language:     lang,
		Size:         stat.Size(),
		Hash:         hash,
		ModTime:      stat.ModTime().UTC().Format(time.RFC3339),
		Exports:      exports,
		Imports:      imports,
		Dependencies: deps,
		LineCount:    lines,
	}, nil
}

// GetCache returns the current file hash cache.
func (s *Scanner) GetCache() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	cache := make(map[string]string, len(s.cache))
	for k, v := range s.cache {
		cache[k] = v
	}
	return cache
}

// --- Language detection ---

var extToLang = map[string]string{
	".go":    "go",
	".py":    "python",
	".js":    "javascript",
	".jsx":   "javascript",
	".ts":    "typescript",
	".tsx":   "typescript",
	".rs":    "rust",
	".java":  "java",
	".kt":    "kotlin",
	".rb":    "ruby",
	".php":   "php",
	".c":     "c",
	".h":     "c",
	".cpp":   "cpp",
	".hpp":   "cpp",
	".cs":    "csharp",
	".swift": "swift",
	".dart":  "dart",
	".lua":   "lua",
	".r":     "r",
	".scala": "scala",
	".ex":    "elixir",
	".exs":   "elixir",
	".erl":   "erlang",
	".zig":   "zig",
	".nim":   "nim",
	".v":     "vlang",
	".sh":    "shell",
	".bash":  "shell",
	".zsh":   "shell",
	".sql":   "sql",
	".vue":   "vue",
	".svelte":"svelte",
	".html":  "html",
	".css":   "css",
	".scss":  "scss",
	".less":  "less",
	".yaml":  "yaml",
	".yml":   "yaml",
	".toml":  "toml",
	".json":  "json",
	".xml":   "xml",
	".md":    "markdown",
	".proto": "protobuf",
	".graphql":"graphql",
	".gql":   "graphql",
}

func detectLanguage(ext string) string {
	ext = strings.ToLower(ext)
	if lang, ok := extToLang[ext]; ok {
		return lang
	}
	return "unknown"
}

func isSourceFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := extToLang[ext]
	return ok
}

// --- Import/Export extraction via regex ---

// Go patterns
var (
	goImportSingle = regexp.MustCompile(`import\s+"([^"]+)"`)
	goImportBlock  = regexp.MustCompile(`import\s*\(\s*([\s\S]*?)\s*\)`)
	goImportLine   = regexp.MustCompile(`"([^"]+)"`)
	goFuncExport   = regexp.MustCompile(`^func\s+([A-Z]\w*)\s*\(`)
	goTypeExport   = regexp.MustCompile(`^type\s+([A-Z]\w*)\s+`)
	goVarExport    = regexp.MustCompile(`^var\s+([A-Z]\w*)\s+`)
	goConstExport  = regexp.MustCompile(`^const\s+([A-Z]\w*)\s+`)
)

// Python patterns
var (
	pyImportFrom = regexp.MustCompile(`from\s+([\w.]+)\s+import`)
	pyImport     = regexp.MustCompile(`^import\s+([\w.]+)`)
	pyDefExport  = regexp.MustCompile(`^def\s+(\w+)\s*\(`)
	pyClassExport = regexp.MustCompile(`^class\s+(\w+)`)
)

// JavaScript/TypeScript patterns
var (
	jsImportFrom   = regexp.MustCompile(`import\s+.*?from\s+['"]([^'"]+)['"]`)
	jsRequire      = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	jsExportFunc   = regexp.MustCompile(`export\s+(?:default\s+)?(?:async\s+)?function\s+(\w+)`)
	jsExportClass  = regexp.MustCompile(`export\s+(?:default\s+)?class\s+(\w+)`)
	jsExportConst  = regexp.MustCompile(`export\s+(?:const|let|var)\s+(\w+)`)
	jsExportDefault = regexp.MustCompile(`export\s+default\s+(\w+)`)
)

// Rust patterns
var (
	rsUse      = regexp.MustCompile(`use\s+([\w:]+)`)
	rsMod      = regexp.MustCompile(`mod\s+(\w+)`)
	rsPubFn    = regexp.MustCompile(`pub\s+(?:async\s+)?fn\s+(\w+)`)
	rsPubStruct = regexp.MustCompile(`pub\s+struct\s+(\w+)`)
)

// Java/Kotlin patterns
var (
	javaImport    = regexp.MustCompile(`import\s+([\w.]+)`)
	javaClass     = regexp.MustCompile(`(?:public\s+)?class\s+(\w+)`)
	javaPublicFn  = regexp.MustCompile(`public\s+\w+\s+(\w+)\s*\(`)
)

func extractImports(content, lang string) []string {
	var imports []string
	seen := make(map[string]bool)

	addImport := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			imports = append(imports, s)
		}
	}

	switch lang {
	case "go":
		// Single imports
		for _, m := range goImportSingle.FindAllStringSubmatch(content, -1) {
			addImport(m[1])
		}
		// Block imports
		for _, m := range goImportBlock.FindAllStringSubmatch(content, -1) {
			for _, line := range goImportLine.FindAllStringSubmatch(m[1], -1) {
				addImport(line[1])
			}
		}

	case "python":
		for _, m := range pyImportFrom.FindAllStringSubmatch(content, -1) {
			addImport(m[1])
		}
		for _, line := range strings.Split(content, "\n") {
			if m := pyImport.FindStringSubmatch(line); m != nil {
				addImport(m[1])
			}
		}

	case "javascript", "typescript", "vue", "svelte":
		for _, m := range jsImportFrom.FindAllStringSubmatch(content, -1) {
			addImport(m[1])
		}
		for _, m := range jsRequire.FindAllStringSubmatch(content, -1) {
			addImport(m[1])
		}

	case "rust":
		for _, m := range rsUse.FindAllStringSubmatch(content, -1) {
			addImport(m[1])
		}
		for _, m := range rsMod.FindAllStringSubmatch(content, -1) {
			addImport(m[1])
		}

	case "java", "kotlin":
		for _, m := range javaImport.FindAllStringSubmatch(content, -1) {
			addImport(m[1])
		}
	}

	sort.Strings(imports)
	return imports
}

func extractExports(content, lang string) []string {
	var exports []string
	seen := make(map[string]bool)

	addExport := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			exports = append(exports, s)
		}
	}

	lines := strings.Split(content, "\n")

	switch lang {
	case "go":
		for _, line := range lines {
			if m := goFuncExport.FindStringSubmatch(line); m != nil {
				addExport(m[1])
			}
			if m := goTypeExport.FindStringSubmatch(line); m != nil {
				addExport(m[1])
			}
			if m := goVarExport.FindStringSubmatch(line); m != nil {
				addExport(m[1])
			}
			if m := goConstExport.FindStringSubmatch(line); m != nil {
				addExport(m[1])
			}
		}

	case "python":
		for _, line := range lines {
			if m := pyDefExport.FindStringSubmatch(line); m != nil {
				if !strings.HasPrefix(m[1], "_") {
					addExport(m[1])
				}
			}
			if m := pyClassExport.FindStringSubmatch(line); m != nil {
				addExport(m[1])
			}
		}

	case "javascript", "typescript", "vue", "svelte":
		for _, m := range jsExportFunc.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}
		for _, m := range jsExportClass.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}
		for _, m := range jsExportConst.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}
		for _, m := range jsExportDefault.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}

	case "rust":
		for _, m := range rsPubFn.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}
		for _, m := range rsPubStruct.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}

	case "java", "kotlin":
		for _, m := range javaClass.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}
		for _, m := range javaPublicFn.FindAllStringSubmatch(content, -1) {
			addExport(m[1])
		}
	}

	sort.Strings(exports)
	return exports
}

func resolveLocalDeps(imports []string, filePath, root string) []string {
	var deps []string
	dir := filepath.Dir(filePath)

	for _, imp := range imports {
		// Skip stdlib / external packages
		if isExternalImport(imp) {
			continue
		}

		// Try to resolve relative imports to actual files
		resolved := resolveImportPath(imp, dir, root)
		if resolved != "" {
			deps = append(deps, resolved)
		}
	}

	sort.Strings(deps)
	return deps
}

func isExternalImport(imp string) bool {
	// Go standard library or remote packages
	if strings.Contains(imp, ".") && strings.Contains(imp, "/") {
		return true
	}
	// Python stdlib
	pythonStdlib := map[string]bool{
		"os": true, "sys": true, "json": true, "re": true,
		"math": true, "datetime": true, "collections": true,
		"typing": true, "pathlib": true, "functools": true,
		"itertools": true, "abc": true, "io": true,
		"logging": true, "unittest": true, "dataclasses": true,
	}
	if pythonStdlib[imp] {
		return true
	}
	// Node built-in modules
	nodeBuiltin := map[string]bool{
		"fs": true, "path": true, "os": true, "http": true,
		"https": true, "crypto": true, "util": true, "stream": true,
		"events": true, "child_process": true, "url": true,
		"querystring": true, "buffer": true, "net": true,
	}
	if nodeBuiltin[imp] || strings.HasPrefix(imp, "node:") {
		return true
	}
	// Skip npm packages (no relative path prefix)
	if !strings.HasPrefix(imp, ".") && !strings.HasPrefix(imp, "/") {
		// Could be a local Python import though
		if !strings.Contains(imp, ".") {
			return false // might be local, let resolver handle it
		}
		return true
	}
	return false
}

func resolveImportPath(imp, dir, root string) string {
	// Handle relative JS/TS imports
	if strings.HasPrefix(imp, ".") {
		candidate := filepath.Join(dir, imp)
		candidate = filepath.ToSlash(candidate)

		// Try with extensions
		for _, ext := range []string{"", ".ts", ".tsx", ".js", ".jsx", ".py", ".go"} {
			check := candidate + ext
			fullPath := filepath.Join(root, check)
			if _, err := os.Stat(fullPath); err == nil {
				return check
			}
		}
		// Try as directory with index
		for _, idx := range []string{"/index.ts", "/index.tsx", "/index.js", "/index.jsx"} {
			check := candidate + idx
			fullPath := filepath.Join(root, check)
			if _, err := os.Stat(fullPath); err == nil {
				return check
			}
		}
	}

	// Handle Python relative imports (dot notation)
	if !strings.HasPrefix(imp, ".") && !strings.HasPrefix(imp, "/") {
		parts := strings.Split(imp, ".")
		candidate := filepath.Join(parts...)
		candidate = filepath.ToSlash(candidate)
		for _, ext := range []string{".py", ".go", ".ts", ".js"} {
			check := candidate + ext
			fullPath := filepath.Join(root, check)
			if _, err := os.Stat(fullPath); err == nil {
				return check
			}
		}
	}

	return ""
}

// --- Utility functions ---

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hashBytes(data), nil
}

func hashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:8]) // Use first 8 bytes for brevity
}
