package main

import (
	"log"
	"os"

	"github.com/Laurelin/inked"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: preview <html-file>")
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	renderer, err := inked.New()
	if err != nil {
		log.Fatal(err)
	}

	output, err := os.Create("preview.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	if err := renderer.Render(output, string(source)); err != nil {
		log.Fatal(err)
	}
}