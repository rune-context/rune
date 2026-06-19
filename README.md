```text
╭──────────────────────────────────────────────────────────────╮
│ * Welcome to the Rune context index !                      │
╰──────────────────────────────────────────────────────────────╯

╔══════════╗    ╔══╗    ╔══╗    ╔══╗        ╔══╗    ╔══════════╗
║██████████║    ║██║    ║██║    ║██║        ║██║    ║██████████║
║██╔════╗██║    ║██║    ║██║    ║██╚═╗      ║██║    ║██╔═══════╝
║██║    ║██║    ║██║    ║██║    ║████║      ║██║    ║██║        
║██╚════╝██║    ║██║    ║██║    ║██╔═╝╚═╗   ║██║    ║██╚═════╗  
║██████████║    ║██║    ║██║    ║██║  ║██║  ║██║    ║████████║  
║██╔════╗██║    ║██║    ║██║    ║██║  ║██╚═╗║██║    ║██╔═════╝  
║██║    ║██║    ║██╚════╝██║    ║██║  ║████║║██║    ║██║        
║██║    ║██║    ║██████████║    ║██║  ╚════╝║██║    ║██████████║
╚══╝    ╚══╝    ╚══════════╝    ╚══╝        ╚══╝    ╚══════════╝
```

**Repository Context Protocol**

Git stores history. Rune stores understanding.

Rune scans your repository and generates **`RUNE.md`** — a single skill file that any AI coding agent can read to understand your codebase. No MCP server needed. No configuration. Just a file.

Works with **Antigravity · Claude · Cursor · Codex · Windsurf · Copilot** and any agent that reads project files.

<br/>

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

<br/>

## Quick Start

```bash
cd your-project
rune init       # Create .rune/ config
rune index      # Scan and generate RUNE.md
```

That's it. Your AI agents will now read `RUNE.md` automatically.

<br/>

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

<br/>

## Commands

| Command | Description |
|---------|-------------|
| `rune init` | Initialize `.rune/` config |
| `rune index` | Full scan → generates `RUNE.md` |
| `rune update` | Incremental update of `RUNE.md` |
| `rune context <query>` | Get targeted context (stdout) |
| `rune doctor` | Check repository health |
| `rune serve` | Start MCP server (optional, stdio) |

<br/>

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

<br/>

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

> [!NOTE]
> **Ownership & LLM Updates**
> - **Human-authored:** These files are intended to be human-authored guardrails. Rune is local-first and offline; it will never automatically overwrite or modify these files during `rune update` or `rune index`.
> - **Agent updates:** AI coding agents should avoid modifying these files unless you explicitly instruct them to document a new architectural decision or project convention.

<br/>

## Supported Languages

Rune extracts imports, exports, and dependencies from:

Go · Python · JavaScript · TypeScript · Rust · Java · Kotlin · Ruby · PHP · C/C++ · C# · Swift · Dart · Lua · Scala · Elixir · Erlang · Zig · Vue · Svelte

<br/>

## Performance

| Metric | Target |
|--------|--------|
| Cold index (10k files) | < 30s |
| Incremental update | < 1s |
| Startup | < 100ms |
| Memory | < 200MB |

<br/>

## Design

- **Single binary** — no runtime dependencies
- **Single output** — one `RUNE.md` file
- **Local-first** — no login, no cloud, fully offline
- **Language-agnostic** — works with any language
- **Incremental** — only re-scans changed files
- **Human-readable** — `RUNE.md` is plain markdown

<br/>

## Support

Rune Ctx is, and always will be, completely free and open-source.
If you find it useful and want to support its continued development, you can buy me a coffee!

[![Donate via PayPal](https://img.shields.io/badge/DONATE-PAYPAL-00457c?logo=paypal)](https://paypal.me/wawanbsetyawan)

<br/>

## License

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center">
  <img src="runes.jpg" alt="Rune Logo" width="600">
</p>