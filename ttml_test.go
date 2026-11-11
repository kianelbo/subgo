package subgo

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestTTMLDecode(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<tt xmlns="http://www.w3.org/ns/ttml"><body><div>
  <p begin="00:00:01.250" end="00:00:02.500">Hello <span>world</span> &amp; all<br/>again</p>
  <p begin="2s" dur="1500ms">Second cue</p>
</div></body></tt>`

	got, err := (ttmlFormat{}).Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	want := Subtitle{Events: []Event{
		{Start: 1250 * time.Millisecond, End: 2500 * time.Millisecond, Text: "Hello world & all\nagain"},
		{Start: 2 * time.Second, End: 3500 * time.Millisecond, Text: "Second cue"},
	}}
	if len(got.Events) != len(want.Events) {
		t.Fatalf("event count = %d, want %d", len(got.Events), len(want.Events))
	}
	for i := range want.Events {
		if got.Events[i] != want.Events[i] {
			t.Errorf("event %d = %+v, want %+v", i, got.Events[i], want.Events[i])
		}
	}
}

func TestTTMLDecodeSkipsMalformedCue(t *testing.T) {
	input := `<tt><body><div>
<p begin="invalid" end="00:00:02.000">ignored</p>
<p begin="00:00:01.000">missing end and duration</p>
<p begin="00:00:01.000" dur="invalid">invalid duration</p>
<p begin="00:00:02.000" end="00:00:03.000">kept</p>
</div></body></tt>`
	got, err := (ttmlFormat{}).Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if len(got.Events) != 1 || got.Events[0].Text != "kept" {
		t.Fatalf("events = %+v, want one valid event", got.Events)
	}
}

func TestTTMLEncodeAndRoundTrip(t *testing.T) {
	original := Subtitle{Events: []Event{
		{Start: 1250 * time.Millisecond, End: 2500 * time.Millisecond, Text: "Hello <world> & everyone"},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second\nline"},
	}}
	var output bytes.Buffer
	if err := (ttmlFormat{}).Encode(&output, original); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if !strings.Contains(output.String(), `xmlns="`+ttmlNamespace+`"`) ||
		!strings.Contains(output.String(), "Hello &lt;world&gt; &amp; everyone") {
		t.Errorf("encoded output missing TTML namespace or escaped text:\n%s", output.String())
	}

	got, err := (ttmlFormat{}).Decode(&output)
	if err != nil {
		t.Fatalf("Decode encoded output error: %v", err)
	}
	if len(got.Events) != len(original.Events) {
		t.Fatalf("event count = %d, want %d", len(got.Events), len(original.Events))
	}
	for i := range original.Events {
		if got.Events[i] != original.Events[i] {
			t.Errorf("event %d = %+v, want %+v", i, got.Events[i], original.Events[i])
		}
	}
}

func TestTTMLEncodeEmpty(t *testing.T) {
	var output bytes.Buffer
	if err := (ttmlFormat{}).Encode(&output, Subtitle{}); err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	got, err := (ttmlFormat{}).Decode(&output)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if len(got.Events) != 0 {
		t.Errorf("decoded %d events from empty subtitle, want none", len(got.Events))
	}
}

func TestTTMLDecodeInvalidXML(t *testing.T) {
	if _, err := (ttmlFormat{}).Decode(strings.NewReader(`<tt><p>`)); err == nil {
		t.Error("expected an error for malformed XML")
	}
	if _, err := (ttmlFormat{}).Decode(strings.NewReader(`<tt><p begin="1s" end="2s">&invalid;</p></tt>`)); err == nil {
		t.Error("expected an error for malformed XML inside a cue")
	}
	if _, err := (ttmlFormat{}).Decode(strings.NewReader(`<tt><div></tt>`)); err == nil {
		t.Error("expected an error for malformed XML outside a cue")
	}
}

func TestTTMLTimeParsing(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"00:00:01.25", 1250 * time.Millisecond},
		{"01:02:03.004", time.Hour + 2*time.Minute + 3004*time.Millisecond},
		{"1.5s", 1500 * time.Millisecond},
		{"250ms", 250 * time.Millisecond},
		{"1m", time.Minute},
		{"1h", time.Hour},
		{"00:00:00.123456789", 123456789 * time.Nanosecond},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTTMLTime(tt.input)
			if err != nil {
				t.Fatalf("parseTTMLTime error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseTTMLTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTTMLTimeParsingErrors(t *testing.T) {
	tests := []string{
		"",
		"invalid",
		"invalids",
		"xx:00:01",
		"00:xx:01",
		"00:00:xx",
		"00:60:00",
		"00:00:60",
		"00:00:01.",
		"00:00:01.1234567890",
		"00:00:01.abc",
		"00:00:01.2.3",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := parseTTMLTime(input); err == nil {
				t.Errorf("parseTTMLTime(%q) succeeded, want error", input)
			}
		})
	}
}
