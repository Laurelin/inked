package inked

import (
	"reflect"
	"testing"
)

func TestApplyUtilityClassesResolvesUtilities(t *testing.T) {
	style, err := applyUtilityClasses(
		initialStyle(),
		"block text-xl font-bold leading-tight text-right mx-4 pt-2",
	)
	if err != nil {
		t.Fatalf("apply classes: %v", err)
	}

	want := computedStyle{
		Display: displayBlock, Alignment: alignRight, Weight: weightBold,
		FontSize: 15, LineHeightFactor: 1.25,
		Margin:  boxEdges{Left: 12, Right: 12},
		Padding: boxEdges{Top: 6},
	}
	if !reflect.DeepEqual(style, want) {
		t.Fatalf("style = %+v, want %+v", style, want)
	}
}

func TestApplyUtilityClassesRejectsInvalidSpacing(t *testing.T) {
	if _, err := applyUtilityClasses(initialStyle(), "mz-2"); err == nil {
		t.Fatal("accepted invalid spacing axis")
	}
}

func TestDocumentedUtilitiesAreSupported(t *testing.T) {
	classes := []string{
		"block", "inline",
		"text-xs", "text-sm", "text-base", "text-lg", "text-xl",
		"font-normal", "font-bold",
		"leading-none", "leading-tight", "leading-normal", "leading-relaxed",
		"text-left", "text-center", "text-right",
		"m-0", "mx-1", "my-2", "mt-4", "mr-6", "mb-8", "ml-1",
		"p-0", "px-1", "py-2", "pt-4", "pr-6", "pb-8", "pl-1",
	}

	for _, class := range classes {
		if _, err := applyUtilityClasses(initialStyle(), class); err != nil {
			t.Errorf("%s: %v", class, err)
		}
	}
}

func TestComputeElementStyleInheritsTypographyOnly(t *testing.T) {
	parent := initialStyle()
	parent.Weight = weightBold
	parent.Margin.Left = 24

	child := computeElementStyle(parent, "span")

	if child.Weight != weightBold {
		t.Fatal("font weight was not inherited")
	}
	if child.Margin.Left != 0 {
		t.Fatal("margin was inherited")
	}
}
