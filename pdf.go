package inked

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type objectNumber int

type objectOffset int

type objectBody string

type contentStream string

func writePDF(destination io.Writer, document documentLayout) error {
	serializedPDF, err := serializePDF(document)
	if err != nil {
		return err
	}
	if _, err := io.Copy(destination, bytes.NewReader(serializedPDF)); err != nil {
		return fmt.Errorf("write PDF: %w", err)
	}
	return nil
}

func serializePDF(document documentLayout) ([]byte, error) {
	objectBodies, err := buildObjectBodies(document)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	objectOffsets := writeIndirectObjects(&buffer, objectBodies)
	writeCrossReferenceTable(&buffer, objectOffsets)
	return buffer.Bytes(), nil
}

func buildObjectBodies(document documentLayout) ([]objectBody, error) {
	pageCount := len(document.Pages)
	normalFontObjectNumber := objectNumber(3 + pageCount*2)
	boldFontObjectNumber := normalFontObjectNumber + 1

	objectBodies := []objectBody{
		"<< /Type /Catalog /Pages 2 0 R >>",
		formatPageTreeDictionary(pageCount),
	}
	for index, page := range document.Pages {
		pageContent, err := encodePageContent(page)
		if err != nil {
			return nil, err
		}
		pageObjectNumber := objectNumber(3 + index*2)
		contentObjectNumber := pageObjectNumber + 1
		objectBodies = append(objectBodies,
			formatPageDictionary(
				contentObjectNumber,
				normalFontObjectNumber,
				boldFontObjectNumber,
			),
			formatStreamObject(pageContent),
		)
	}
	objectBodies = append(objectBodies,
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>",
	)
	return objectBodies, nil
}

func formatPageTreeDictionary(pageCount int) objectBody {
	var pageReferences strings.Builder
	for index := range pageCount {
		fmt.Fprintf(&pageReferences, "%d 0 R ", 3+index*2)
	}
	return objectBody(fmt.Sprintf(
		"<< /Type /Pages /Kids [%s] /Count %d >>",
		pageReferences.String(),
		pageCount,
	))
}

func formatPageDictionary(
	contentObjectNumber objectNumber,
	normalFontObjectNumber objectNumber,
	boldFontObjectNumber objectNumber,
) objectBody {
	return objectBody(fmt.Sprintf(
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] "+
			"/Resources << /Font << /F1 %d 0 R /F2 %d 0 R >> >> "+
			"/Contents %d 0 R >>",
		normalFontObjectNumber,
		boldFontObjectNumber,
		contentObjectNumber,
	))
}

func encodePageContent(page pageLayout) (contentStream, error) {
	var stream strings.Builder
	for _, run := range page.Runs {
		encodedText, err := encodePDFLiteralString(run.Text)
		if err != nil {
			return "", err
		}
		fontResourceName := "F1"
		if run.Style.Weight == weightBold {
			fontResourceName = "F2"
		}
		fmt.Fprintf(
			&stream,
			"BT /%s %s Tf 1 0 0 1 %s %s Tm (%s) Tj ET\n",
			fontResourceName,
			formatPDFNumber(run.Style.FontSize),
			formatPDFNumber(run.X),
			formatPDFNumber(letterPageHeight-run.Y),
			encodedText,
		)
	}
	return contentStream(stream.String()), nil
}

func formatStreamObject(content contentStream) objectBody {
	return objectBody(fmt.Sprintf(
		"<< /Length %d >>\nstream\n%sendstream",
		len([]byte(content)),
		content,
	))
}

func writeIndirectObjects(
	buffer *bytes.Buffer,
	objectBodies []objectBody,
) []objectOffset {
	objectOffsets := make([]objectOffset, len(objectBodies))
	for index, body := range objectBodies {
		objectOffsets[index] = objectOffset(buffer.Len())
		fmt.Fprintf(buffer, "%d 0 obj\n%s\nendobj\n", index+1, body)
	}
	return objectOffsets
}

func writeCrossReferenceTable(buffer *bytes.Buffer, objectOffsets []objectOffset) {
	crossReferenceOffset := buffer.Len()
	fmt.Fprintf(buffer, "xref\n0 %d\n", len(objectOffsets)+1)
	buffer.WriteString("0000000000 65535 f \n")
	for _, offset := range objectOffsets {
		fmt.Fprintf(buffer, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(
		buffer,
		"trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(objectOffsets)+1,
		crossReferenceOffset,
	)
}

func formatPDFNumber(value float64) string {
	formattedNumber := strconv.FormatFloat(value, 'f', 2, 64)
	formattedNumber = strings.TrimRight(formattedNumber, "0")
	formattedNumber = strings.TrimRight(formattedNumber, ".")
	if formattedNumber == "-0" || formattedNumber == "" {
		return "0"
	}
	return formattedNumber
}
