package inked

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	letterPageWidth   = 612.0
	letterPageHeight  = 792.0
	defaultPageMargin = 72.0
)

type documentLayout struct {
	Pages []pageLayout
}

type pageLayout struct {
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

type whitespaceToken struct {
	Style computedStyle
}

func (whitespaceToken) inlineToken() {}

type lineBreakToken struct{}

func (lineBreakToken) inlineToken() {}

type horizontalSpacingToken struct {
	Width float64
}

func (horizontalSpacingToken) inlineToken() {}

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

type layoutState struct {
	layout  documentLayout
	cursorY float64
}

func layoutDocument(document resolvedDocument) (documentLayout, error) {
	state := layoutState{}
	state.startPage()
	if err := state.layoutNodes(
		document.Children,
		defaultPageMargin,
		letterPageWidth-2*defaultPageMargin,
		alignLeft,
	); err != nil {
		return documentLayout{}, err
	}
	return state.layout, nil
}

func (state *layoutState) startPage() {
	state.layout.Pages = append(state.layout.Pages, pageLayout{})
	state.cursorY = defaultPageMargin
}

func (state *layoutState) layoutNodes(
	nodes []resolvedNode,
	x float64,
	width float64,
	alignment textAlignment,
) error {
	var inlineNodes []resolvedNode
	for _, node := range nodes {
		element, isBlock := asBlockElement(node)
		if isBlock {
			if err := state.layoutInlineNodes(inlineNodes, x, width, alignment); err != nil {
				return err
			}
			inlineNodes = nil
			if err := state.layoutBlockElement(element, x, width); err != nil {
				return err
			}
			continue
		}
		inlineNodes = append(inlineNodes, node)
	}
	return state.layoutInlineNodes(inlineNodes, x, width, alignment)
}

func asBlockElement(node resolvedNode) (resolvedElement, bool) {
	element, ok := node.(resolvedElement)
	return element, ok && element.Style.Display == displayBlock
}

func (state *layoutState) layoutBlockElement(element resolvedElement, x, width float64) error {
	style := element.Style
	state.cursorY += style.Margin.Top + style.Padding.Top

	contentX := x + style.Margin.Left + style.Padding.Left
	contentWidth := width -
		style.Margin.Left - style.Margin.Right -
		style.Padding.Left - style.Padding.Right
	if contentWidth < 1 {
		contentWidth = 1
	}

	if err := state.layoutNodes(element.Children, contentX, contentWidth, style.Alignment); err != nil {
		return err
	}
	state.cursorY += style.Padding.Bottom + style.Margin.Bottom
	return nil
}

func (state *layoutState) layoutInlineNodes(
	nodes []resolvedNode,
	x float64,
	width float64,
	alignment textAlignment,
) error {
	tokens := collectInlineTokens(nodes)
	line := lineBox{}
	for _, token := range tokens {
		var err error
		line, err = state.appendTokenToLine(line, token, x, width, alignment)
		if err != nil {
			return err
		}
	}
	return state.placeLine(line, x, width, alignment)
}

func (state *layoutState) appendTokenToLine(
	line lineBox,
	token inlineToken,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	switch value := token.(type) {
	case lineBreakToken:
		return lineBox{}, state.placeLine(line, x, width, alignment)
	case horizontalSpacingToken:
		return state.appendPartWithWrap(line, linePart{Width: value.Width}, x, width, alignment)
	case whitespaceToken:
		return state.appendWhitespaceToLine(line, value, x, width, alignment)
	case wordToken:
		return state.appendWordToLine(line, value, x, width, alignment)
	default:
		return line, nil
	}
}

func (state *layoutState) appendWhitespaceToLine(
	line lineBox,
	token whitespaceToken,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	if len(line.Parts) == 0 {
		return line, nil
	}
	part, err := measureLinePart(" ", token.Style)
	if err != nil {
		return line, err
	}
	return state.appendPartWithWrap(line, part, x, width, alignment)
}

func (state *layoutState) appendWordToLine(
	line lineBox,
	token wordToken,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	parts, err := measureWordParts(token, width)
	if err != nil {
		return line, err
	}
	for _, part := range parts {
		line, err = state.appendPartWithWrap(line, part, x, width, alignment)
		if err != nil {
			return line, err
		}
	}
	return line, nil
}

func (state *layoutState) appendPartWithWrap(
	line lineBox,
	part linePart,
	x float64,
	width float64,
	alignment textAlignment,
) (lineBox, error) {
	if line.Width+part.Width <= width || len(line.Parts) == 0 {
		return appendPartToLine(line, part), nil
	}
	if err := state.placeLine(line, x, width, alignment); err != nil {
		return line, err
	}
	if part.Text == " " {
		return lineBox{}, nil
	}
	return appendPartToLine(lineBox{}, part), nil
}

func appendPartToLine(line lineBox, part linePart) lineBox {
	line.Parts = append(line.Parts, part)
	line.Width += part.Width
	height := part.Style.FontSize * part.Style.LineHeightFactor
	if part.Text != "" && height > line.Height {
		line.Height = height
	}
	return line
}

func measureLinePart(text string, style computedStyle) (linePart, error) {
	width, err := measureTextWidth(text, style)
	if err != nil {
		return linePart{}, err
	}
	return linePart{Text: text, Style: style, Width: width}, nil
}

func measureWordParts(token wordToken, width float64) ([]linePart, error) {
	part, err := measureLinePart(token.Text, token.Style)
	if err != nil {
		return nil, err
	}
	if part.Width <= width {
		return []linePart{part}, nil
	}
	return splitLongWordIntoParts(token, width)
}

func splitLongWordIntoParts(token wordToken, width float64) ([]linePart, error) {
	var parts []linePart
	remainingText := token.Text
	for remainingText != "" {
		fittingText, remainingSuffix, err := fitTextPrefix(
			remainingText,
			token.Style,
			width,
		)
		if err != nil {
			return nil, err
		}
		part, err := measureLinePart(fittingText, token.Style)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
		remainingText = remainingSuffix
	}
	return parts, nil
}

func fitTextPrefix(text string, style computedStyle, width float64) (string, string, error) {
	splitIndex := 0
	for index, character := range text {
		candidateEnd := index + utf8.RuneLen(character)
		candidateWidth, err := measureTextWidth(text[:candidateEnd], style)
		if err != nil {
			return "", "", err
		}
		if candidateWidth > width && splitIndex > 0 {
			break
		}
		splitIndex = candidateEnd
	}
	return text[:splitIndex], text[splitIndex:], nil
}

func (state *layoutState) placeLine(
	line lineBox,
	x float64,
	width float64,
	alignment textAlignment,
) error {
	if len(line.Parts) == 0 {
		return nil
	}
	state.ensurePageCapacity(line.Height)
	currentX := alignedLineX(x, width, line.Width, alignment)
	page := &state.layout.Pages[len(state.layout.Pages)-1]
	for _, part := range line.Parts {
		if part.Text != "" {
			page.Runs = append(page.Runs, textRun{
				X: currentX, Y: state.cursorY + part.Style.FontSize,
				Text: part.Text, Style: part.Style,
			})
		}
		currentX += part.Width
	}
	state.cursorY += line.Height
	return nil
}

func (state *layoutState) ensurePageCapacity(height float64) {
	if state.cursorY+height <= letterPageHeight-defaultPageMargin {
		return
	}
	state.startPage()
}

func alignedLineX(x, width, lineWidth float64, alignment textAlignment) float64 {
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
	tokens               []inlineToken
	hasText              bool
	hasPendingWhitespace bool
}

func collectInlineTokens(nodes []resolvedNode) []inlineToken {
	collector := tokenCollector{}
	for _, node := range nodes {
		collector.collectNode(node)
	}
	return collector.tokens
}

func (collector *tokenCollector) collectNode(node resolvedNode) {
	switch value := node.(type) {
	case resolvedText:
		collector.collectTextTokens(value)
	case resolvedLineBreak:
		collector.tokens = append(collector.tokens, lineBreakToken{})
		collector.hasPendingWhitespace = false
	case resolvedElement:
		collector.collectElementTokens(value)
	}
}

func (collector *tokenCollector) collectElementTokens(element resolvedElement) {
	leadingSpacing := element.Style.Margin.Left + element.Style.Padding.Left
	if leadingSpacing > 0 {
		collector.tokens = append(
			collector.tokens,
			horizontalSpacingToken{Width: leadingSpacing},
		)
	}
	for _, child := range element.Children {
		collector.collectNode(child)
	}
	trailingSpacing := element.Style.Padding.Right + element.Style.Margin.Right
	if trailingSpacing > 0 {
		collector.tokens = append(
			collector.tokens,
			horizontalSpacingToken{Width: trailingSpacing},
		)
	}
}

func (collector *tokenCollector) collectTextTokens(text resolvedText) {
	fields := strings.Fields(text.Text)
	if len(fields) == 0 {
		collector.recordWhitespace(text.Text)
		return
	}

	collector.appendLeadingWhitespace(text)
	collector.appendWords(fields, text.Style)
	collector.hasPendingWhitespace = endsWithWhitespace(text.Text)
}

func (collector *tokenCollector) recordWhitespace(text string) {
	collector.hasPendingWhitespace =
		collector.hasPendingWhitespace || containsWhitespace(text)
}

func (collector *tokenCollector) appendLeadingWhitespace(text resolvedText) {
	if collector.hasText &&
		(collector.hasPendingWhitespace || startsWithWhitespace(text.Text)) {
		collector.tokens = append(collector.tokens, whitespaceToken{Style: text.Style})
	}
}

func (collector *tokenCollector) appendWords(words []string, style computedStyle) {
	for index, word := range words {
		if index > 0 {
			collector.tokens = append(collector.tokens, whitespaceToken{Style: style})
		}
		collector.tokens = append(collector.tokens, wordToken{Text: word, Style: style})
		collector.hasText = true
	}
}

func startsWithWhitespace(text string) bool {
	character, _ := utf8.DecodeRuneInString(text)
	return unicode.IsSpace(character)
}

func endsWithWhitespace(text string) bool {
	character, _ := utf8.DecodeLastRuneInString(text)
	return unicode.IsSpace(character)
}

func containsWhitespace(text string) bool {
	return strings.IndexFunc(text, unicode.IsSpace) >= 0
}
