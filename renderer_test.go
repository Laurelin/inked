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

	if !bytes.Contains(output.Bytes(), []byte("/Count 1")) {
		t.Fatal("PDF does not contain one page")
	}
}

func TestRenderTextPDF(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	var output bytes.Buffer

	if err := renderer.Render(&output, "<p>Hello, world!</p>"); err != nil {
		t.Fatalf("render text PDF: %v", err)
	}

	if !bytes.HasPrefix(output.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a PDF")
	}

	if !bytes.Contains(output.Bytes(), []byte("%%EOF")) {
		t.Fatal("PDF has no end marker")
	}

	if !bytes.Contains(output.Bytes(), []byte("(Hello,) Tj")) {
		t.Fatal("PDF does not contain rendered text")
	}
}

func TestRenderRejectsUnknownUtility(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	err = renderer.Render(&bytes.Buffer{}, `<p class="unknown">Hello</p>`)
	if err == nil {
		t.Fatal("renderer accepted unknown utility")
	}
}

func TestRenderPropagatesWriterFailure(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	if err := renderer.Render(failingWriter{}, "<p>Hello</p>"); err == nil {
		t.Fatal("renderer accepted writer failure")
	}
}
