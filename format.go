package subgo

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Format represents a subtitle file format (SRT, ASS, etc.).
type Format interface {
	Name() string
	Extensions() []string
	Decode(r io.Reader) (Subtitle, error)
	Encode(w io.Writer, s Subtitle) error
}

var formats = make(map[string]Format)

func init() {
	for _, f := range []Format{assFormat{}, srtFormat{}, vttFormat{}} {
		for _, ext := range f.Extensions() {
			formats[normalizeExt(ext)] = f
		}
	}
}

// DetectFormat returns a Format based on the filename's extension.
func DetectFormat(filename string) (Format, error) {
	ext := normalizeExt(filepath.Ext(filename))
	format, ok := formats[ext]
	if !ok {
		return nil, errors.New("unknown format: " + ext)
	}
	return format, nil
}

// Load reads a subtitle file, detecting format from the filename extension.
func Load(filename string) (Subtitle, error) {
	format, err := DetectFormat(filename)
	if err != nil {
		return Subtitle{}, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return Subtitle{}, err
	}
	defer file.Close()

	return format.Decode(file)
}

// Save writes the subtitle to a file, detecting format from the filename extension.
func (s Subtitle) Save(filename string) error {
	format, err := DetectFormat(filename)
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
