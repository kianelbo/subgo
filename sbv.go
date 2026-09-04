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

type sbvFormat struct{}

func (sbvFormat) Name() string { return "sbv" }

func (sbvFormat) Extensions() []string { return []string{".sbv"} }

func (sbvFormat) Decode(r io.Reader) (Subtitle, error) {
	scanner := bufio.NewScanner(r)
	var events []Event
	var timing string
	var textLines []string

	flush := func() error {
		if timing == "" {
			return nil
		}
		start, end, err := parseSBVTimingLine(timing)
		if err != nil {
			return err
		}
		events = append(events, Event{Start: start, End: end, Text: strings.Join(textLines, "\n")})
		timing = ""
		textLines = nil
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if err := flush(); err != nil {
				return Subtitle{}, err
			}
			continue
		}
		if timing == "" {
			timing = strings.TrimSpace(line)
		} else {
			textLines = append(textLines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return Subtitle{}, err
	}
	if err := flush(); err != nil {
		return Subtitle{}, err
	}

	return Subtitle{Events: events}, nil
}

func (sbvFormat) Encode(w io.Writer, s Subtitle) error {
	buf := &bytes.Buffer{}
	for _, e := range s.Events {
		fmt.Fprintf(buf, "%s,%s\n", formatSBVTimestamp(e.Start), formatSBVTimestamp(e.End))
		if e.Text != "" {
			buf.WriteString(e.Text)
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func parseSBVTimingLine(line string) (time.Duration, time.Duration, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid SBV timing line: %q", line)
	}
	start, err := parseSBVTimestamp(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	end, err := parseSBVTimestamp(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func parseSBVTimestamp(s string) (time.Duration, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid SBV timestamp: %q", s)
	}
	timeFields := strings.Split(parts[0], ":")
	if len(timeFields) != 3 {
		return 0, fmt.Errorf("invalid SBV timestamp: %q", s)
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
	ms, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return time.Hour*time.Duration(h) + time.Minute*time.Duration(m) + time.Second*time.Duration(sec) + time.Millisecond*time.Duration(ms), nil
}

func formatSBVTimestamp(d time.Duration) string {
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
	return fmt.Sprintf("%d:%02d:%02d.%03d", h, m, sec, ms)
}
