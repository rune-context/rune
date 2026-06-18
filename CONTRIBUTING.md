# Contributing to Rune

Thank you for your interest in contributing to Rune! We want to keep the codebase simple, fast, and easy to maintain.

Please review the following guidelines before submitting pull requests or making modifications to the codebase.

---

## Design Philosophy

All contributions must respect the foundational principles of Rune:

1. **Local-First & Offline:** Rune runs entirely on the local machine. Never add network dependencies, cloud APIs, telemetry, or remote lookups to the core CLI.
2. **Zero Runtime Dependencies:** Rune is compiled as a single static binary. Rely on Go's standard library. Do not pull in large third-party frameworks or SDKs.
3. **Performance First:** Scanning repositories must be extremely fast. Use regex-based heuristics for parsing files; do not run full AST parsers.
4. **Platform-Agnostic:** Normalize all paths using `/` (via `filepath.ToSlash`) to ensure consistent behavior across Windows, macOS, and Linux.

---

## Development Setup

To build and test Rune locally, you will need Go installed.

### Build from Source
```bash
make build
```

### Run Tests
```bash
make test
```

---

## Coding Conventions

### 1. Package Structure
- **`cmd/`**: Contains CLI command entry points and routing (e.g., `cmd.Init`, `cmd.Update`).
- **`internal/`**: Contains sub-modules implementing specific logic (e.g., `scanner`, `mcp`, `graph`, `summary`). Code in `internal` is not importable by external packages.
- **`main.go`**: The CLI router that parses arguments and invokes commands from the `cmd` package.

### 2. Error Handling
- Never ignore errors. Return errors up the call stack.
- Wrap errors at package boundaries to provide contextual trace details:
  ```go
  return fmt.Errorf("loading config: %w", err)
  ```

### 3. Concurrency
- When processing files concurrently (e.g., in the scanner), limit the number of goroutines using a semaphore channel to prevent running out of file descriptors:
  ```go
  semaphore := make(chan struct{}, 32)
  ```

### 4. Idempotency & Safe Updates
- Utility scripts or initialization commands must be idempotent.
- Never overwrite human-owned files in `.rune/` (such as `spec.md` or `conventions.md`) without checking for their existence or receiving explicit instruction.

---

## Keeping Context Updated

Rune is self-describing. When you add new CLI commands or modify internal systems:
1. Update `.rune/spec.md` if the overall capabilities/scope of Rune changes.
2. Update `.rune/conventions.md` if you introduce new development rules or architectural requirements.
3. Add a feature flow description under `.rune/features/` if you add a new system flow.
4. Regenerate `RUNE.md` by building and running the binary:
   ```bash
   ./rune index
   ```
