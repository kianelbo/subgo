package subgo

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const ttmlNamespace = "http://www.w3.org/ns/ttml"

// ttmlFormat implements the Timed Text Markup Language subtitle format.
type ttmlFormat struct{}

func (ttmlFormat) Name() string { return "ttml" }

func (ttmlFormat) Extensions() []string { return []string{".ttml", ".xml"} }

func (ttmlFormat) Decode(r io.Reader) (Subtitle, error) {
	decoder := xml.NewDecoder(r)
	var events []Event

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Subtitle{}, err
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "p" {
			continue
		}

		begin, end, dur := "", "", ""
		for _, attr := range start.Attr {
			switch attr.Name.Local {
			case "begin":
				begin = attr.Value
			case "end":
				end = attr.Value
			case "dur":
				dur = attr.Value
			}
		}

		text, err := readTTMLText(decoder)
		if err != nil {
			return Subtitle{}, err
		}

		startTime, startErr := parseTTMLTime(begin)
		endTime, endErr := parseTTMLTime(end)
		if end == "" && dur != "" {
			duration, err := parseTTMLTime(dur)
			if err == nil {
				endTime = startTime + duration
				endErr = nil
			} else {
				endErr = err
			}
		}
		if startErr != nil || endErr != nil {
			continue
		}
		events = append(events, Event{Start: startTime, End: endTime, Text: text})
	}

	return Subtitle{Events: events}, nil
}

func readTTMLText(decoder *xml.Decoder) (string, error) {
	var text strings.Builder
	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("unexpected end of TTML paragraph")
			}
			return "", err
		}

		switch token := token.(type) {
		case xml.StartElement:
			depth++
			if token.Name.Local == "br" {
				text.WriteByte('\n')
			}
		case xml.EndElement:
			depth--
		case xml.CharData:
			text.Write([]byte(token))
		}
	}
	return text.String(), nil
}

func (ttmlFormat) Encode(w io.Writer, s Subtitle) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}

	encoder := xml.NewEncoder(w)
	root := xml.StartElement{
		Name: xml.Name{Local: "tt"},
		Attr: []xml.Attr{{Name: xml.Name{Local: "xmlns"}, Value: ttmlNamespace}},
	}
	body := xml.StartElement{Name: xml.Name{Local: "body"}}
	div := xml.StartElement{Name: xml.Name{Local: "div"}}
	if err := encoder.EncodeToken(root); err != nil {
		return err
	}
	if err := encoder.EncodeToken(body); err != nil {
		return err
	}
	if err := encoder.EncodeToken(div); err != nil {
		return err
	}
	for _, event := range s.Events {
		paragraph := xml.StartElement{
			Name: xml.Name{Local: "p"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "begin"}, Value: formatVTTTimestamp(event.Start)},
				{Name: xml.Name{Local: "end"}, Value: formatVTTTimestamp(event.End)},
			},
		}
		if err := encoder.EncodeToken(paragraph); err != nil {
			return err
		}
		for i, line := range strings.Split(event.Text, "\n") {
			if i > 0 {
				br := xml.StartElement{Name: xml.Name{Local: "br"}}
				if err := encoder.EncodeToken(br); err != nil {
					return err
				}
				if err := encoder.EncodeToken(br.End()); err != nil {
					return err
				}
			}
			if err := encoder.EncodeToken(xml.CharData(line)); err != nil {
				return err
			}
		}
		if err := encoder.EncodeToken(paragraph.End()); err != nil {
			return err
		}
	}
	if err := encoder.EncodeToken(div.End()); err != nil {
		return err
	}
	if err := encoder.EncodeToken(body.End()); err != nil {
		return err
	}
	if err := encoder.EncodeToken(root.End()); err != nil {
		return err
	}
	return encoder.Flush()
}

func parseTTMLTime(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty TTML time")
	}

	// TTML also permits offset times such as "1.5s" and "250ms".
	for _, unit := range []struct {
		suffix string
		scale  time.Duration
	}{
		{"ms", time.Millisecond},
		{"h", time.Hour},
		{"m", time.Minute},
		{"s", time.Second},
	} {
		if strings.HasSuffix(value, unit.suffix) {
			n, err := strconv.ParseFloat(strings.TrimSuffix(value, unit.suffix), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid TTML time %q: %w", value, err)
			}
			return time.Duration(n * float64(unit.scale)), nil
		}
	}

	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid TTML clock time %q", value)
	}
	hours, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid TTML clock time %q: %w", value, err)
	}
	minutes, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid TTML clock time %q: %w", value, err)
	}

	secondParts := strings.Split(parts[2], ".")
	if len(secondParts) > 2 {
		return 0, fmt.Errorf("invalid TTML clock time %q", value)
	}
	seconds, err := strconv.ParseInt(secondParts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid TTML clock time %q: %w", value, err)
	}
	var fraction time.Duration
	if len(secondParts) == 2 {
		digits := secondParts[1]
		if len(digits) == 0 || len(digits) > 9 {
			return 0, fmt.Errorf("invalid TTML clock time %q", value)
		}
		nanos, err := strconv.ParseInt(digits+strings.Repeat("0", 9-len(digits)), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid TTML clock time %q: %w", value, err)
		}
		fraction = time.Duration(nanos)
	}
	if hours < 0 || minutes < 0 || minutes >= 60 || seconds < 0 || seconds >= 60 {
		return 0, fmt.Errorf("invalid TTML clock time %q", value)
	}
	return time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		fraction, nil
}
