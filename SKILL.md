# Rune Context Skill

Rune is a Repository Context Protocol (RCP) tool. It enables AI coding agents to understand the workspace, navigate files, track dependencies, and adhere to conventions without reading every source file or exceeding context window limits.

---

## Capabilities

AI coding agents can use Rune to:
1. **Analyze Repository Structure:** View directory trees and file summaries.
2. **Resolve Dependencies:** Examine what files import or depend on other files.
3. **Retrieve Targeted Context:** Fetch a minimal relevant subset of files and conventions matching a specific programming task.
4. **Follow Conventions:** Read and adhere to human-defined project rules and style guidelines.

---

## Usage for AI Agents

When executing tasks in this repository, you should discover and utilize the files inside `.rune/` and the generated `RUNE.md` file in the root.

### Available CLI Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `rune index` | `rune index` | Performs a full scan of the codebase and regenerates `RUNE.md`. |
| `rune update` | `rune update` | Performs a fast, incremental update of the indexes for changed files. |
| `rune context` | `rune context "<query>"` | Returns a stdout block of relevant files and summaries for a specific query. |
| `rune doctor` | `rune doctor` | Checks repository index health and verifies dependency graph consistency. |
| `rune serve` | `rune serve` | Starts a Model Context Protocol (MCP) server over standard I/O. |

### Repository Structure

- **`RUNE.md`**: The master context file containing the repository overview, file tree, dependencies, key exports, and coding conventions.
- **`.rune/spec.md`**: High-level project specification.
- **`.rune/conventions.md`**: Custom development conventions and styling rules.
- **`.rune/graph.json`**: Complete dependency graph representation of the codebase.
- **`.rune/files/`**: Compact markdown summaries for each source file.
- **`.rune/features/`**: High-level workflow flowcharts and architectural flows.

---

## When to Run Rune Commands
- **After adding/modifying files:** Run `rune update` to synchronize the context index.
- **After structural changes:** Run `rune index` to rebuild the dependency graph and regenerate the full `RUNE.md` context.
- **Before major changes:** Run `rune context "<task description>"` to identify files related to the target feature.
