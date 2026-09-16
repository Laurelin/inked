package inked

import (
	"fmt"
	"strconv"
	"strings"
)

type displayMode uint8

const (
	displayInline displayMode = iota
	displayBlock
)

type textAlignment uint8

const (
	alignLeft textAlignment = iota
	alignCenter
	alignRight
)

type fontWeight uint8

const (
	weightNormal fontWeight = iota
	weightBold
)

type boxEdges struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

type computedStyle struct {
	Display          displayMode
	Alignment        textAlignment
	Weight           fontWeight
	FontSize         float64
	LineHeightFactor float64
	Margin           boxEdges
	Padding          boxEdges
}

var fontSizePointsByUtility = map[string]float64{
	"text-xs": 9, "text-sm": 10.5, "text-base": 12, "text-lg": 13.5, "text-xl": 15,
}

var fontWeightByUtility = map[string]fontWeight{
	"font-normal": weightNormal, "font-bold": weightBold,
}

var lineHeightFactorByUtility = map[string]float64{
	"leading-none": 1, "leading-tight": 1.25,
	"leading-normal": 1.5, "leading-relaxed": 1.625,
}

var textAlignmentByUtility = map[string]textAlignment{
	"text-left": alignLeft, "text-center": alignCenter, "text-right": alignRight,
}

var supportedSpacingAxes = map[string]struct{}{
	"": {}, "x": {}, "y": {}, "t": {}, "r": {}, "b": {}, "l": {},
}

var spacingAxisSetters = map[string]func(*boxEdges, float64){
	"": func(target *boxEdges, points float64) {
		*target = boxEdges{points, points, points, points}
	},
	"x": func(target *boxEdges, points float64) { target.Left, target.Right = points, points },
	"y": func(target *boxEdges, points float64) { target.Top, target.Bottom = points, points },
	"t": func(target *boxEdges, points float64) { target.Top = points },
	"r": func(target *boxEdges, points float64) { target.Right = points },
	"b": func(target *boxEdges, points float64) { target.Bottom = points },
	"l": func(target *boxEdges, points float64) { target.Left = points },
}

func initialStyle() computedStyle {
	return computedStyle{
		Display:          displayInline,
		FontSize:         12,
		LineHeightFactor: 1.5,
	}
}

func computeElementStyle(parent computedStyle, elementName string) computedStyle {
	style := initialStyle()
	style.Alignment = parent.Alignment
	style.Weight = parent.Weight
	style.FontSize = parent.FontSize
	style.LineHeightFactor = parent.LineHeightFactor

	if elementName == "body" || elementName == "div" || elementName == "p" {
		style.Display = displayBlock
	}
	if elementName == "p" {
		style.Margin.Bottom = 12
	}

	return style
}

func applyUtilityClasses(style computedStyle, classes string) (computedStyle, error) {
	for _, class := range strings.Fields(classes) {
		if err := applyUtilityClass(&style, class); err != nil {
			return computedStyle{}, err
		}
	}
	return style, nil
}

func applyUtilityClass(style *computedStyle, class string) error {
	if applyDisplayUtility(style, class) ||
		applyTypographyUtility(style, class) ||
		applySpacingUtility(style, class) {
		return nil
	}
	return fmt.Errorf("unsupported utility class %q", class)
}

func applyDisplayUtility(style *computedStyle, class string) bool {
	switch class {
	case "block":
		style.Display = displayBlock
	case "inline":
		style.Display = displayInline
	default:
		return false
	}
	return true
}

func applyTypographyUtility(style *computedStyle, class string) bool {
	if fontSizePoints, ok := fontSizePointsByUtility[class]; ok {
		style.FontSize = fontSizePoints
		return true
	}
	if weight, ok := fontWeightByUtility[class]; ok {
		style.Weight = weight
		return true
	}
	if lineHeightFactor, ok := lineHeightFactorByUtility[class]; ok {
		style.LineHeightFactor = lineHeightFactor
		return true
	}
	if alignment, ok := textAlignmentByUtility[class]; ok {
		style.Alignment = alignment
		return true
	}
	return false
}

func applySpacingUtility(style *computedStyle, class string) bool {
	parts := strings.Split(class, "-")
	if len(parts) != 2 {
		return false
	}

	spacingPoints, ok := parseSpacingPoints(parts[1])
	if !ok {
		return false
	}

	return applySpacingPoints(style, parts[0], spacingPoints)
}

func parseSpacingPoints(rawScale string) (float64, bool) {
	scale, err := strconv.Atoi(rawScale)
	if err != nil {
		return 0, false
	}
	switch scale {
	case 0, 1, 2, 4, 6, 8:
		return float64(scale) * 3, true
	default:
		return 0, false
	}
}

func applySpacingPoints(style *computedStyle, prefix string, points float64) bool {
	target, axis := selectSpacingEdges(style, prefix)
	if target == nil || !isSupportedSpacingAxis(axis) {
		return false
	}
	setSpacingEdges(target, axis, points)
	return true
}

func isSupportedSpacingAxis(axis string) bool {
	_, supported := supportedSpacingAxes[axis]
	return supported
}

func selectSpacingEdges(style *computedStyle, prefix string) (*boxEdges, string) {
	if strings.HasPrefix(prefix, "m") {
		return &style.Margin, strings.TrimPrefix(prefix, "m")
	}
	if strings.HasPrefix(prefix, "p") {
		return &style.Padding, strings.TrimPrefix(prefix, "p")
	}
	return nil, ""
}

func setSpacingEdges(target *boxEdges, axis string, points float64) {
	spacingAxisSetters[axis](target, points)
}
