package formats

import (
	"errors"
	"io"
	"os"
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

// Load reads a subtitle file, detecting format from the filename extension.
func Load(filename string) (subtitle.Subtitle, error) {
	format, err := DetectFromFilename(filename)
	if err != nil {
		return subtitle.Subtitle{}, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return subtitle.Subtitle{}, err
	}
	defer file.Close()

	return format.Decode(file)
}

// Save writes a subtitle to a file, detecting format from the filename extension.
func Save(filename string, s subtitle.Subtitle) error {
	format, err := DetectFromFilename(filename)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return format.Encode(file, s)
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
