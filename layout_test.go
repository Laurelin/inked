package inked

import (
	"strings"
	"testing"
)

func TestLayoutDocumentAppliesAlignmentAndSpacing(t *testing.T) {
	document := mustResolveDocument(t, `<p class="text-center mx-4 px-2">Hello</p>`)
	layout := mustLayoutDocument(t, document)
	run := layout.Pages[0].Runs[0]

	if run.X <= defaultPageMargin+18 {
		t.Fatalf("centered run x = %v", run.X)
	}
	if run.Text != "Hello" {
		t.Fatalf("run text = %q", run.Text)
	}
}

func TestLayoutDocumentWrapsLongText(t *testing.T) {
	htmlSource := `<p>` + strings.Repeat("wrapping ", 30) + `</p>`
	layout := mustLayoutDocument(t, mustResolveDocument(t, htmlSource))
	runs := layout.Pages[0].Runs

	if runs[len(runs)-1].Y == runs[0].Y {
		t.Fatal("text did not wrap")
	}
}

func TestLayoutDocumentSplitsLongWord(t *testing.T) {
	htmlSource := `<p>` + strings.Repeat("W", 100) + `</p>`
	layout := mustLayoutDocument(t, mustResolveDocument(t, htmlSource))

	if len(layout.Pages[0].Runs) < 2 {
		t.Fatal("long word was not split")
	}
}

func TestLayoutDocumentHandlesInlineSpacingAndLineBreak(t *testing.T) {
	htmlSource := `<p>Hello<span class="px-2"> world</span><br>next</p>`
	layout := mustLayoutDocument(t, mustResolveDocument(t, htmlSource))
	runs := layout.Pages[0].Runs

	if runs[len(runs)-1].Y == runs[0].Y {
		t.Fatal("break did not start a new line")
	}
}

func TestLayoutDocumentPaginates(t *testing.T) {
	htmlSource := `<body>` + strings.Repeat("<p>line</p>", 30) + `</body>`
	layout := mustLayoutDocument(t, mustResolveDocument(t, htmlSource))

	if len(layout.Pages) < 2 {
		t.Fatal("document did not paginate")
	}
}

func TestLayoutDocumentRejectsUnsupportedCharacter(t *testing.T) {
	document := mustResolveDocument(t, `<p>snowman: ☃</p>`)
	if _, err := layoutDocument(document); err == nil {
		t.Fatal("layout accepted unsupported character")
	}
}

func mustLayoutDocument(t *testing.T, document resolvedDocument) documentLayout {
	t.Helper()
	layout, err := layoutDocument(document)
	if err != nil {
		t.Fatalf("layout document: %v", err)
	}
	return layout
}
