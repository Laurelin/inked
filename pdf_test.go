package inked

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestBuildPDFWritesValidOffsets(t *testing.T) {
	data, err := buildPDF(laidOutDocument{Pages: []laidOutPage{{}}})
	if err != nil {
		t.Fatalf("build PDF: %v", err)
	}
	pdf := string(data)

	start := strings.LastIndex(pdf, "startxref\n")
	valueStart := start + len("startxref\n")
	valueEnd := strings.IndexByte(pdf[valueStart:], '\n') + valueStart
	xrefOffset, err := strconv.Atoi(pdf[valueStart:valueEnd])
	if err != nil {
		t.Fatalf("parse startxref: %v", err)
	}
	if pdf[xrefOffset:xrefOffset+4] != "xref" {
		t.Fatalf("startxref points to %q", pdf[xrefOffset:xrefOffset+4])
	}

	xrefLines := strings.Split(pdf[xrefOffset:], "\n")
	objectOffset, err := strconv.Atoi(xrefLines[3][:10])
	if err != nil {
		t.Fatalf("parse object offset: %v", err)
	}
	if !strings.HasPrefix(pdf[objectOffset:], "1 0 obj") {
		t.Fatalf("object offset points to %q", pdf[objectOffset:objectOffset+7])
	}
}

func TestBuildPDFIsDeterministic(t *testing.T) {
	document := laidOutDocument{Pages: []laidOutPage{{Runs: []textRun{{
		X: 72, Y: 84, Text: "Hello", Style: defaultStyle(),
	}}}}}
	first, err := buildPDF(document)
	if err != nil {
		t.Fatalf("build first PDF: %v", err)
	}
	second, err := buildPDF(document)
	if err != nil {
		t.Fatalf("build second PDF: %v", err)
	}
	if string(first) != string(second) {
		t.Fatal("PDF output is not deterministic")
	}
}

func TestWritePDFPropagatesWriterFailure(t *testing.T) {
	document := laidOutDocument{Pages: []laidOutPage{{}}}
	if err := writePDF(failingWriter{}, document); err == nil {
		t.Fatal("writePDF accepted writer failure")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
