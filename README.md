# subgo

A Go library and CLI for processing subtitle files.

## Installation

```bash
go install github.com/kianelbo/subgo/cmd/subgo@latest
```

## CLI Usage

```bash
subgo <input-file> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-o, --output` | Output file (default: output.srt) |
| `-s, --shift` | Shift timing (e.g., `300ms`, `-1s`) |
| `-t, --stretch` | Stretch factor (e.g., `1.05` for 5% longer) |
| `-c, --clamp` | Clamp negative times to zero (default: true) |
| `--trim-first` | Remove first n events |
| `--trim-last` | Remove last n events |
| `--trim-before` | Remove events before timestamp |
| `--trim-after` | Remove events after timestamp |
| `--remove-hi` | Remove hearing impaired annotations |

#### Example

```bash
subgo eraserhead.srt --trim-before 1m --shift 300ms --remove-hi -o eraserhead.srt
```

## Library Usage

```go
import "github.com/kianelbo/subgo"

sub, _ := subgo.Load("input.ass")
sub = sub.Shift(300*time.Millisecond, true)
sub = sub.Stretch(1.05, 0)
sub.Save("output.srt")
```

## Supported Formats
- SRT
- ASS/SSA
