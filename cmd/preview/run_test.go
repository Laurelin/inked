package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadHTMLSourceRequiresPath(t *testing.T) {
	if _, err := readHTMLSource(nil); err == nil {
		t.Fatal("readHTMLSource accepted no path")
	}
}

func TestRunWritesPreview(t *testing.T) {
	directory := t.TempDir()
	htmlPath := filepath.Join(directory, "source.html")
	outputPath := filepath.Join(directory, "preview.pdf")

	if err := os.WriteFile(htmlPath, []byte("<p>Hello, world!</p>"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if err := run([]string{htmlPath}, outputPath); err != nil {
		t.Fatalf("run preview: %v", err)
	}

	previewPDF, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read preview: %v", err)
	}

	if !bytes.HasPrefix(previewPDF, []byte("%PDF-")) {
		t.Fatal("preview is not a PDF")
	}
}

func TestRenderPreviewRejectsDirectoryOutput(t *testing.T) {
	if err := renderPreview(t.TempDir(), "<p>Hello</p>"); err == nil {
		t.Fatal("renderPreview accepted a directory output path")
	}
}
