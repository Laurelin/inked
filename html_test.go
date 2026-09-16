package inked

import (
	"strings"
	"testing"
)

func TestResolveParagraph(t *testing.T) {
	document := mustResolve(t, `<p class="text-lg font-bold">Hello, Inked.</p>`)
	body := document.Children[0].(resolvedElement)
	paragraph := body.Children[0].(resolvedElement)
	text := paragraph.Children[0].(resolvedText)

	if text.Text != "Hello, Inked." {
		t.Fatalf("text = %q", text.Text)
	}
	if text.Style.FontSize != 13.5 || text.Style.Weight != weightBold {
		t.Fatalf("text style = %+v", text.Style)
	}
}

func TestResolveRejectsUnknownClass(t *testing.T) {
	_, err := resolveSource(`<p class="shadow-xl">Hello</p>`)
	if err == nil || !strings.Contains(err.Error(), `unsupported utility class "shadow-xl"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestResolveRejectsUnknownElement(t *testing.T) {
	_, err := resolveSource(`<table></table>`)
	if err == nil || !strings.Contains(err.Error(), "unsupported HTML element <table>") {
		t.Fatalf("error = %v", err)
	}
}

func mustResolve(t *testing.T, source string) resolvedDocument {
	t.Helper()
	document, err := resolveSource(source)
	if err != nil {
		t.Fatalf("resolve source: %v", err)
	}
	return document
}

func resolveSource(source string) (resolvedDocument, error) {
	document, err := parseHTML(source)
	if err != nil {
		return resolvedDocument{}, err
	}
	return resolveDocument(document)
}
