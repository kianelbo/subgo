package subgo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectFormatCaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"test.srt", "srt"},
		{"test.ASS", "ass"},
		{"test.sRt", "srt"},
		{"test.ssA", "ass"},
		{"dir/test.SRT", "srt"},
		{"../testsrt.SRt", "srt"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			f, err := DetectFormat(tt.input)
			if err != nil {
				t.Fatalf("DetectFormat error: %v", err)
			}
			if f.Name() != tt.want {
				t.Errorf("got format %q, want %q", f.Name(), tt.want)
			}
		})
	}
}

func TestDetectFormatUnknown(t *testing.T) {
	_, err := DetectFormat("test.unknown")
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestDetectFormatNoExtension(t *testing.T) {
	_, err := DetectFormat("testfile")
	if err == nil {
		t.Error("expected error for file without extension")
	}
}

func TestNormalizeExt(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{".srt", "srt"},
		{"srt", "srt"},
		{".SRT", "srt"},
		{"SRT", "srt"},
		{".Srt", "srt"},
		{"", ""},
		{".ASS", "ass"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeExt(tt.input)
			if got != tt.want {
				t.Errorf("normalizeExt(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "subgo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test subtitle
	original := Subtitle{
		Events: []Event{
			{Start: 1 * time.Second, End: 2 * time.Second, Text: "Hello"},
			{Start: 3 * time.Second, End: 4 * time.Second, Text: "World"},
		},
	}

	// Save it
	filename := filepath.Join(tmpDir, "test.srt")
	if err := original.Save(filename); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Load it back
	loaded, err := Load(filename)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	// Compare
	if len(loaded.Events) != len(original.Events) {
		t.Fatalf("event count mismatch: got %d, want %d", len(loaded.Events), len(original.Events))
	}

	for i, orig := range original.Events {
		load := loaded.Events[i]
		if load.Start != orig.Start || load.End != orig.End || load.Text != orig.Text {
			t.Errorf("event %d mismatch: got %+v, want %+v", i, load, orig)
		}
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	if _, err := Load("/nonexistent/path/file.srt"); err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadUnknownFormat(t *testing.T) {
	if _, err := Load("test.unknownformat"); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestSaveUnknownFormat(t *testing.T) {
	sub := Subtitle{}
	if err := sub.Save("test.unknownformat"); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestSaveInvalidPath(t *testing.T) {
	sub := Subtitle{}
	if err := sub.Save("/nonexistent/directory/file.srt"); err == nil {
		t.Error("expected error for invalid path")
	}
}
