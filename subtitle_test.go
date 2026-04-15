package subgo

import (
	"testing"
	"time"
)

func TestShift(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "First"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second"},
		},
	}

	tests := []struct {
		name       string
		delta      time.Duration
		clamp      bool
		wantStarts []time.Duration
		wantEnds   []time.Duration
	}{
		{
			name:       "positive shift",
			delta:      500 * time.Millisecond,
			clamp:      true,
			wantStarts: []time.Duration{1500 * time.Millisecond, 3500 * time.Millisecond},
			wantEnds:   []time.Duration{2500 * time.Millisecond, 4500 * time.Millisecond},
		},
		{
			name:       "negative shift with clamp",
			delta:      -2 * time.Second,
			clamp:      true,
			wantStarts: []time.Duration{0, 1 * time.Second},
			wantEnds:   []time.Duration{0, 2 * time.Second},
		},
		{
			name:       "negative shift without clamp",
			delta:      -2 * time.Second,
			clamp:      false,
			wantStarts: []time.Duration{-1 * time.Second, 1 * time.Second},
			wantEnds:   []time.Duration{0, 2 * time.Second},
		},
		{
			name:       "zero shift",
			delta:      0,
			clamp:      true,
			wantStarts: []time.Duration{1 * time.Second, 3 * time.Second},
			wantEnds:   []time.Duration{2 * time.Second, 4 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sub.Shift(tt.delta, tt.clamp)

			if len(result.Events) != len(sub.Events) {
				t.Fatalf("got %d events, want %d", len(result.Events), len(sub.Events))
			}

			for i, e := range result.Events {
				if e.Start != tt.wantStarts[i] {
					t.Errorf("event %d: got start %v, want %v", i, e.Start, tt.wantStarts[i])
				}
				if e.End != tt.wantEnds[i] {
					t.Errorf("event %d: got end %v, want %v", i, e.End, tt.wantEnds[i])
				}
			}
		})
	}
}

func TestShiftPreservesText(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "Hello\nWorld"},
		},
	}

	result := sub.Shift(1*time.Second, true)

	if result.Events[0].Text != "Hello\nWorld" {
		t.Errorf("text not preserved: got %q", result.Events[0].Text)
	}
}

func TestStretch(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 2 * time.Second, End: 4 * time.Second, Text: "First"},
			{Start: 6 * time.Second, End: 8 * time.Second, Text: "Second"},
		},
	}

	tests := []struct {
		name       string
		factor     float64
		anchor     time.Duration
		wantStarts []time.Duration
		wantEnds   []time.Duration
	}{
		{
			name:       "stretch 2x from zero",
			factor:     2.0,
			anchor:     0,
			wantStarts: []time.Duration{4 * time.Second, 12 * time.Second},
			wantEnds:   []time.Duration{8 * time.Second, 16 * time.Second},
		},
		{
			name:       "compress 0.5x from zero",
			factor:     0.5,
			anchor:     0,
			wantStarts: []time.Duration{1 * time.Second, 3 * time.Second},
			wantEnds:   []time.Duration{2 * time.Second, 4 * time.Second},
		},
		{
			name:       "stretch 2x from anchor",
			factor:     2.0,
			anchor:     2 * time.Second,
			wantStarts: []time.Duration{2 * time.Second, 10 * time.Second},
			wantEnds:   []time.Duration{6 * time.Second, 14 * time.Second},
		},
		{
			name:       "no stretch",
			factor:     1.0,
			anchor:     0,
			wantStarts: []time.Duration{2 * time.Second, 6 * time.Second},
			wantEnds:   []time.Duration{4 * time.Second, 8 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sub.Stretch(tt.factor, tt.anchor)

			if len(result.Events) != len(sub.Events) {
				t.Fatalf("got %d events, want %d", len(result.Events), len(sub.Events))
			}

			for i, e := range result.Events {
				if e.Start != tt.wantStarts[i] {
					t.Errorf("event %d: got start %v, want %v", i, e.Start, tt.wantStarts[i])
				}
				if e.End != tt.wantEnds[i] {
					t.Errorf("event %d: got end %v, want %v", i, e.End, tt.wantEnds[i])
				}
			}
		})
	}
}

