package inked

import (
	"fmt"
	"io"
)

type Renderer struct{}

func New() (*Renderer, error) {
	return &Renderer{}, nil
}

func (r *Renderer) Render(destination io.Writer, htmlSource string) error {
	htmlDocument, err := parseHTML(htmlSource)
	if err != nil {
		return fmt.Errorf("parse HTML: %w", err)
	}

	resolvedTree, err := resolveDocument(htmlDocument)
	if err != nil {
		return fmt.Errorf("resolve document: %w", err)
	}

	layout, err := layoutDocument(resolvedTree)
	if err != nil {
		return fmt.Errorf("layout document: %w", err)
	}

	return writePDF(destination, layout)
}
