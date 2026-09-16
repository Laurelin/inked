package inked

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func writePDF(destination io.Writer, document laidOutDocument) error {
	data, err := buildPDF(document)
	if err != nil {
		return err
	}
	if _, err := io.Copy(destination, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("write PDF: %w", err)
	}
	return nil
}

func buildPDF(document laidOutDocument) ([]byte, error) {
	objects, err := pdfObjects(document)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	output.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	offsets := writeObjects(&output, objects)
	writeXref(&output, offsets)
	return output.Bytes(), nil
}

func pdfObjects(document laidOutDocument) ([]string, error) {
	pageCount := len(document.Pages)
	fontNormal := 3 + pageCount*2
	fontBold := fontNormal + 1

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		pagesObject(pageCount),
	}
	for index, page := range document.Pages {
		content, err := pageContent(page)
		if err != nil {
			return nil, err
		}
		pageNumber := 3 + index*2
		contentNumber := pageNumber + 1
		objects = append(objects,
			pageObject(contentNumber, fontNormal, fontBold),
			streamObject(content),
		)
	}
	objects = append(objects,
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>",
	)
	return objects, nil
}

func pagesObject(pageCount int) string {
	var children strings.Builder
	for index := range pageCount {
		fmt.Fprintf(&children, "%d 0 R ", 3+index*2)
	}
	return fmt.Sprintf(
		"<< /Type /Pages /Kids [%s] /Count %d >>",
		children.String(),
		pageCount,
	)
}

func pageObject(contentNumber, fontNormal, fontBold int) string {
	return fmt.Sprintf(
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] "+
			"/Resources << /Font << /F1 %d 0 R /F2 %d 0 R >> >> "+
			"/Contents %d 0 R >>",
		fontNormal,
		fontBold,
		contentNumber,
	)
}

func pageContent(page laidOutPage) (string, error) {
	var content strings.Builder
	for _, run := range page.Runs {
		text, err := escapePDFText(run.Text)
		if err != nil {
			return "", err
		}
		font := "F1"
		if run.Style.Weight == weightBold {
			font = "F2"
		}
		fmt.Fprintf(
			&content,
			"BT /%s %s Tf 1 0 0 1 %s %s Tm (%s) Tj ET\n",
			font,
			pdfNumber(run.Style.FontSize),
			pdfNumber(run.X),
			pdfNumber(letterHeight-run.Y),
			text,
		)
	}
	return content.String(), nil
}

func streamObject(content string) string {
	return fmt.Sprintf(
		"<< /Length %d >>\nstream\n%sendstream",
		len([]byte(content)),
		content,
	)
}

func writeObjects(output *bytes.Buffer, objects []string) []int {
	offsets := make([]int, len(objects))
	for index, object := range objects {
		offsets[index] = output.Len()
		fmt.Fprintf(output, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}
	return offsets
}

func writeXref(output *bytes.Buffer, offsets []int) {
	xrefOffset := output.Len()
	fmt.Fprintf(output, "xref\n0 %d\n", len(offsets)+1)
	output.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		fmt.Fprintf(output, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(
		output,
		"trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(offsets)+1,
		xrefOffset,
	)
}

func pdfNumber(value float64) string {
	result := strconv.FormatFloat(value, 'f', 2, 64)
	result = strings.TrimRight(result, "0")
	result = strings.TrimRight(result, ".")
	if result == "-0" || result == "" {
		return "0"
	}
	return result
}
