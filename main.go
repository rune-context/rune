// Rune - Repository Context Protocol
//
// Scans your repository and generates RUNE.md — a single skill file
// that any AI coding agent can read to understand your codebase.
//
// Git stores history. Rune stores understanding.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rune-context/rune/cmd"
	"github.com/rune-context/rune/internal/mcp"
	"github.com/rune-context/rune/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "init":
		root := getRoot()
		if err := cmd.Init(root); err != nil {
			fatal(err)
		}

	case "index":
		root := getRoot()
		if err := cmd.Index(root); err != nil {
			fatal(err)
		}

	case "update":
		root := getRoot()
		if err := cmd.Update(root); err != nil {
			fatal(err)
		}

	case "context":
		root := getRoot()
		query := strings.Join(os.Args[2:], " ")
		if query == "" {
			fmt.Fprintln(os.Stderr, "Usage: rune context <query>")
			os.Exit(1)
		}
		if err := cmd.Context(root, query); err != nil {
			fatal(err)
		}

	case "doctor":
		root := getRoot()
		if err := cmd.Doctor(root); err != nil {
			fatal(err)
		}

	case "serve":
		root := getRoot()
		server := mcp.NewServer(root)
		if err := server.Serve(); err != nil {
			fatal(err)
		}

	case "--version", "-v", "version":
		fmt.Println(version.Short())

	case "--help", "-h", "help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func getRoot() string {
	// Check for --root flag
	for i, arg := range os.Args {
		if arg == "--root" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if strings.HasPrefix(arg, "--root=") {
			return strings.TrimPrefix(arg, "--root=")
		}
	}

	// Default to current directory
	dir, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	// Walk up to find .rune or .git
	for {
		if _, err := os.Stat(filepath.Join(dir, ".rune")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fall back to cwd
	cwd, _ := os.Getwd()
	return cwd
}

func printUsage() {
	fmt.Println(`Rune - Repository Context Protocol

  Generates RUNE.md — a skill file that any AI agent can read.
  Works with Antigravity, Claude, Cursor, Codex, Windsurf, and more.

Usage:
  rune <command> [options]

Commands:
  init       Initialize .rune/ in current repository
  index      Scan repository and generate RUNE.md
  update     Incrementally update RUNE.md
  context    Get context for a query (stdout)
  doctor     Check repository health
  serve      Start MCP server (optional, stdio)
  version    Show version

Options:
  --root     Specify repository root (default: auto-detect)
  --help     Show this help

Quick Start:
  rune init
  rune index    → generates RUNE.md at repo root

The RUNE.md file is automatically read by AI coding agents
as project context. No MCP server needed.`)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
