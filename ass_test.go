package subgo

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestASSDecode(t *testing.T) {
	input := `[Script Info]
; This is a comment
ScriptType: v4.00+
PlayResX: 384
PlayResY: 288

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,2,2,2,10,10,10,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:01.00,0:00:02.50,Default,,0,0,0,,Hello World
Dialogue: 0,0:00:03.00,0:00:04.00,Default,,0,0,0,,Second line\NWith newline
Dialogue: 0,1:30:45.50,1:30:50.00,Default,,0,0,0,,Long timestamp
`

	format := assFormat{}
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
		{1, 3 * time.Second, 4 * time.Second, "Second line\nWith newline"},
		{2, 1*time.Hour + 30*time.Minute + 45*time.Second + 500*time.Millisecond, 1*time.Hour + 30*time.Minute + 50*time.Second, "Long timestamp"},
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

func TestASSDecodeEmptyFile(t *testing.T) {
	format := assFormat{}
	sub, err := format.Decode(strings.NewReader(""))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(sub.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(sub.Events))
	}
}

func TestASSDecodeNoEventsSection(t *testing.T) {
	input := `[Script Info]
ScriptType: v4.00+

[V4+ Styles]
Format: Name, Fontname, Fontsize
Style: Default,Arial,20
`

	format := assFormat{}
	sub, err := format.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(sub.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(sub.Events))
	}
}

func TestASSDecodeNewlineVariants(t *testing.T) {
	input := `[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:01.00,0:00:02.00,Default,,0,0,0,,Line1\NLine2
Dialogue: 0,0:00:03.00,0:00:04.00,Default,,0,0,0,,Line1\nLine2
`

	format := assFormat{}
	sub, err := format.Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if len(sub.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(sub.Events))
	}

	// Both \N and \n should be converted to actual newlines
	for i, e := range sub.Events {
		if e.Text != "Line1\nLine2" {
			t.Errorf("event %d: text = %q, want %q", i, e.Text, "Line1\nLine2")
		}
	}
}

func TestASSEncode(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2500 * time.Millisecond, Text: "Hello World"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second line\nWith newline"},
		},
	}

	format := assFormat{}
	var buf bytes.Buffer
	err := format.Encode(&buf, sub)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	output := buf.String()

	// Check sections exist
	if !strings.Contains(output, "[Script Info]") {
		t.Error("missing [Script Info] section")
	}
	if !strings.Contains(output, "[V4+ Styles]") {
		t.Error("missing [V4+ Styles] section")
	}
	if !strings.Contains(output, "[Events]") {
		t.Error("missing [Events] section")
	}

	// Check dialogue lines
	if !strings.Contains(output, "Dialogue: 0,0:00:01.00,0:00:02.50,Default,,0,0,0,,Hello World") {
		t.Errorf("first dialogue not encoded correctly:\n%s", output)
	}
	if !strings.Contains(output, "Dialogue: 0,0:00:03.00,0:00:04.00,Default,,0,0,0,,Second line\\NWith newline") {
		t.Errorf("second dialogue not encoded correctly (newline should be \\N):\n%s", output)
	}
}

func TestASSEncodeEmpty(t *testing.T) {
	sub := Subtitle{}

	format := assFormat{}
	var buf bytes.Buffer
	err := format.Encode(&buf, sub)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	output := buf.String()
	// Should still have headers even with no events
	if !strings.Contains(output, "[Script Info]") {
		t.Error("missing [Script Info] section")
	}
	if !strings.Contains(output, "[Events]") {
		t.Error("missing [Events] section")
	}
}

