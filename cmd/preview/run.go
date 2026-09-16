package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Laurelin/inked"
)

func run(args []string, outputPath string) error {
	source, err := readSource(args)
	if err != nil {
		return err
	}

	return renderPreview(outputPath, string(source))
}

func readSource(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("usage: preview <html-file>")
	}

	source, err := os.ReadFile(args[0])
	if err != nil {
		return nil, fmt.Errorf("read HTML: %w", err)
	}

	return source, nil
}

func renderPreview(outputPath, source string) error {
	renderer, err := inked.New()
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create preview: %w", err)
	}
	defer output.Close()

	if err := renderer.Render(output, source); err != nil {
		return fmt.Errorf("render preview: %w", err)
	}

	return nil
}
