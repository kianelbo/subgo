package subgo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// vttFormat implements the WebVTT subtitle format.
type vttFormat struct{}

func (vttFormat) Name() string { return "vtt" }

func (vttFormat) Extensions() []string { return []string{".vtt"} }

func (vttFormat) Decode(r io.Reader) (Subtitle, error) {
	scanner := bufio.NewScanner(r)
	var events []Event

	// First line must be "WEBVTT" (possibly with additional text)
	if !scanner.Scan() {
		return Subtitle{}, fmt.Errorf("empty VTT file")
	}
	firstLine := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(firstLine, "WEBVTT") {
		return Subtitle{}, fmt.Errorf("invalid VTT file: missing WEBVTT header")
	}

	// Skip header metadata until first blank line
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			break
		}
	}

	for {
		// Skip blank lines and find next cue
		var line string
		for scanner.Scan() {
			line = strings.TrimSpace(scanner.Text())
			if line != "" {
				break
			}
		}
		if line == "" {
			break
		}

		// Skip NOTE blocks
		if strings.HasPrefix(line, "NOTE") {
			for scanner.Scan() {
				if strings.TrimSpace(scanner.Text()) == "" {
					break
				}
			}
			continue
		}

		// Skip STYLE blocks
		if strings.HasPrefix(line, "STYLE") {
			for scanner.Scan() {
				if strings.TrimSpace(scanner.Text()) == "" {
					break
				}
			}
			continue
		}

		// Skip REGION blocks
		if strings.HasPrefix(line, "REGION") {
			for scanner.Scan() {
				if strings.TrimSpace(scanner.Text()) == "" {
					break
				}
			}
			continue
		}

		// Check if this line is a timing line or a cue identifier
		var timing string
		if strings.Contains(line, "-->") {
			timing = line
		} else {
			// This is a cue identifier, next line should be timing
			if !scanner.Scan() {
				break
			}
			timing = strings.TrimSpace(scanner.Text())
		}

		start, end, err := parseVTTTimingLine(timing)
		if err != nil {
			// Skip malformed cues
			continue
		}

		// Read text lines until blank
		var textLines []string
		for scanner.Scan() {
			l := scanner.Text()
			if strings.TrimSpace(l) == "" {
				break
			}
			textLines = append(textLines, l)
		}

		text := strings.Join(textLines, "\n")
		events = append(events, Event{
			Start: start,
			End:   end,
			Text:  text,
		})
	}

	if err := scanner.Err(); err != nil {
		return Subtitle{}, err
	}

	return Subtitle{Events: events}, nil
}

func (vttFormat) Encode(w io.Writer, s Subtitle) error {
	buf := &bytes.Buffer{}

	// Write header
	buf.WriteString("WEBVTT\n\n")

	for i, e := range s.Events {
		// Write cue identifier
		fmt.Fprintf(buf, "%d\n", i+1)
		// Write timing line
		fmt.Fprintf(buf, "%s --> %s\n", formatVTTTimestamp(e.Start), formatVTTTimestamp(e.End))
		// Write text
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

func parseVTTTimingLine(line string) (time.Duration, time.Duration, error) {
	// Remove cue settings (anything after the end timestamp)
	parts := strings.Split(line, "-->")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid VTT timing line: %q", line)
	}

	startStr := strings.TrimSpace(parts[0])

	// End part may contain cue settings after the timestamp
	endPart := strings.TrimSpace(parts[1])
	endFields := strings.Fields(endPart)
	if len(endFields) == 0 {
		return 0, 0, fmt.Errorf("invalid VTT timing line: %q", line)
	}
	endStr := endFields[0]

	start, err := parseVTTTimestamp(startStr)
	if err != nil {
		return 0, 0, err
	}
	end, err := parseVTTTimestamp(endStr)
	if err != nil {
		return 0, 0, err
	}

	return start, end, nil
}

func parseVTTTimestamp(s string) (time.Duration, error) {
	// Format: HH:MM:SS.mmm or MM:SS.mmm
	s = strings.TrimSpace(s)

	// Split by dot for milliseconds
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid VTT timestamp: %q", s)
	}

	timePart := parts[0]
	msPart := parts[1]

	// Parse time part (either H:MM:SS or MM:SS)
	timeFields := strings.Split(timePart, ":")
	var h, m, sec int
	var err error

	switch len(timeFields) {
	case 2:
		// MM:SS format
		h = 0
		m, err = strconv.Atoi(timeFields[0])
		if err != nil {
			return 0, err
		}
		sec, err = strconv.Atoi(timeFields[1])
		if err != nil {
			return 0, err
		}
	case 3:
		// HH:MM:SS format
		h, err = strconv.Atoi(timeFields[0])
		if err != nil {
			return 0, err
		}
		m, err = strconv.Atoi(timeFields[1])
		if err != nil {
			return 0, err
		}
		sec, err = strconv.Atoi(timeFields[2])
		if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("invalid VTT timestamp: %q", s)
	}

	ms, err := strconv.Atoi(msPart)
	if err != nil {
		return 0, err
	}

	d := time.Hour*time.Duration(h) +
		time.Minute*time.Duration(m) +
		time.Second*time.Duration(sec) +
		time.Millisecond*time.Duration(ms)

	return d, nil
}

func formatVTTTimestamp(d time.Duration) string {
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

	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, sec, ms)
}
