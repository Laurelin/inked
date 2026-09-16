package main

import (
	"log"
	"os"
)

func main() {
	if err := run(os.Args[1:], "preview.pdf"); err != nil {
		log.Fatal(err)
	}
}
