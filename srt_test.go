package subgo

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSRTDecode(t *testing.T) {
	input := `1
00:00:01,000 --> 00:00:02,500
Hello World

2
00:00:03,000 --> 00:00:04,000
Second line
With multiple lines

3
00:01:30,500 --> 00:01:35,000
Third subtitle
`

	format := srtFormat{}
	sub, err := format.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if len(sub.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(sub.Events))
	}

	tests := []struct {
		idx   int
		start time.Duration
		end   time.Duration
		text  string
	}{
		{0, 1 * time.Second, 2500 * time.Millisecond, "Hello World"},
		{1, 3 * time.Second, 4 * time.Second, "Second line\nWith multiple lines"},
		{2, 90500 * time.Millisecond, 95 * time.Second, "Third subtitle"},
	}

	for _, tt := range tests {
		e := sub.Events[tt.idx]
		if e.Start != tt.start {
			t.Errorf("event %d: start = %v, want %v", tt.idx, e.Start, tt.start)
		}
		if e.End != tt.end {
			t.Errorf("event %d: end = %v, want %v", tt.idx, e.End, tt.end)
		}
		if e.Text != tt.text {
			t.Errorf("event %d: text = %q, want %q", tt.idx, e.Text, tt.text)
		}
	}
}

func TestSRTDecodeEmptyFile(t *testing.T) {
	format := srtFormat{}
	sub, err := format.Decode(strings.NewReader(""))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(sub.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(sub.Events))
	}
}

func TestSRTDecodeWithExtraBlankLines(t *testing.T) {
	input := `

1
00:00:01,000 --> 00:00:02,000
First


2
00:00:03,000 --> 00:00:04,000
Second

`

	format := srtFormat{}
	sub, err := format.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if len(sub.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(sub.Events))
	}
}

func TestSRTEncode(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2500 * time.Millisecond, Text: "Hello World"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second line\nWith newline"},
		},
	}

	format := srtFormat{}
	var buf bytes.Buffer
	err := format.Encode(&buf, sub)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	output := buf.String()

	// Check structure
	if !strings.Contains(output, "1\n00:00:01,000 --> 00:00:02,500\nHello World") {
		t.Errorf("first subtitle not encoded correctly:\n%s", output)
	}
	if !strings.Contains(output, "2\n00:00:03,000 --> 00:00:04,000\nSecond line\nWith newline") {
		t.Errorf("second subtitle not encoded correctly:\n%s", output)
	}
}

func TestSRTEncodeEmpty(t *testing.T) {
	sub := Subtitle{}

	format := srtFormat{}
	var buf bytes.Buffer
	err := format.Encode(&buf, sub)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestSRTRoundTrip(t *testing.T) {
	original := Subtitle{
		Events: []Event{
			{Start: 0, End: 1 * time.Second, Text: "First"},
			{Start: 1*time.Hour + 30*time.Minute + 45*time.Second + 123*time.Millisecond, End: 1*time.Hour + 30*time.Minute + 50*time.Second, Text: "Long timestamp"},
			{Start: 5 * time.Second, End: 10 * time.Second, Text: "Multi\nLine\nText"},
		},
	}

	format := srtFormat{}

	// Encode
	var buf bytes.Buffer
	if err := format.Encode(&buf, original); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Decode
	decoded, err := format.Decode(&buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	// Compare
	if len(decoded.Events) != len(original.Events) {
		t.Fatalf("event count mismatch: got %d, want %d", len(decoded.Events), len(original.Events))
	}

	for i, orig := range original.Events {
		dec := decoded.Events[i]
		if dec.Start != orig.Start {
			t.Errorf("event %d: start = %v, want %v", i, dec.Start, orig.Start)
		}
		if dec.End != orig.End {
			t.Errorf("event %d: end = %v, want %v", i, dec.End, orig.End)
		}
		if dec.Text != orig.Text {
			t.Errorf("event %d: text = %q, want %q", i, dec.Text, orig.Text)
		}
	}
}

func TestSRTTimestampParsing(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"00:00:00,000", 0},
		{"00:00:01,000", 1 * time.Second},
		{"00:00:00,001", 1 * time.Millisecond},
		{"00:00:00,999", 999 * time.Millisecond},
		{"00:01:00,000", 1 * time.Minute},
		{"01:00:00,000", 1 * time.Hour},
		{"12:34:56,789", 12*time.Hour + 34*time.Minute + 56*time.Second + 789*time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTimestamp(tt.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSRTTimestampParsingErrors(t *testing.T) {
	tests := []string{
		"",
		"00:00:00",     // missing milliseconds
		"00:00:00.000", // wrong separator (dot instead of comma)
		"invalid",
		"00:00,000",    // missing seconds
		"00:00:00:000", // wrong separator
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseTimestamp(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestSRTTimestampParsingEdgeCases(t *testing.T) {
	// These parse successfully even if not ideal
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"00:00:00,00", 0}, // short ms parses as 0
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTimestamp(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSRTTimestampFormatting(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{0, "00:00:00,000"},
		{1 * time.Second, "00:00:01,000"},
		{1 * time.Millisecond, "00:00:00,001"},
		{999 * time.Millisecond, "00:00:00,999"},
		{1 * time.Minute, "00:01:00,000"},
		{1 * time.Hour, "01:00:00,000"},
		{12*time.Hour + 34*time.Minute + 56*time.Second + 789*time.Millisecond, "12:34:56,789"},
		{-1 * time.Second, "00:00:00,000"}, // negative clamped to zero
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSRTTimingLineParsing(t *testing.T) {
	tests := []struct {
		input     string
		wantStart time.Duration
		wantEnd   time.Duration
	}{
		{"00:00:01,000 --> 00:00:02,000", 1 * time.Second, 2 * time.Second},
		{"00:00:00,000 --> 00:00:00,500", 0, 500 * time.Millisecond},
		{"01:30:00,000 --> 02:00:00,000", 90 * time.Minute, 120 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			start, end, err := parseTimingLine(tt.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if start != tt.wantStart {
				t.Errorf("start = %v, want %v", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("end = %v, want %v", end, tt.wantEnd)
			}
		})
	}
}

func TestSRTTimingLineParsingErrors(t *testing.T) {
	tests := []string{
		"",
		"00:00:01,000",                 // missing arrow and end
		"00:00:01,000 -> 00:00:02,000", // wrong arrow
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, _, err := parseTimingLine(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestSRTTimingLineNoSpaces(t *testing.T) {
	// No spaces around arrow still works
	start, end, err := parseTimingLine("00:00:01,000-->00:00:02,000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 1*time.Second {
		t.Errorf("start = %v, want 1s", start)
	}
	if end != 2*time.Second {
		t.Errorf("end = %v, want 2s", end)
	}
}
