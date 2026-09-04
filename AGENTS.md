# Agent Guide

## Project Shape

- The root package is the public subtitle library. Start with [subtitle.go](subtitle.go) for the `Event` and `Subtitle` model and immutable-style transformations.
- [format.go](format.go) owns the `Format` interface, extension-based format detection, and file loading/saving.
- [srt.go](srt.go), [ass.go](ass.go), [vtt.go](vtt.go), and [sbv.go](sbv.go) implement the supported codecs. Keep concrete format types unexported and register them in `format.go`.
- [cmd/subgo/main.go](cmd/subgo/main.go) is the Cobra CLI boundary. It loads, applies operations, and saves; library behavior belongs in the root package.

## Development Commands

Run these from the repository root:

```text
go test ./... -v
go vet ./...
gofmt -l .
```

CI also uses `golangci-lint` with [.golangci.yml](.golangci.yml). The module targets Go 1.21; the linter configuration targets Go 1.22.

## Conventions

- Use table-driven tests with `t.Run`; use `t.TempDir()` for filesystem tests.
- Return errors from library and codec code. The CLI wraps load and save failures with `load: ...` and `save: ...`.
- Add or update focused package tests with changes to parsing, encoding, transformations, or format detection. CLI operation wiring currently has no dedicated tests.
- Preserve the public API and standard-library-only subtitle processing unless a dependency is necessary.

## Behavior To Preserve

- Format detection is based on the input/output filename extension and is case-insensitive; extensionless or unknown extensions fail.
- CLI operations run in this order: remove IDs, trim, shift, stretch, then remove hearing-impaired annotations.
- `RemoveIds` accepts 1-based event indices.
- ASS encoding uses centiseconds, so round-trip timestamps can lose sub-centisecond precision.
- ASS and VTT skip malformed cues, while SRT and SBV return an error for malformed timing.
- `Save` creates or truncates the destination before encoding completes; consider this when changing file I/O.

See [README.md](README.md) for user-facing installation, CLI flags, examples, and supported formats.