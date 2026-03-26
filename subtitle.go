package subgo

import "time"

// Event represents a single subtitle cue.
type Event struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

// Subtitle is a collection of events.
type Subtitle struct {
	Events []Event
}

// Shift returns a new Subtitle with all events shifted by delta.
// If clamp is true, negative times are clamped to zero.
func (s Subtitle) Shift(delta time.Duration, clamp bool) Subtitle {
	out := Subtitle{Events: make([]Event, len(s.Events))}
	for i, e := range s.Events {
		start := e.Start + delta
		end := e.End + delta
		if clamp {
			if start < 0 {
				start = 0
			}
			if end < 0 {
				end = 0
			}
		}
		out.Events[i] = Event{
			Start: start,
			End:   end,
			Text:  e.Text,
		}
	}
	return out
}

// Stretch returns a new Subtitle with all times scaled around an anchor.
// newT = (t - anchor) * factor + anchor.
func (s Subtitle) Stretch(factor float64, anchor time.Duration) Subtitle {
	out := Subtitle{Events: make([]Event, len(s.Events))}
	for i, e := range s.Events {
		start := time.Duration(float64(e.Start-anchor)*factor) + anchor
		end := time.Duration(float64(e.End-anchor)*factor) + anchor
		out.Events[i] = Event{
			Start: start,
			End:   end,
			Text:  e.Text,
		}
	}
	return out
}

// TrimFirst removes the first n events.
func (s Subtitle) TrimFirst(n int) Subtitle {
	if n <= 0 {
		return s
	}
	if n >= len(s.Events) {
		return Subtitle{}
	}
	out := Subtitle{Events: make([]Event, len(s.Events)-n)}
	copy(out.Events, s.Events[n:])
	return out
}

// TrimLast removes the last n events.
func (s Subtitle) TrimLast(n int) Subtitle {
	if n <= 0 {
		return s
	}
	if n >= len(s.Events) {
		return Subtitle{}
	}
	out := Subtitle{Events: make([]Event, len(s.Events)-n)}
	copy(out.Events, s.Events[:len(s.Events)-n])
	return out
}

// TrimBefore removes all events that start before the given timestamp.
func (s Subtitle) TrimBefore(t time.Duration) Subtitle {
	var events []Event
	for _, e := range s.Events {
		if e.Start >= t {
			events = append(events, e)
		}
	}
	return Subtitle{Events: events}
}

// TrimAfter removes all events that start after the given timestamp.
func (s Subtitle) TrimAfter(t time.Duration) Subtitle {
	var events []Event
	for _, e := range s.Events {
		if e.Start <= t {
			events = append(events, e)
		}
	}
	return Subtitle{Events: events}
}
