# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`pkgsh` — unified terminal package manager for Ubuntu/Debian. TUI built with Go + Bubble Tea that abstracts apt, snap, flatpak, dpkg, pip, npm, and AppImage into a single interface.

Full design spec: `docs/superpowers/specs/2026-04-05-pkgsh-design.md`

## Build & Run

```bash
go build ./cmd/pkgsh          # build binary
go run ./cmd/pkgsh            # run without building
CGO_ENABLED=0 go build -o pkgsh ./cmd/pkgsh  # static binary for distribution
```

## Test

```bash
go test ./...                 # all tests
go test ./internal/adapters/apt/...  # single adapter
go test -run TestName ./...   # single test
```

## Release

```bash
# .deb packaging (requires nfpm installed)
nfpm package --packager deb --target dist/
```

## Architecture

Four strict layers — dependencies only flow downward:

1. **UI** (`internal/ui/`) — Bubble Tea models for list, detail, and log panels
2. **Domain** (`internal/domain/`) — `PackageManager` interface, `Package` struct, `Operation` type, filter/sort logic
3. **Adapters** (`internal/adapters/<manager>/`) — one package per manager, each implements `PackageManager`
4. **System** — `exec.Cmd` with streaming stdout/stderr; no shell interpolation, always `[]string` args

At startup, all adapters run in parallel goroutines and stream results into the UI progressively. State lives in memory only — no disk cache.

## Key Design Rules

- **No shell string interpolation ever.** All commands are `[]string` slices passed to `exec.Cmd` directly. This is the primary security invariant.
- **`PackageManager` interface is the adapter contract.** Adding a new manager = implement the interface + register it. No other layers change.
- **`Operation` streams output.** Operations return an `*Operation` that exposes an `io.Reader` — the UI reads from it to display real-time log output.
- **Bulk operations group by manager.** When multiple packages from different managers are selected, commands are batched per manager (e.g., one `apt remove` call, one `npm uninstall` call).
- **sudo prompt is intercepted.** If a sudo password prompt is detected in the output stream, the UI opens a secure modal input instead of failing.

## Flags

```bash
pkgsh --manager apt|snap|flatpak|dpkg|pip|npm|appimage  # pre-filter by manager (comma-separated)
pkgsh --upgradeable    # show only packages with updates available
pkgsh --native         # show only OS-native packages
pkgsh --search <term>  # start with search active
```

## Distribution

- Static binary: `CGO_ENABLED=0 go build`
- `.deb`: generated via `nfpm` using `nfpm.yaml`
- Releases via GitHub Actions on `tag v*` — builds amd64 + arm64, packages `.deb`, publishes to GitHub Releases with SHA256 checksums
- License: MIT
