package inked

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	letterWidth  = 612.0
	letterHeight = 792.0
	pageInset    = 72.0
)

type laidOutDocument struct {
	Pages []laidOutPage
}

type laidOutPage struct {
	Runs []textRun
}

type textRun struct {
	X     float64
	Y     float64
	Text  string
	Style computedStyle
}

type inlineToken interface {
	inlineToken()
}

type wordToken struct {
	Text  string
	Style computedStyle
}

func (wordToken) inlineToken() {}

type spaceToken struct {
	Style computedStyle
}

func (spaceToken) inlineToken() {}

type breakToken struct{}

func (breakToken) inlineToken() {}

type spacerToken struct {
	Width float64
}

func (spacerToken) inlineToken() {}

type linePart struct {
	Text  string
	Style computedStyle
	Width float64
}

type lineBox struct {
	Parts  []linePart
	Width  float64
	Height float64
}

type layoutEngine struct {
	document laidOutDocument
	y        float64
}

func layoutDocument(document resolvedDocument) (laidOutDocument, error) {
	engine := layoutEngine{}
	engine.newPage()
	if err := engine.layoutContainer(document.Children, pageInset, letterWidth-2*pageInset, alignLeft); err != nil {
		return laidOutDocument{}, err
	}
	return engine.document, nil
}

func (engine *layoutEngine) newPage() {
	engine.document.Pages = append(engine.document.Pages, laidOutPage{})
	engine.y = pageInset
}

func (engine *layoutEngine) layoutContainer(
	nodes []resolvedNode,
	x float64,
	width float64,
	alignment textAlignment,
) error {
	var inline []resolvedNode
	for _, node := range nodes {
		element, block := blockElement(node)
		if block {
			if err := engine.layoutInline(inline, x, width, alignment); err != nil {
				return err
			}
			inline = nil
			if err := engine.layoutBlock(element, x, width); err != nil {
				return err
			}
			continue
		}
		inline = append(inline, node)
	}
	return engine.layoutInline(inline, x, width, alignment)
}

func blockElement(node resolvedNode) (resolvedElement, bool) {
	element, ok := node.(resolvedElement)
	return element, ok && element.Style.Display == displayBlock
}

func (engine *layoutEngine) layoutBlock(element resolvedElement, x, width float64) error {
	style := element.Style
	engine.y += style.Margin.Top + style.Padding.Top

	contentX := x + style.Margin.Left + style.Padding.Left
	contentWidth := width -
		style.Margin.Left - style.Margin.Right -
		style.Padding.Left - style.Padding.Right
	if contentWidth < 1 {
		contentWidth = 1
	}

	if err := engine.layoutContainer(element.Children, contentX, contentWidth, style.Alignment); err != nil {
		return err
	}
	engine.y += style.Padding.Bottom + style.Margin.Bottom
	return nil
}

func (engine *layoutEngine) layoutInline(
	nodes []resolvedNode,
	x float64,
	width float64,
	alignment textAlignment,
) error {
	tokens := collectInlineTokens(nodes)
	line := lineBox{}
	for _, token := range tokens {
		var err error
		line, err = engine.placeToken(line, token, x, width, alignment)
		if err != nil {
			return err
		}
	}
	return engine.emitLine(line, x, width, alignment)
}

func (engine *layoutEngine) placeToken(
	line lineBox,
	token inlineToken,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	switch value := token.(type) {
	case breakToken:
		return lineBox{}, engine.emitLine(line, x, width, alignment)
	case spacerToken:
		return engine.placePart(line, linePart{Width: value.Width}, x, width, alignment)
	case spaceToken:
		return engine.placeSpace(line, value, x, width, alignment)
	case wordToken:
		return engine.placeWord(line, value, x, width, alignment)
	default:
		return line, nil
	}
}

func (engine *layoutEngine) placeSpace(
	line lineBox,
	token spaceToken,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	if len(line.Parts) == 0 {
		return line, nil
	}
	part, err := measuredPart(" ", token.Style)
	if err != nil {
		return line, err
	}
	return engine.placePart(line, part, x, width, alignment)
}

func (engine *layoutEngine) placeWord(
	line lineBox,
	token wordToken,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	parts, err := splitWord(token, width)
	if err != nil {
		return line, err
	}
	for _, part := range parts {
		line, err = engine.placePart(line, part, x, width, alignment)
		if err != nil {
			return line, err
		}
	}
	return line, nil
}

func (engine *layoutEngine) placePart(
	line lineBox,
	part linePart,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	if line.Width+part.Width <= width || len(line.Parts) == 0 {
		return appendPart(line, part), nil
	}
	if err := engine.emitLine(line, x, width, alignment); err != nil {
		return line, err
	}
	if part.Text == " " {
		return lineBox{}, nil
	}
	return appendPart(lineBox{}, part), nil
}

func appendPart(line lineBox, part linePart) lineBox {
	line.Parts = append(line.Parts, part)
	line.Width += part.Width
	height := part.Style.FontSize * part.Style.LineHeightFactor
	if part.Text != "" && height > line.Height {
		line.Height = height
	}
	return line
}

