package inked

import (
	"fmt"
	"io"

	"github.com/signintech/gopdf"
	"golang.org/x/image/font/gofont/goregular"
)

type Renderer struct{}

func New() (*Renderer, error) {
	return &Renderer{}, nil
}

func (r *Renderer) Render(dst io.Writer, source string) error {
	document, err := parseHTML(source)
	if err != nil {
		return fmt.Errorf("parse HTML: %w", err)
	}

	var pdf gopdf.GoPdf

	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeLetter,
	})
	pdf.AddPage()

	text := documentText(document)
	if text != "" {
		if err := pdf.AddTTFFontData("Go", goregular.TTF); err != nil {
			return fmt.Errorf("add default font: %w", err)
		}

		if err := pdf.SetFont("Go", "", 12); err != nil {
			return fmt.Errorf("set default font: %w", err)
		}

		pdf.SetXY(72, 72)

		if err := pdf.Text(text); err != nil {
			return fmt.Errorf("draw text: %w", err)
		}
	}

	if _, err := pdf.WriteTo(dst); err != nil {
		return fmt.Errorf("write PDF: %w", err)
	}

	return nil
}