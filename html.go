package inked

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func parseHTML(htmlSource string) (*html.Node, error) {
	return html.Parse(strings.NewReader(htmlSource))
}

type resolvedDocument struct {
	Children []resolvedNode
}

type resolvedNode interface {
	resolvedNode()
}

type resolvedElement struct {
	Style    computedStyle
	Children []resolvedNode
}

func (resolvedElement) resolvedNode() {}

type resolvedText struct {
	Style computedStyle
	Text  string
}

func (resolvedText) resolvedNode() {}

type resolvedLineBreak struct{}

func (resolvedLineBreak) resolvedNode() {}

func resolveDocument(root *html.Node) (resolvedDocument, error) {
	children, err := resolveChildren(root, initialStyle())
	if err != nil {
		return resolvedDocument{}, err
	}
	return resolvedDocument{Children: children}, nil
}

func resolveChildren(parent *html.Node, inheritedStyle computedStyle) ([]resolvedNode, error) {
	var children []resolvedNode
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		resolvedChildren, err := resolveNode(child, inheritedStyle)
		if err != nil {
			return nil, err
		}
		children = append(children, resolvedChildren...)
	}
	return children, nil
}

func resolveNode(node *html.Node, inheritedStyle computedStyle) ([]resolvedNode, error) {
	switch node.Type {
	case html.TextNode:
		return []resolvedNode{resolvedText{Style: inheritedStyle, Text: node.Data}}, nil
	case html.ElementNode:
		return resolveElement(node, inheritedStyle)
	default:
		return nil, nil
	}
}

func resolveElement(node *html.Node, inheritedStyle computedStyle) ([]resolvedNode, error) {
	if isIgnoredElement(node.Data) {
		return nil, nil
	}
	if node.Data == "html" {
		return resolveChildren(node, inheritedStyle)
	}
	if node.Data == "br" {
		return []resolvedNode{resolvedLineBreak{}}, nil
	}
	return resolveLayoutElement(node, inheritedStyle)
}

func resolveLayoutElement(node *html.Node, inheritedStyle computedStyle) ([]resolvedNode, error) {
	if !isSupportedElement(node.Data) {
		return nil, fmt.Errorf("unsupported HTML element <%s>", node.Data)
	}

	style, err := applyUtilityClasses(
		computeElementStyle(inheritedStyle, node.Data),
		elementClasses(node),
	)
	if err != nil {
		return nil, fmt.Errorf("<%s>: %w", node.Data, err)
	}

	children, err := resolveChildren(node, style)
	if err != nil {
		return nil, err
	}
	return []resolvedNode{resolvedElement{Style: style, Children: children}}, nil
}

func isIgnoredElement(name string) bool {
	return name == "head" || name == "script" || name == "style"
}

func isSupportedElement(name string) bool {
	return name == "body" || name == "div" || name == "p" || name == "span"
}

func elementClasses(node *html.Node) string {
	for _, attribute := range node.Attr {
		if attribute.Key == "class" {
			return attribute.Val
		}
	}
	return ""
}
