package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Laurelin/inked"
)

func run(args []string, outputPath string) error {
	htmlSource, err := readHTMLSource(args)
	if err != nil {
		return err
	}

	return renderPreview(outputPath, string(htmlSource))
}

func readHTMLSource(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("usage: preview <html-file>")
	}

	htmlSource, err := os.ReadFile(args[0])
	if err != nil {
		return nil, fmt.Errorf("read HTML: %w", err)
	}

	return htmlSource, nil
}

func renderPreview(outputPath, htmlSource string) error {
	renderer, err := inked.New()
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create preview: %w", err)
	}
	defer outputFile.Close()

	if err := renderer.Render(outputFile, htmlSource); err != nil {
		return fmt.Errorf("render preview: %w", err)
	}

	return nil
}
