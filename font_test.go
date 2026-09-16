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

func TestEncodePDFLiteralString(t *testing.T) {
	escaped, err := encodePDFLiteralString(`a(b)\c`)
	if err != nil {
		t.Fatalf("escape text: %v", err)
	}
	if escaped != `a\(b\)\\c` {
		t.Fatalf("escaped text = %q", escaped)
	}
}

func TestMeasureTextWidthUsesBoldMetrics(t *testing.T) {
	normal := initialStyle()
	bold := normal
	bold.Weight = weightBold

	normalWidth, err := measureTextWidth("Hello", normal)
	if err != nil {
		t.Fatalf("measure normal: %v", err)
	}
	boldWidth, err := measureTextWidth("Hello", bold)
	if err != nil {
		t.Fatalf("measure bold: %v", err)
	}
	if boldWidth <= normalWidth {
		t.Fatalf("bold width %v <= normal width %v", boldWidth, normalWidth)
	}
}