func measuredPart(text string, style computedStyle) (linePart, error) {
	width, err := measureText(text, style)
	if err != nil {
		return linePart{}, err
	}
	return linePart{Text: text, Style: style, Width: width}, nil
}

func splitWord(token wordToken, width float64) ([]linePart, error) {
	part, err := measuredPart(token.Text, token.Style)
	if err != nil {
		return nil, err
	}
	if part.Width <= width {
		return []linePart{part}, nil
	}
	return splitLongWord(token, width)
}

func splitLongWord(token wordToken, width float64) ([]linePart, error) {
	var result []linePart
	remaining := token.Text
	for remaining != "" {
		prefix, rest, err := fittingPrefix(remaining, token.Style, width)
		if err != nil {
			return nil, err
		}
		part, err := measuredPart(prefix, token.Style)
		if err != nil {
			return nil, err
		}
		result = append(result, part)
		remaining = rest
	}
	return result, nil
}

func fittingPrefix(text string, style computedStyle, width float64) (string, string, error) {
	cut := 0
	for index, character := range text {
		candidateEnd := index + utf8.RuneLen(character)
		candidateWidth, err := measureText(text[:candidateEnd], style)
		if err != nil {
			return "", "", err
		}
		if candidateWidth > width && cut > 0 {
			break
		}
		cut = candidateEnd
	}
	return text[:cut], text[cut:], nil
}

func (engine *layoutEngine) emitLine(
	line lineBox,
	x float64,
	width float64,
	alignment textAlignment,
) error {
	if len(line.Parts) == 0 {
		return nil
	}
	engine.ensureSpace(line.Height)
	currentX := alignedX(x, width, line.Width, alignment)
	page := &engine.document.Pages[len(engine.document.Pages)-1]
	for _, part := range line.Parts {
		if part.Text != "" {
			page.Runs = append(page.Runs, textRun{
				X: currentX, Y: engine.y + part.Style.FontSize,
				Text: part.Text, Style: part.Style,
			})
		}
		currentX += part.Width
	}
	engine.y += line.Height
	return nil
}

func (engine *layoutEngine) ensureSpace(height float64) {
	if engine.y+height <= letterHeight-pageInset {
		return
	}
	engine.newPage()
}

func alignedX(x, width, lineWidth float64, alignment textAlignment) float64 {
	switch alignment {
	case alignCenter:
		return x + (width-lineWidth)/2
	case alignRight:
		return x + width - lineWidth
	default:
		return x
	}
}

type tokenCollector struct {
	tokens       []inlineToken
	hasText      bool
	pendingSpace bool
}

func collectInlineTokens(nodes []resolvedNode) []inlineToken {
	collector := tokenCollector{}
	for _, node := range nodes {
		collector.collect(node)
	}
	return collector.tokens
}

func (collector *tokenCollector) collect(node resolvedNode) {
	switch value := node.(type) {
	case resolvedText:
		collector.collectText(value)
	case resolvedBreak:
		collector.tokens = append(collector.tokens, breakToken{})
		collector.pendingSpace = false
	case resolvedElement:
		collector.collectElement(value)
	}
}

func (collector *tokenCollector) collectElement(element resolvedElement) {
	horizontal := element.Style.Margin.Left + element.Style.Padding.Left
	if horizontal > 0 {
		collector.tokens = append(collector.tokens, spacerToken{Width: horizontal})
	}
	for _, child := range element.Children {
		collector.collect(child)
	}
	horizontal = element.Style.Padding.Right + element.Style.Margin.Right
	if horizontal > 0 {
		collector.tokens = append(collector.tokens, spacerToken{Width: horizontal})
	}
}

func (collector *tokenCollector) collectText(text resolvedText) {
	fields := strings.Fields(text.Text)
	if len(fields) == 0 {
		collector.collectWhitespace(text.Text)
		return
	}

	collector.appendLeadingSpace(text)
	collector.appendFields(fields, text.Style)
	collector.pendingSpace = endsWithSpace(text.Text)
}

func (collector *tokenCollector) collectWhitespace(text string) {
	collector.pendingSpace = collector.pendingSpace || textHasSpace(text)
}

func (collector *tokenCollector) appendLeadingSpace(text resolvedText) {
	if collector.hasText && (collector.pendingSpace || startsWithSpace(text.Text)) {
		collector.tokens = append(collector.tokens, spaceToken{Style: text.Style})
	}
}

func (collector *tokenCollector) appendFields(fields []string, style computedStyle) {
	for index, field := range fields {
		if index > 0 {
			collector.tokens = append(collector.tokens, spaceToken{Style: style})
		}
		collector.tokens = append(collector.tokens, wordToken{Text: field, Style: style})
		collector.hasText = true
	}
}

func startsWithSpace(text string) bool {
	character, _ := utf8.DecodeRuneInString(text)
	return unicode.IsSpace(character)
}

func endsWithSpace(text string) bool {
	character, _ := utf8.DecodeLastRuneInString(text)
	return unicode.IsSpace(character)
}

func textHasSpace(text string) bool {
	return strings.IndexFunc(text, unicode.IsSpace) >= 0
}