func TestTrimFirst(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "First"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second"},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: "Third"},
		},
	}

	tests := []struct {
		name      string
		n         int
		wantCount int
		wantFirst string
	}{
		{"trim 0", 0, 3, "First"},
		{"trim 1", 1, 2, "Second"},
		{"trim 2", 2, 1, "Third"},
		{"trim all", 3, 0, ""},
		{"trim more than all", 5, 0, ""},
		{"trim negative", -1, 3, "First"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sub.TrimFirst(tt.n)

			if len(result.Events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(result.Events), tt.wantCount)
			}

			if tt.wantCount > 0 && result.Events[0].Text != tt.wantFirst {
				t.Errorf("first event text = %q, want %q", result.Events[0].Text, tt.wantFirst)
			}
		})
	}
}

func TestTrimLast(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "First"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second"},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: "Third"},
		},
	}

	tests := []struct {
		name      string
		n         int
		wantCount int
		wantLast  string
	}{
		{"trim 0", 0, 3, "Third"},
		{"trim 1", 1, 2, "Second"},
		{"trim 2", 2, 1, "First"},
		{"trim all", 3, 0, ""},
		{"trim more than all", 5, 0, ""},
		{"trim negative", -1, 3, "Third"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sub.TrimLast(tt.n)

			if len(result.Events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(result.Events), tt.wantCount)
			}

			if tt.wantCount > 0 && result.Events[len(result.Events)-1].Text != tt.wantLast {
				t.Errorf("last event text = %q, want %q", result.Events[len(result.Events)-1].Text, tt.wantLast)
			}
		})
	}
}

func TestTrimBefore(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "First"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second"},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: "Third"},
		},
	}

	tests := []struct {
		name      string
		t         time.Duration
		wantCount int
		wantFirst string
	}{
		{"before 0", 0, 3, "First"},
		{"before 1s", 1 * time.Second, 3, "First"},
		{"before 2s", 2 * time.Second, 2, "Second"},
		{"before 3s", 3 * time.Second, 2, "Second"},
		{"before 4s", 4 * time.Second, 1, "Third"},
		{"before 6s", 6 * time.Second, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sub.TrimBefore(tt.t)

			if len(result.Events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(result.Events), tt.wantCount)
			}

			if tt.wantCount > 0 && result.Events[0].Text != tt.wantFirst {
				t.Errorf("first event text = %q, want %q", result.Events[0].Text, tt.wantFirst)
			}
		})
	}
}

func TestTrimAfter(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "First"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second"},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: "Third"},
		},
	}

	tests := []struct {
		name      string
		t         time.Duration
		wantCount int
		wantLast  string
	}{
		{"after 6s", 6 * time.Second, 3, "Third"},
		{"after 5s", 5 * time.Second, 3, "Third"},
		{"after 4s", 4 * time.Second, 2, "Second"},
		{"after 3s", 3 * time.Second, 2, "Second"},
		{"after 2s", 2 * time.Second, 1, "First"},
		{"after 0s", 0, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sub.TrimAfter(tt.t)

			if len(result.Events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(result.Events), tt.wantCount)
			}

			if tt.wantCount > 0 && result.Events[len(result.Events)-1].Text != tt.wantLast {
				t.Errorf("last event text = %q, want %q", result.Events[len(result.Events)-1].Text, tt.wantLast)
			}
		})
	}
}

func TestEmptySubtitle(t *testing.T) {
	sub := Subtitle{}

	// All operations should handle empty subtitles gracefully
	t.Run("shift empty", func(t *testing.T) {
		result := sub.Shift(1*time.Second, true)
		if len(result.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(result.Events))
		}
	})

	t.Run("stretch empty", func(t *testing.T) {
		result := sub.Stretch(2.0, 0)
		if len(result.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(result.Events))
		}
	})

	t.Run("trim first empty", func(t *testing.T) {
		result := sub.TrimFirst(1)
		if len(result.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(result.Events))
		}
	})

	t.Run("trim last empty", func(t *testing.T) {
		result := sub.TrimLast(1)
		if len(result.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(result.Events))
		}
	})

	t.Run("trim before empty", func(t *testing.T) {
		result := sub.TrimBefore(1 * time.Second)
		if len(result.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(result.Events))
		}
	})

	t.Run("trim after empty", func(t *testing.T) {
		result := sub.TrimAfter(1 * time.Second)
		if len(result.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(result.Events))
		}
	})
}

