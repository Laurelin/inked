package inked

import (
	"fmt"
	"io"
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

	resolved, err := resolveDocument(document)
	if err != nil {
		return fmt.Errorf("resolve document: %w", err)
	}

	laidOut, err := layoutDocument(resolved)
	if err != nil {
		return fmt.Errorf("layout document: %w", err)
	}

	return writePDF(dst, laidOut)
}
