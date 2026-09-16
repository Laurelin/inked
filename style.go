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

type edges struct {
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
	Margin           edges
	Padding          edges
}

var fontSizeUtilities = map[string]float64{
	"text-xs": 9, "text-sm": 10.5, "text-base": 12, "text-lg": 13.5, "text-xl": 15,
}

var fontWeightUtilities = map[string]fontWeight{
	"font-normal": weightNormal, "font-bold": weightBold,
}

var lineHeightUtilities = map[string]float64{
	"leading-none": 1, "leading-tight": 1.25,
	"leading-normal": 1.5, "leading-relaxed": 1.625,
}

var alignmentUtilities = map[string]textAlignment{
	"text-left": alignLeft, "text-center": alignCenter, "text-right": alignRight,
}

var validSpacingAxes = map[string]bool{
	"": true, "x": true, "y": true, "t": true, "r": true, "b": true, "l": true,
}

var spacingSetters = map[string]func(*edges, float64){
	"":  func(target *edges, value float64) { *target = edges{value, value, value, value} },
	"x": func(target *edges, value float64) { target.Left, target.Right = value, value },
	"y": func(target *edges, value float64) { target.Top, target.Bottom = value, value },
	"t": func(target *edges, value float64) { target.Top = value },
	"r": func(target *edges, value float64) { target.Right = value },
	"b": func(target *edges, value float64) { target.Bottom = value },
	"l": func(target *edges, value float64) { target.Left = value },
}

func defaultStyle() computedStyle {
	return computedStyle{
		Display:          displayInline,
		FontSize:         12,
		LineHeightFactor: 1.5,
	}
}

func elementStyle(parent computedStyle, element string) computedStyle {
	style := defaultStyle()
	style.Alignment = parent.Alignment
	style.Weight = parent.Weight
	style.FontSize = parent.FontSize
	style.LineHeightFactor = parent.LineHeightFactor

	if element == "body" || element == "div" || element == "p" {
		style.Display = displayBlock
	}
	if element == "p" {
		style.Margin.Bottom = 12
	}

	return style
}

func applyClasses(style computedStyle, classes string) (computedStyle, error) {
	for _, class := range strings.Fields(classes) {
		if err := applyClass(&style, class); err != nil {
			return computedStyle{}, err
		}
	}
	return style, nil
}

func applyClass(style *computedStyle, class string) error {
	if applyDisplay(style, class) ||
		applyTypography(style, class) ||
		applySpacing(style, class) {
		return nil
	}
	return fmt.Errorf("unsupported utility class %q", class)
}

func applyDisplay(style *computedStyle, class string) bool {
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

func applyTypography(style *computedStyle, class string) bool {
	if value, ok := fontSizeUtilities[class]; ok {
		style.FontSize = value
		return true
	}
	if value, ok := fontWeightUtilities[class]; ok {
		style.Weight = value
		return true
	}
	if value, ok := lineHeightUtilities[class]; ok {
		style.LineHeightFactor = value
		return true
	}
	if value, ok := alignmentUtilities[class]; ok {
		style.Alignment = value
		return true
	}
	return false
}

func applySpacing(style *computedStyle, class string) bool {
	parts := strings.Split(class, "-")
	if len(parts) != 2 {
		return false
	}

	value, ok := spacingValue(parts[1])
	if !ok {
		return false
	}

	return setSpacing(style, parts[0], value)
}

func spacingValue(raw string) (float64, bool) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	switch value {
	case 0, 1, 2, 4, 6, 8:
		return float64(value) * 3, true
	default:
		return 0, false
	}
}

func setSpacing(style *computedStyle, prefix string, value float64) bool {
	target, axis := spacingTarget(style, prefix)
	if target == nil || !validSpacingAxis(axis) {
		return false
	}
	applyEdges(target, axis, value)
	return true
}

func validSpacingAxis(axis string) bool {
	return validSpacingAxes[axis]
}

func spacingTarget(style *computedStyle, prefix string) (*edges, string) {
	if strings.HasPrefix(prefix, "m") {
		return &style.Margin, strings.TrimPrefix(prefix, "m")
	}
	if strings.HasPrefix(prefix, "p") {
		return &style.Padding, strings.TrimPrefix(prefix, "p")
	}
	return nil, ""
}

func applyEdges(target *edges, axis string, value float64) {
	spacingSetters[axis](target, value)
}
