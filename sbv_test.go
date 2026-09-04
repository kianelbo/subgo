package subgo

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSBVDecode(t *testing.T) {
	input := `0:00:01.000,0:00:02.500
Hello World

1:03:04.250,1:03:05.000
Second line
With multiple lines
`

	sub, err := (sbvFormat{}).Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	want := Subtitle{Events: []Event{
		{Start: time.Second, End: 2500 * time.Millisecond, Text: "Hello World"},
		{Start: time.Hour + 3*time.Minute + 4*time.Second + 250*time.Millisecond, End: time.Hour + 3*time.Minute + 5*time.Second, Text: "Second line\nWith multiple lines"},
	}}
	if len(sub.Events) != len(want.Events) {
		t.Fatalf("event count = %d, want %d", len(sub.Events), len(want.Events))
	}
	for i := range want.Events {
		if sub.Events[i] != want.Events[i] {
			t.Errorf("event %d = %+v, want %+v", i, sub.Events[i], want.Events[i])
		}
	}
}

func TestSBVDecodeEmptyFile(t *testing.T) {
	sub, err := (sbvFormat{}).Decode(strings.NewReader(""))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(sub.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(sub.Events))
	}
}

func TestSBVDecodeWithExtraBlankLines(t *testing.T) {
	input := `

0:00:01.000,0:00:02.000
First


0:00:03.000,0:00:04.000
Second

`

	sub, err := (sbvFormat{}).Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(sub.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(sub.Events))
	}
}

func TestSBVDecodeTimingWhitespace(t *testing.T) {
	input := " 0:00:01.000 , 0:00:02.000 \nText\n"
	sub, err := (sbvFormat{}).Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(sub.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(sub.Events))
	}
	if sub.Events[0].Start != time.Second || sub.Events[0].End != 2*time.Second {
		t.Errorf("timestamps = %v-%v, want 1s-2s", sub.Events[0].Start, sub.Events[0].End)
	}
}

func TestSBVDecodeMalformedTiming(t *testing.T) {
	tests := []string{
		"",
		"0:00:01.000",
		"0:00:01.000,0:00:02",
		"0:00:01,000,0:00:02.000",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := (sbvFormat{}).Decode(strings.NewReader(input + "\nText\n"))
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestSBVEncode(t *testing.T) {
	sub := Subtitle{Events: []Event{
		{Start: time.Second, End: 2500 * time.Millisecond, Text: "Hello World"},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second\nLine"},
	}}
	var buf bytes.Buffer
	if err := (sbvFormat{}).Encode(&buf, sub); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	want := "0:00:01.000,0:00:02.500\nHello World\n\n0:00:03.000,0:00:04.000\nSecond\nLine\n\n"
	if buf.String() != want {
		t.Errorf("encoded output = %q, want %q", buf.String(), want)
	}
}

func TestSBVEncodeEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := (sbvFormat{}).Encode(&buf, Subtitle{}); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestSBVRoundTrip(t *testing.T) {
	original := Subtitle{Events: []Event{
		{Start: 1*time.Hour + 30*time.Minute + 45*time.Second + 123*time.Millisecond, End: 1*time.Hour + 30*time.Minute + 50*time.Second, Text: "Long timestamp"},
	}}
	var buf bytes.Buffer
	format := sbvFormat{}
	if err := format.Encode(&buf, original); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := format.Decode(&buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(decoded.Events) != 1 || decoded.Events[0] != original.Events[0] {
		t.Errorf("decoded = %+v, want %+v", decoded.Events, original.Events)
	}
}

func TestSBVTimestampParsing(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"0:00:00.000", 0},
		{"0:00:01.000", time.Second},
		{"0:00:00.001", time.Millisecond},
		{"0:01:00.000", time.Minute},
		{"1:00:00.000", time.Hour},
		{"12:34:56.789", 12*time.Hour + 34*time.Minute + 56*time.Second + 789*time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseSBVTimestamp(tt.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSBVTimestampParsingErrors(t *testing.T) {
	tests := []string{
		"",
		"0:00:00",
		"0:00:00,000",
		"invalid",
		"0:00.000",
		"0:00:00.000.000",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseSBVTimestamp(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestSBVTimingLineParsing(t *testing.T) {
	start, end, err := parseSBVTimingLine("0:00:01.000,0:00:02.000")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if start != time.Second || end != 2*time.Second {
		t.Errorf("timestamps = %v-%v, want 1s-2s", start, end)
	}
}

func TestSBVTimingLineParsingErrors(t *testing.T) {
	tests := []string{
		"",
		"0:00:01.000",
		"0:00:01.000,0:00:02.000,0:00:03.000",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, _, err := parseSBVTimingLine(input)
			if err == nil {
				t.Errorf("expected error for input %q", input)
			}
		})
	}
}

func TestSBVTimestampFormatting(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{0, "0:00:00.000"},
		{time.Second, "0:00:01.000"},
		{time.Millisecond, "0:00:00.001"},
		{time.Minute, "0:01:00.000"},
		{time.Hour, "1:00:00.000"},
		{12*time.Hour + 34*time.Minute + 56*time.Second + 789*time.Millisecond, "12:34:56.789"},
		{-time.Second, "0:00:00.000"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatSBVTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
