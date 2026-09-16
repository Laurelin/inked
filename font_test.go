package inked

import "testing"

func TestEncodeWinANSISupportsWesternText(t *testing.T) {
	encoded, err := encodeWinANSI("café — €")
	if err != nil {
		t.Fatalf("encode text: %v", err)
	}
	if string(encoded) != "caf\xe9 \x97 \x80" {
		t.Fatalf("encoded text = %q", encoded)
	}
}

func TestEscapePDFText(t *testing.T) {
	escaped, err := escapePDFText(`a(b)\c`)
	if err != nil {
		t.Fatalf("escape text: %v", err)
	}
	if escaped != `a\(b\)\\c` {
		t.Fatalf("escaped text = %q", escaped)
	}
}

func TestMeasureTextUsesBoldMetrics(t *testing.T) {
	normal := defaultStyle()
	bold := normal
	bold.Weight = weightBold

	normalWidth, err := measureText("Hello", normal)
	if err != nil {
		t.Fatalf("measure normal: %v", err)
	}
	boldWidth, err := measureText("Hello", bold)
	if err != nil {
		t.Fatalf("measure bold: %v", err)
	}
	if boldWidth <= normalWidth {
		t.Fatalf("bold width %v <= normal width %v", boldWidth, normalWidth)
	}
}
