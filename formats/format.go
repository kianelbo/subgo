package formats

import (
	"errors"
	"io"
	"path/filepath"
	"strings"

	"subgo/subtitle"
)

// Format represents a subtitle file format (SRT, ASS, etc.).
type Format interface {
	Name() string
	Extensions() []string
	Decode(r io.Reader) (subtitle.Subtitle, error)
	Encode(w io.Writer, s subtitle.Subtitle) error
}

// DetectFromFilename tries to choose a Format based on the filename's extension.
func DetectFromFilename(name string) (Format, error) {
	m := map[string]Format{
		"srt": srtFormat{},
		"ass": assFormat{},
		"ssa": assFormat{},
	}

	ext := normalizeExt(filepath.Ext(name))
	format, ok := m[ext]
	if !ok {
		return nil, errors.New("unknown format")
	}
	return format, nil
}

// EncodeTo writes the subtitle in the specified format.
func EncodeTo(w io.Writer, s subtitle.Subtitle, f Format) error {
	if f == nil {
		return errors.New("format is nil")
	}
	return f.Encode(w, s)
}

// DecodeFrom reads a subtitle using the specified format.
func DecodeFrom(r io.Reader, f Format) (subtitle.Subtitle, error) {
	if f == nil {
		return subtitle.Subtitle{}, errors.New("format is nil")
	}
	return f.Decode(r)
}

func normalizeExt(ext string) string {
	if ext == "" {
		return ""
	}
	// Remove leading dot if present
	if ext[0] == '.' {
		ext = ext[1:]
	}
	return strings.ToLower(ext)
}
