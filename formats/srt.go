package formats

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"subgo/subtitle"
)

// srtFormat implements the SRT subtitle format.
type srtFormat struct{}

func (srtFormat) Name() string { return "srt" }

func (srtFormat) Extensions() []string { return []string{".srt"} }

func (srtFormat) Decode(r io.Reader) (subtitle.Subtitle, error) {
	scanner := bufio.NewScanner(r)
	var events []subtitle.Event

	for {
		// Read index line (ignored)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Try to parse index, but ignore value
		if _, err := strconv.Atoi(line); err != nil {
			// not a valid index; try to treat as timing line
			// this also allows slightly malformed files
		}

		// Timing line
		if !scanner.Scan() {
			break
		}
		timing := strings.TrimSpace(scanner.Text())
		start, end, err := parseTimingLine(timing)
		if err != nil {
			return subtitle.Subtitle{}, err
		}

		// Text lines until blank
		var textLines []string
		for scanner.Scan() {
			l := scanner.Text()
			if strings.TrimSpace(l) == "" {
				break
			}
			textLines = append(textLines, l)
		}

		text := strings.Join(textLines, "\n")
		events = append(events, subtitle.Event{
			Start: start,
			End:   end,
			Text:  text,
		})
	}

	if err := scanner.Err(); err != nil {
		return subtitle.Subtitle{}, err
	}

	return subtitle.Subtitle{Events: events}, nil
}

func (srtFormat) Encode(w io.Writer, s subtitle.Subtitle) error {
	buf := &bytes.Buffer{}
	for i, e := range s.Events {
		fmt.Fprintf(buf, "%d\n", i+1)
		fmt.Fprintf(buf, "%s --> %s\n", formatTimestamp(e.Start), formatTimestamp(e.End))
		if e.Text != "" {
			lines := strings.Split(e.Text, "\n")
			for _, l := range lines {
				buf.WriteString(l)
				buf.WriteByte('\n')
			}
		}
		buf.WriteByte('\n')
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func parseTimingLine(line string) (time.Duration, time.Duration, error) {
	parts := strings.Split(line, "-->")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid SRT timing line: %q", line)
	}
	start, err := parseTimestamp(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	end, err := parseTimestamp(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func parseTimestamp(s string) (time.Duration, error) {
	// Format: HH:MM:SS,mmm
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid SRT timestamp: %q", s)
	}
	timePart := parts[0]
	msPart := parts[1]

	timeFields := strings.Split(timePart, ":")
	if len(timeFields) != 3 {
		return 0, fmt.Errorf("invalid SRT timestamp: %q", s)
	}

	h, err := strconv.Atoi(timeFields[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(timeFields[1])
	if err != nil {
		return 0, err
	}
	sec, err := strconv.Atoi(timeFields[2])
	if err != nil {
		return 0, err
	}
	ms, err := strconv.Atoi(msPart)
	if err != nil {
		return 0, err
	}

	d := time.Hour*time.Duration(h) + time.Minute*time.Duration(m) + time.Second*time.Duration(sec) + time.Millisecond*time.Duration(ms)
	return d, nil
}

func formatTimestamp(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d / time.Hour)
	d -= time.Duration(h) * time.Hour
	m := int(d / time.Minute)
	d -= time.Duration(m) * time.Minute
	sec := int(d / time.Second)
	d -= time.Duration(sec) * time.Second
	ms := int(d / time.Millisecond)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, sec, ms)
}
