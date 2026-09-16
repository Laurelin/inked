package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadSourceRequiresPath(t *testing.T) {
	if _, err := readSource(nil); err == nil {
		t.Fatal("readSource accepted no path")
	}
}

func TestRunWritesPreview(t *testing.T) {
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "source.html")
	outputPath := filepath.Join(directory, "preview.pdf")

	if err := os.WriteFile(sourcePath, []byte("<p>Hello, world!</p>"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if err := run([]string{sourcePath}, outputPath); err != nil {
		t.Fatalf("run preview: %v", err)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read preview: %v", err)
	}

	if !bytes.HasPrefix(output, []byte("%PDF-")) {
		t.Fatal("preview is not a PDF")
	}
}

func TestRenderPreviewRejectsDirectoryOutput(t *testing.T) {
	if err := renderPreview(t.TempDir(), "<p>Hello</p>"); err == nil {
		t.Fatal("renderPreview accepted a directory output path")
	}
}
