package inked

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestSerializePDFProducesValidOffsets(t *testing.T) {
	serializedPDF, err := serializePDF(documentLayout{Pages: []pageLayout{{}}})
	if err != nil {
		t.Fatalf("serialize PDF: %v", err)
	}
	pdf := string(serializedPDF)

	start := strings.LastIndex(pdf, "startxref\n")
	valueStart := start + len("startxref\n")
	valueEnd := strings.IndexByte(pdf[valueStart:], '\n') + valueStart
	crossReferenceOffset, err := strconv.Atoi(pdf[valueStart:valueEnd])
	if err != nil {
		t.Fatalf("parse startxref: %v", err)
	}
	if pdf[crossReferenceOffset:crossReferenceOffset+4] != "xref" {
		t.Fatalf(
			"startxref points to %q",
			pdf[crossReferenceOffset:crossReferenceOffset+4],
		)
	}

	crossReferenceLines := strings.Split(pdf[crossReferenceOffset:], "\n")
	objectOffset, err := strconv.Atoi(crossReferenceLines[3][:10])
	if err != nil {
		t.Fatalf("parse object offset: %v", err)
	}
	if !strings.HasPrefix(pdf[objectOffset:], "1 0 obj") {
		t.Fatalf("object offset points to %q", pdf[objectOffset:objectOffset+7])
	}
}

func TestSerializePDFIsDeterministic(t *testing.T) {
	document := documentLayout{Pages: []pageLayout{{Runs: []textRun{{
		X: 72, Y: 84, Text: "Hello", Style: initialStyle(),
	}}}}}
	firstPDF, err := serializePDF(document)
	if err != nil {
		t.Fatalf("serialize first PDF: %v", err)
	}
	secondPDF, err := serializePDF(document)
	if err != nil {
		t.Fatalf("serialize second PDF: %v", err)
	}
	if string(firstPDF) != string(secondPDF) {
		t.Fatal("PDF output is not deterministic")
	}
}

func TestWritePDFPropagatesWriterFailure(t *testing.T) {
	document := documentLayout{Pages: []pageLayout{{}}}
	if err := writePDF(failingWriter{}, document); err == nil {
		t.Fatal("writePDF accepted writer failure")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
