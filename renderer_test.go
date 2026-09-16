package inked

import (
	"bytes"
	"testing"
)

func TestRendererRenderProducesBlankPDF(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	var pdfOutput bytes.Buffer

	if err := renderer.Render(&pdfOutput, ""); err != nil {
		t.Fatalf("render blank PDF: %v", err)
	}

	if !bytes.HasPrefix(pdfOutput.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a PDF")
	}

	if !bytes.Contains(pdfOutput.Bytes(), []byte("%%EOF")) {
		t.Fatal("PDF has no end marker")
	}

	if !bytes.Contains(pdfOutput.Bytes(), []byte("/Count 1")) {
		t.Fatal("PDF does not contain one page")
	}
}

func TestRendererRenderProducesTextPDF(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	var pdfOutput bytes.Buffer

	if err := renderer.Render(&pdfOutput, "<p>Hello, world!</p>"); err != nil {
		t.Fatalf("render text PDF: %v", err)
	}

	if !bytes.HasPrefix(pdfOutput.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a PDF")
	}

	if !bytes.Contains(pdfOutput.Bytes(), []byte("%%EOF")) {
		t.Fatal("PDF has no end marker")
	}

	if !bytes.Contains(pdfOutput.Bytes(), []byte("(Hello,) Tj")) {
		t.Fatal("PDF does not contain rendered text")
	}
}

func TestRendererRenderRejectsUnknownUtility(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	err = renderer.Render(&bytes.Buffer{}, `<p class="unknown">Hello</p>`)
	if err == nil {
		t.Fatal("renderer accepted unknown utility")
	}
}

func TestRendererRenderPropagatesWriterFailure(t *testing.T) {
	renderer, err := New()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	if err := renderer.Render(failingWriter{}, "<p>Hello</p>"); err == nil {
		t.Fatal("renderer accepted writer failure")
	}
}