func TestASSRoundTrip(t *testing.T) {
	original := Subtitle{
		Events: []Event{
			{Start: 0, End: 1 * time.Second, Text: "First"},
			{Start: 1*time.Hour + 30*time.Minute + 45*time.Second + 120*time.Millisecond, End: 1*time.Hour + 30*time.Minute + 50*time.Second, Text: "Long timestamp"},
			{Start: 5 * time.Second, End: 10 * time.Second, Text: "Multi\nLine\nText"},
		},
	}

	format := assFormat{}

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
		// ASS uses centiseconds, so we lose some precision
		// Allow 10ms tolerance
		if abs(dec.Start-orig.Start) > 10*time.Millisecond {
			t.Errorf("event %d: start = %v, want %v", i, dec.Start, orig.Start)
		}
		if abs(dec.End-orig.End) > 10*time.Millisecond {
			t.Errorf("event %d: end = %v, want %v", i, dec.End, orig.End)
		}
		if dec.Text != orig.Text {
			t.Errorf("event %d: text = %q, want %q", i, dec.Text, orig.Text)
		}
	}
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func TestASSTimestampParsing(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"0:00:00.00", 0},
		{"0:00:01.00", 1 * time.Second},
		{"0:00:00.01", 10 * time.Millisecond},
		{"0:00:00.99", 990 * time.Millisecond},
		{"0:01:00.00", 1 * time.Minute},
		{"1:00:00.00", 1 * time.Hour},
		{"1:30:45.50", 1*time.Hour + 30*time.Minute + 45*time.Second + 500*time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseASSTimestamp(tt.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestASSTimestampParsingErrors(t *testing.T) {
	tests := []string{
		"",
		"0:00:00",    // missing centiseconds
		"0:00:00,00", // wrong separator
		"invalid",
		"00:00.00", // missing hours
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseASSTimestamp(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestASSTimestampFormatting(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{0, "0:00:00.00"},
		{1 * time.Second, "0:00:01.00"},
		{10 * time.Millisecond, "0:00:00.01"},
		{990 * time.Millisecond, "0:00:00.99"},
		{1 * time.Minute, "0:01:00.00"},
		{1 * time.Hour, "1:00:00.00"},
		{1*time.Hour + 30*time.Minute + 45*time.Second + 500*time.Millisecond, "1:30:45.50"},
		{-1 * time.Second, "0:00:00.00"}, // negative clamped to zero
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatASSTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestASSDialogueParsing(t *testing.T) {
	formatFields := []string{"layer", "start", "end", "style", "name", "marginl", "marginr", "marginv", "effect", "text"}

	tests := []struct {
		name      string
		input     string
		wantStart time.Duration
		wantEnd   time.Duration
		wantText  string
	}{
		{
			name:      "simple",
			input:     "0,0:00:01.00,0:00:02.00,Default,,0,0,0,,Hello",
			wantStart: 1 * time.Second,
			wantEnd:   2 * time.Second,
			wantText:  "Hello",
		},
		{
			name:      "with commas in text",
			input:     "0,0:00:01.00,0:00:02.00,Default,,0,0,0,,Hello, World, Test",
			wantStart: 1 * time.Second,
			wantEnd:   2 * time.Second,
			wantText:  "Hello, World, Test",
		},
		{
			name:      "with newlines",
			input:     "0,0:00:01.00,0:00:02.00,Default,,0,0,0,,Line1\\NLine2",
			wantStart: 1 * time.Second,
			wantEnd:   2 * time.Second,
			wantText:  "Line1\nLine2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parseASSDialogue(tt.input, formatFields)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if event.Start != tt.wantStart {
				t.Errorf("start = %v, want %v", event.Start, tt.wantStart)
			}
			if event.End != tt.wantEnd {
				t.Errorf("end = %v, want %v", event.End, tt.wantEnd)
			}
			if event.Text != tt.wantText {
				t.Errorf("text = %q, want %q", event.Text, tt.wantText)
			}
		})
	}
}

func TestASSFieldsParsing(t *testing.T) {
	input := "Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text"
	fields := parseASSFields(input)

	expected := []string{"layer", "start", "end", "style", "name", "marginl", "marginr", "marginv", "effect", "text"}

	if len(fields) != len(expected) {
		t.Fatalf("got %d fields, want %d", len(fields), len(expected))
	}

	for i, f := range fields {
		if f != expected[i] {
			t.Errorf("field %d = %q, want %q", i, f, expected[i])
		}
	}
}
