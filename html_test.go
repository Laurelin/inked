package inked

import (
	"testing"
	"golang.org/x/net/html"
	"slices"
)

func TestParseParagraph(t *testing.T) {
	document, err := parseHTML(`<p>Hello, Inked.</p>`)
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}

	if got, want := documentText(document), "Hello, Inked."; got != want {
		t.Fatalf("document text = %q, want %q", got, want)
	}
}

func TestWalkSkipsChildren(t *testing.T) {
	root := &html.Node{Type: html.ElementNode, Data: "div"}
	skipped := &html.Node{Type: html.ElementNode, Data: "script"}
	hidden := &html.Node{Type: html.TextNode, Data: "hidden"}
	paragraph := &html.Node{Type: html.ElementNode, Data: "p"}
	visible := &html.Node{Type: html.TextNode, Data: "visible"}

	root.AppendChild(skipped)
	skipped.AppendChild(hidden)
	root.AppendChild(paragraph)
	paragraph.AppendChild(visible)

	var visited []string

	walk(root, func(node *html.Node) bool {
		visited = append(visited, node.Data)
		return node != skipped
	})

	want := []string{"div", "script", "p", "visible"}
	if !slices.Equal(visited, want) {
		t.Fatalf("visited %q, want %q", visited, want)
	}
}