func TestChainedOperations(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "First"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "Second"},
			{Start: 5 * time.Second, End: 6 * time.Second, Text: "Third"},
			{Start: 7 * time.Second, End: 8 * time.Second, Text: "Fourth"},
		},
	}

	// Trim first, then shift, then trim after
	result := sub.
		TrimFirst(1).
		Shift(-2*time.Second, true).
		TrimAfter(3 * time.Second)

	if len(result.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result.Events))
	}

	// After TrimFirst(1): Second(3s), Third(5s), Fourth(7s)
	// After Shift(-2s): Second(1s), Third(3s), Fourth(5s)
	// After TrimAfter(3s): Second(1s), Third(3s)

	if result.Events[0].Text != "Second" {
		t.Errorf("first event = %q, want Second", result.Events[0].Text)
	}
	if result.Events[0].Start != 1*time.Second {
		t.Errorf("first event start = %v, want 1s", result.Events[0].Start)
	}
	if result.Events[1].Text != "Third" {
		t.Errorf("second event = %q, want Third", result.Events[1].Text)
	}
	if result.Events[1].Start != 3*time.Second {
		t.Errorf("second event start = %v, want 3s", result.Events[1].Start)
	}
}

func TestImmutability(t *testing.T) {
	original := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "First [waving]"},
		},
	}

	// Operations should not modify the original
	_ = original.Shift(5*time.Second, true)
	_ = original.Stretch(2.0, 0)
	_ = original.TrimFirst(1)
	_ = original.RemoveHI()

	if original.Events[0].Start != 1*time.Second || original.Events[0].End != 2*time.Second {
		t.Errorf("original was modified: (start = %v, end = %v)", original.Events[0].Start, original.Events[0].End)
	}
	if len(original.Events) != 1 {
		t.Errorf("original was modified: (len = %d)", len(original.Events))
	}
	if original.Events[0].Text != "First [waving]" {
		t.Errorf("original text was modified: got %q", original.Events[0].Text)
	}
}

func TestRemoveHI(t *testing.T) {
	tests := []struct {
		name      string
		input     Subtitle
		wantCount int
		wantTexts []string
	}{
		{
			name: "remove parentheses HI",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "(sobbing) I can't believe it"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"I can't believe it"},
		},
		{
			name: "remove brackets HI",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "[loud noise] What was that?"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"What was that?"},
		},
		{
			name: "remove hash HI",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "#gulps# What?!"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"What?!"},
		},
		{
			name: "remove HI at end",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "Hello there (laughing)"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"Hello there"},
		},
		{
			name: "remove HI in middle",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "Hello (sighs) there"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"Hello there"},
		},
		{
			name: "remove multiple HI annotations",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "#sobbing# I can't [door slams] believe it (crying)"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"I can't believe it"},
		},
		{
			name: "remove event when text becomes empty",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "(sobbing)"},
					{Start: 3 * time.Second, End: 4 * time.Second, Text: "Hello"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"Hello"},
		},
		{
			name: "remove event with only brackets",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "[music playing]"},
				},
			},
			wantCount: 0,
			wantTexts: []string{},
		},
		{
			name: "preserve text without HI",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "Hello, how are you?"},
				},
			},
			wantCount: 1,
			wantTexts: []string{"Hello, how are you?"},
		},
		{
			name: "empty subtitle",
			input: Subtitle{
				Events: []Event{},
			},
			wantCount: 0,
			wantTexts: []string{},
		},
		{
			name: "multiple events mixed",
			input: Subtitle{
				Events: []Event{
					{Start: 1 * time.Second, End: 2 * time.Second, Text: "(music)"},
					{Start: 3 * time.Second, End: 4 * time.Second, Text: "Hello (laughing) world"},
					{Start: 5 * time.Second, End: 6 * time.Second, Text: "[silence]"},
					{Start: 7 * time.Second, End: 8 * time.Second, Text: "Goodbye"},
				},
			},
			wantCount: 2,
			wantTexts: []string{"Hello world", "Goodbye"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.RemoveHI()

			if len(result.Events) != tt.wantCount {
				t.Errorf("got %d events, want %d", len(result.Events), tt.wantCount)
			}

			for i, wantText := range tt.wantTexts {
				if i < len(result.Events) && result.Events[i].Text != wantText {
					t.Errorf("event %d: got text %q, want %q", i, result.Events[i].Text, wantText)
				}
			}
		})
	}
}

func TestRemoveHIPreservesTimestamps(t *testing.T) {
	sub := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "(sobbing) Hello"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "World"},
		},
	}

	result := sub.RemoveHI()

	if len(result.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result.Events))
	}

	if result.Events[0].Start != 1*time.Second || result.Events[0].End != 2*time.Second {
		t.Errorf("first event timestamps changed: got %v-%v", result.Events[0].Start, result.Events[0].End)
	}
	if result.Events[1].Start != 3*time.Second || result.Events[1].End != 4*time.Second {
		t.Errorf("second event timestamps changed: got %v-%v", result.Events[1].Start, result.Events[1].End)
	}
}
