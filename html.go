package inked

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func parseHTML(source string) (*html.Node, error) {
	return html.Parse(strings.NewReader(source))
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

type resolvedBreak struct{}

func (resolvedBreak) resolvedNode() {}

func resolveDocument(document *html.Node) (resolvedDocument, error) {
	children, err := resolveChildren(document, defaultStyle())
	if err != nil {
		return resolvedDocument{}, err
	}
	return resolvedDocument{Children: children}, nil
}

func resolveChildren(parent *html.Node, inherited computedStyle) ([]resolvedNode, error) {
	var result []resolvedNode
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		nodes, err := resolveNode(child, inherited)
		if err != nil {
			return nil, err
		}
		result = append(result, nodes...)
	}
	return result, nil
}

func resolveNode(node *html.Node, inherited computedStyle) ([]resolvedNode, error) {
	switch node.Type {
	case html.TextNode:
		return []resolvedNode{resolvedText{Style: inherited, Text: node.Data}}, nil
	case html.ElementNode:
		return resolveElement(node, inherited)
	default:
		return nil, nil
	}
}

func resolveElement(node *html.Node, inherited computedStyle) ([]resolvedNode, error) {
	nodes, handled, err := resolveSpecialElement(node, inherited)
	if handled {
		return nodes, err
	}
	return resolveStyledElement(node, inherited)
}

func resolveSpecialElement(
	node *html.Node,
	inherited computedStyle,
) ([]resolvedNode, bool, error) {
	if ignoredElement(node.Data) {
		return nil, true, nil
	}
	if node.Data == "html" {
		children, err := resolveChildren(node, inherited)
		return children, true, err
	}
	if node.Data == "br" {
		return []resolvedNode{resolvedBreak{}}, true, nil
	}
	return nil, false, nil
}

func resolveStyledElement(node *html.Node, inherited computedStyle) ([]resolvedNode, error) {
	if !supportedElement(node.Data) {
		return nil, fmt.Errorf("unsupported HTML element <%s>", node.Data)
	}

	style, err := applyClasses(elementStyle(inherited, node.Data), classAttribute(node))
	if err != nil {
		return nil, fmt.Errorf("<%s>: %w", node.Data, err)
	}

	children, err := resolveChildren(node, style)
	if err != nil {
		return nil, err
	}
	return []resolvedNode{resolvedElement{Style: style, Children: children}}, nil
}

func ignoredElement(name string) bool {
	return name == "head" || name == "script" || name == "style"
}

func supportedElement(name string) bool {
	return name == "body" || name == "div" || name == "p" || name == "span"
}

func classAttribute(node *html.Node) string {
	for _, attribute := range node.Attr {
		if attribute.Key == "class" {
			return attribute.Val
		}
	}
	return ""
}
