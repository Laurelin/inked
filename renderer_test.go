package inked

import (
	"bytes"
	"testing"
)

func TestRenderBlankPDF(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	var output bytes.Buffer

	if err := renderer.Render(&output, ""); err != nil {
		t.Fatalf("render blank PDF: %v", err)
	}

	if !bytes.HasPrefix(output.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a PDF")
	}

	if !bytes.Contains(output.Bytes(), []byte("%%EOF")) {
		t.Fatal("PDF has no end marker")
	}
}