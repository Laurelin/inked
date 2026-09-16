package inked

import (
	"strings"
	"testing"
)

func TestLayoutAppliesAlignmentAndSpacing(t *testing.T) {
	document := mustResolve(t, `<p class="text-center mx-4 px-2">Hello</p>`)
	layout := mustLayout(t, document)
	run := layout.Pages[0].Runs[0]

	if run.X <= pageInset+18 {
		t.Fatalf("centered run x = %v", run.X)
	}
	if run.Text != "Hello" {
		t.Fatalf("run text = %q", run.Text)
	}
}

func TestLayoutWrapsLongText(t *testing.T) {
	source := `<p>` + strings.Repeat("wrapping ", 30) + `</p>`
	layout := mustLayout(t, mustResolve(t, source))
	runs := layout.Pages[0].Runs

	if runs[len(runs)-1].Y == runs[0].Y {
		t.Fatal("text did not wrap")
	}
}

func TestLayoutSplitsLongWord(t *testing.T) {
	source := `<p>` + strings.Repeat("W", 100) + `</p>`
	layout := mustLayout(t, mustResolve(t, source))

	if len(layout.Pages[0].Runs) < 2 {
		t.Fatal("long word was not split")
	}
}

func TestLayoutHandlesInlineSpacingAndBreak(t *testing.T) {
	source := `<p>Hello<span class="px-2"> world</span><br>next</p>`
	layout := mustLayout(t, mustResolve(t, source))
	runs := layout.Pages[0].Runs

	if runs[len(runs)-1].Y == runs[0].Y {
		t.Fatal("break did not start a new line")
	}
}

func TestLayoutPaginates(t *testing.T) {
	source := `<body>` + strings.Repeat("<p>line</p>", 30) + `</body>`
	layout := mustLayout(t, mustResolve(t, source))

	if len(layout.Pages) < 2 {
		t.Fatal("document did not paginate")
	}
}

func TestLayoutRejectsUnsupportedCharacter(t *testing.T) {
	document := mustResolve(t, `<p>snowman: ☃</p>`)
	if _, err := layoutDocument(document); err == nil {
		t.Fatal("layout accepted unsupported character")
	}
}

func mustLayout(t *testing.T, document resolvedDocument) laidOutDocument {
	t.Helper()
	layout, err := layoutDocument(document)
	if err != nil {
		t.Fatalf("layout document: %v", err)
	}
	return layout
}
