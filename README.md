# Rune

**Repository Context Protocol**

Git stores history. Rune stores understanding.

Rune scans your repository and generates **`RUNE.md`** — a single skill file that any AI coding agent can read to understand your codebase. No MCP server needed. No configuration. Just a file.

Works with **Antigravity · Claude · Cursor · Codex · Windsurf · Copilot** and any agent that reads project files.

---

## Install

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/rune-context/rune/main/install.sh | sh
```

**Windows:**

```powershell
irm https://raw.githubusercontent.com/rune-context/rune/main/install.ps1 | iex
```

**From source:**

```bash
git clone https://github.com/rune-context/rune.git
cd rune
make install
```

---

## Quick Start

```bash
cd your-project
rune init       # Create .rune/ config
rune index      # Scan and generate RUNE.md
```

That's it. Your AI agents will now read `RUNE.md` automatically.

---

## What RUNE.md Contains

A single, human-readable markdown file with:

- **Architecture** — languages, file counts, line counts, structure
- **File Tree** — visual directory tree
- **File Map** — every file with its exports and dependencies
- **Dependencies** — what imports what
- **Key Exports** — all public functions, types, classes

Example output:

```markdown
## Architecture

**24 files** · **3774 lines**

Languages:
- go (15 files)
- typescript (8 files)

Structure:
- internal/: 9 files
- cmd/: 5 files

## File Map

### cmd/
- `cmd/index.go` (go, 73 lines) → Index
- `cmd/init.go` (go, 70 lines) → Init

## Key Exports

**internal/scanner/scanner.go**
- `FileInfo`
- `Scanner`
- `New`
```

---

## Commands

| Command | Description |
|---------|-------------|
| `rune init` | Initialize `.rune/` config |
| `rune index` | Full scan → generates `RUNE.md` |
| `rune update` | Incremental update of `RUNE.md` |
| `rune context <query>` | Get targeted context (stdout) |
| `rune doctor` | Check repository health |
| `rune serve` | Start MCP server (optional, stdio) |

---

## How Agents Use It

AI coding agents automatically discover and read project files like `RUNE.md`. No special setup needed:

| Agent | How it reads RUNE.md |
|-------|---------------------|
| **Antigravity** | Reads project files automatically |
| **Claude Code** | Reads project files in workspace |
| **Cursor** | Reads `.md` files in project root |
| **Codex** | Reads project context files |
| **Windsurf** | Reads project files automatically |
| **Copilot** | Reads via `@workspace` context |

For agents that support MCP, you can optionally run `rune serve` for richer tool-based access.

---

## Customize

Edit `.rune/spec.md` to add a project description:

```markdown
# My Project

A REST API for managing user accounts.

Backend: Go with Gin
Database: PostgreSQL
Auth: JWT tokens
```

Edit `.rune/conventions.md` to add coding rules:

```markdown
# Conventions

- Use repository pattern for data access
- Never use raw SQL, use sqlc
- All endpoints return JSON
- Tests required for all handlers
```

Both files are included in the generated `RUNE.md`.

---

## Supported Languages

Rune extracts imports, exports, and dependencies from:

Go · Python · JavaScript · TypeScript · Rust · Java · Kotlin · Ruby · PHP · C/C++ · C# · Swift · Dart · Lua · Scala · Elixir · Erlang · Zig · Vue · Svelte

---

## Performance

| Metric | Target |
|--------|--------|
| Cold index (10k files) | < 30s |
| Incremental update | < 1s |
| Startup | < 100ms |
| Memory | < 200MB |

---

## Design

- **Single binary** — no runtime dependencies
- **Single output** — one `RUNE.md` file
- **Local-first** — no login, no cloud, fully offline
- **Language-agnostic** — works with any language
- **Incremental** — only re-scans changed files
- **Human-readable** — `RUNE.md` is plain markdown

---

## License

MIT
