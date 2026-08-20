package inked

import (
	"strings"

	"golang.org/x/net/html"
)

func parseHTML(source string) (*html.Node, error) {
	return html.Parse(strings.NewReader(source))
}

func walk(node *html.Node, visit func(*html.Node) bool) {
	if !visit(node) {
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child, visit)
	}
}

func documentText(document *html.Node) string {
	root := findElement(document, "body")
	if root == nil {
		root = document
	}

	var text strings.Builder
	
	visit := func(node *html.Node) bool {
		if node.Type == html.ElementNode &&
			(node.Data == "script" || node.Data == "style") {
			return false
		}

		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		return true
	}

	walk(root, visit)

	return strings.Join(strings.Fields(text.String()), " ")
}

func findElement(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && node.Data == name {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, name); found != nil {
			return found
		}
	}

	return nil
}