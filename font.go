package inked

import (
	"fmt"
	"strings"
)

var pdfTextEscaper = strings.NewReplacer(
	`\`, `\\`,
	`(`, `\(`,
	`)`, `\)`,
)

var helveticaWidths = map[byte]float64{
	' ': 278, '!': 278, '"': 355, '#': 556, '$': 556, '%': 889, '&': 667, '\'': 191,
	'(': 333, ')': 333, '*': 389, '+': 584, ',': 278, '-': 333, '.': 278, '/': 278,
	'0': 556, '1': 556, '2': 556, '3': 556, '4': 556, '5': 556, '6': 556, '7': 556,
	'8': 556, '9': 556, ':': 278, ';': 278, '<': 584, '=': 584, '>': 584, '?': 556,
	'@': 1015, 'A': 667, 'B': 667, 'C': 722, 'D': 722, 'E': 667, 'F': 611, 'G': 778,
	'H': 722, 'I': 278, 'J': 500, 'K': 667, 'L': 556, 'M': 833, 'N': 722, 'O': 778,
	'P': 667, 'Q': 778, 'R': 722, 'S': 667, 'T': 611, 'U': 722, 'V': 667, 'W': 944,
	'X': 667, 'Y': 667, 'Z': 611, '[': 278, '\\': 278, ']': 278, '^': 469, '_': 556,
	'`': 333, 'a': 556, 'b': 556, 'c': 500, 'd': 556, 'e': 556, 'f': 278, 'g': 556,
	'h': 556, 'i': 222, 'j': 222, 'k': 500, 'l': 222, 'm': 833, 'n': 556, 'o': 556,
	'p': 556, 'q': 556, 'r': 333, 's': 500, 't': 278, 'u': 556, 'v': 500, 'w': 722,
	'x': 500, 'y': 500, 'z': 500, '{': 334, '|': 260, '}': 334, '~': 584,
}

var helveticaBoldWidths = map[byte]float64{
	' ': 278, '!': 333, '"': 474, '#': 556, '$': 556, '%': 889, '&': 722, '\'': 238,
	'(': 333, ')': 333, '*': 389, '+': 584, ',': 278, '-': 333, '.': 278, '/': 278,
	'0': 556, '1': 556, '2': 556, '3': 556, '4': 556, '5': 556, '6': 556, '7': 556,
	'8': 556, '9': 556, ':': 333, ';': 333, '<': 584, '=': 584, '>': 584, '?': 611,
	'@': 975, 'A': 722, 'B': 722, 'C': 722, 'D': 722, 'E': 667, 'F': 611, 'G': 778,
	'H': 722, 'I': 278, 'J': 556, 'K': 722, 'L': 611, 'M': 833, 'N': 722, 'O': 778,
	'P': 667, 'Q': 778, 'R': 722, 'S': 667, 'T': 611, 'U': 722, 'V': 667, 'W': 944,
	'X': 667, 'Y': 667, 'Z': 611, '[': 333, '\\': 278, ']': 333, '^': 584, '_': 556,
	'`': 333, 'a': 556, 'b': 611, 'c': 556, 'd': 611, 'e': 556, 'f': 333, 'g': 611,
	'h': 611, 'i': 278, 'j': 278, 'k': 556, 'l': 278, 'm': 889, 'n': 611, 'o': 611,
	'p': 611, 'q': 611, 'r': 389, 's': 556, 't': 333, 'u': 611, 'v': 556, 'w': 778,
	'x': 556, 'y': 556, 'z': 500, '{': 389, '|': 280, '}': 389, '~': 584,
}

var winANSISpecial = map[rune]byte{
	'€': 128, '‚': 130, 'ƒ': 131, '„': 132, '…': 133, '†': 134, '‡': 135,
	'ˆ': 136, '‰': 137, 'Š': 138, '‹': 139, 'Œ': 140, 'Ž': 142, '‘': 145,
	'’': 146, '“': 147, '”': 148, '•': 149, '–': 150, '—': 151, '˜': 152,
	'™': 153, 'š': 154, '›': 155, 'œ': 156, 'ž': 158, 'Ÿ': 159,
}

func encodeWinANSI(text string) ([]byte, error) {
	var encoded []byte
	for _, character := range text {
		value, ok := winANSIByte(character)
		if !ok {
			return nil, fmt.Errorf("character %q is not supported by WinAnsi", character)
		}
		encoded = append(encoded, value)
	}
	return encoded, nil
}

func winANSIByte(character rune) (byte, bool) {
	if character >= 32 && character <= 126 {
		return byte(character), true
	}
	if character >= 160 && character <= 255 {
		return byte(character), true
	}
	value, ok := winANSISpecial[character]
	return value, ok
}

func measureText(text string, style computedStyle) (float64, error) {
	encoded, err := encodeWinANSI(text)
	if err != nil {
		return 0, err
	}

	widths := helveticaWidths
	if style.Weight == weightBold {
		widths = helveticaBoldWidths
	}

	var units float64
	for _, character := range encoded {
		units += widths[character]
		if _, ok := widths[character]; !ok {
			units += 600
		}
	}
	return units * style.FontSize / 1000, nil
}

func escapePDFText(text string) (string, error) {
	encoded, err := encodeWinANSI(text)
	if err != nil {
		return "", err
	}
	return pdfTextEscaper.Replace(string(encoded)), nil
}
