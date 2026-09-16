package inked

import (
	"strings"
	"testing"
)

func TestResolveDocumentAppliesParagraphStyle(t *testing.T) {
	document := mustResolveDocument(t, `<p class="text-lg font-bold">Hello, Inked.</p>`)
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

func TestResolveHTMLSourceRejectsUnknownClass(t *testing.T) {
	_, err := resolveHTMLSource(`<p class="shadow-xl">Hello</p>`)
	if err == nil || !strings.Contains(err.Error(), `unsupported utility class "shadow-xl"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestResolveHTMLSourceRejectsUnknownElement(t *testing.T) {
	_, err := resolveHTMLSource(`<table></table>`)
	if err == nil || !strings.Contains(err.Error(), "unsupported HTML element <table>") {
		t.Fatalf("error = %v", err)
	}
}

func mustResolveDocument(t *testing.T, htmlSource string) resolvedDocument {
	t.Helper()
	document, err := resolveHTMLSource(htmlSource)
	if err != nil {
		t.Fatalf("resolve HTML source: %v", err)
	}
	return document
}

func resolveHTMLSource(htmlSource string) (resolvedDocument, error) {
	htmlDocument, err := parseHTML(htmlSource)
	if err != nil {
		return resolvedDocument{}, err
	}
	return resolveDocument(htmlDocument)
}
