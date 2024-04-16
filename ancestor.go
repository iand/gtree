package gtree

import (
	"log/slog"
	"strings"
)

type AncestorChart struct {
	Title string
	Notes []string
	Root  *AncestorPerson
}

type AncestorPerson struct {
	ID      int
	Details []string
	Father  *AncestorPerson
	Mother  *AncestorPerson
}

type AncestorLayoutOptions struct {
	Debug bool

	LineWidth Pixel // width of any drawn lines
	Margin    Pixel // margin to add to entire drawing
	Hspace    Pixel // the horizontal space to leave between blurbs in different generations
	Vspace    Pixel // the vertical space to leave between blurbs in the same generation
	LineGap   Pixel // the distance to leave between a connecting line and any text

	HookLength Pixel // the length of the line drawn from the parent or a child to the vertical line that joins them

	TitleFontSize   Pixel // size of the font to use for the title of the chart
	TitleLineHeight Pixel // vertical distance to use for spacing the title of the chart

	NoteFontSize   Pixel // size of the font to use for the notes of the chart
	NoteLineHeight Pixel // vertical distance to use for spacing the notes of the chart

	HeadingFontSize   Pixel // size of the font to use for the first line of each blurb
	HeadingLineHeight Pixel // vertical distance to use for spacing the first line of each blurb

	DetailFontSize   Pixel // size of the font to use for the subsequent lines of each blurb
	DetailLineHeight Pixel // vertical distance to use for spacing the fsubsequent lines of each blurb
	DetailWrapWidth  Pixel // maximum width of detail text before wrapping to a new line
}

func DefaultAncestorLayoutOptions() *AncestorLayoutOptions {
	return &AncestorLayoutOptions{
		LineWidth:  2,
		Margin:     16,
		Hspace:     12,
		Vspace:     4,
		LineGap:    8,
		HookLength: 12,

		TitleFontSize:   40,
		TitleLineHeight: 42,

		NoteFontSize:   20,
		NoteLineHeight: 22,

		HeadingFontSize:   20,
		HeadingLineHeight: 22,

		DetailFontSize:   16,
		DetailLineHeight: 18,
		DetailWrapWidth:  18 * 16,
	}
}

func (ch *AncestorChart) Layout(opts *AncestorLayoutOptions) *AncestorLayout {
	if opts == nil {
		opts = DefaultAncestorLayoutOptions()
	}

	l := new(AncestorLayout)
	l.opts = *opts
	l.title = ch.Title
	l.notes = ch.Notes
	l.blurbs = make(map[int]*Blurb)

	// calculate the number of rows needed to fit all of the last generation
	l.rows = 1
	gens := ch.countGenerations(ch.Root)
	for i := 1; i < gens; i++ {
		l.rows *= 2
	}
	l.rows++ // for root person

	rootRow := l.rows/2 + 1

	if l.opts.Debug {
		slog.Info("generations", "gens", gens, "rows", l.rows, "rootRow", rootRow)
	}

	l.addPerson(ch.Root, 0, 0, nil)

	var gridHeight Pixel
	var gridWidth Pixel
	colWidths := make([]Pixel, len(l.grid))

	for col := range l.grid {
		pop := colPopulation(col)

		largestBlurbHeight := Pixel(0)
		largestBlurbWidth := Pixel(0)
		for _, b := range l.grid[col] {
			if b == nil {
				continue
			}
			if b.Height > largestBlurbHeight {
				largestBlurbHeight = b.Height
			}
			if b.Width > largestBlurbWidth {
				largestBlurbWidth = b.Width
			}
		}
		colWidths[col] = largestBlurbWidth + l.opts.Hspace

		// Give each blurb equal vertical space
		colHeight := Pixel(pop) * largestBlurbHeight

		// Add VSpace between each mother and father blurb
		if pop > 1 {
			colHeight += Pixel(pop) / 2 * l.opts.Vspace
		}

		// Add 2*VSpace between each group of mother and father pairs to separate families
		if pop > 2 {
			colHeight += (Pixel(pop)/2 - 1) * l.opts.Vspace * 2
		}

		if colHeight > gridHeight {
			gridHeight = colHeight
		}
		gridWidth += colWidths[col]
	}

	if l.opts.Debug {
		slog.Info("grid", "cols", len(l.grid), "height", gridHeight, "width", gridWidth)
		for i := range l.grid {
			slog.Info("grid rows", "col", i, "rows", len(l.grid[i]), "width", colWidths[i])
		}
	}

	// reposition blurbs

	lowestTopPos := Pixel(200000)
	x := l.opts.Margin
	// number of divisions is 2^col (col 0 has entire vertical space, col 1 splits it in two, col 2 splits in four)
	divisions := 1
	for col := range l.grid {
		spacing := gridHeight / Pixel(divisions)
		for row, b := range l.grid[col] {
			if b == nil {
				continue
			}
			b.LeftPos = x

			// centre the blurb in the division
			y0 := l.opts.Margin + spacing*Pixel(row)
			centre := y0 + spacing/2
			b.TopPos = centre - b.Height/2
			if b.TopPos < lowestTopPos {
				lowestTopPos = b.TopPos
			}
		}

		x += colWidths[col]
		divisions *= 2
	}

	l.width = gridWidth
	l.height = gridHeight

	// Shift everything up to remove any empty space at top

	if lowestTopPos > 0 {
		l.height -= lowestTopPos
		for col := range l.grid {
			for _, b := range l.grid[col] {
				if b == nil {
					continue
				}
				b.TopPos -= lowestTopPos
			}
		}
	}

	// Shift everything down to accomodate title
	titleHeight, _ := l.titleDimensions()
	l.height += titleHeight + l.opts.Vspace*4
	for col := range l.grid {
		for _, b := range l.grid[col] {
			if b == nil {
				continue
			}
			b.TopPos += titleHeight + l.opts.Vspace*4
		}
	}

	// calculate connectors
	for col := range l.grid {
		if col == 0 {
			continue
		}
		for row, b := range l.grid[col] {
			if b == nil {
				continue
			}
			var childIdx int
			if row%2 == 0 {
				// male
				childIdx = row / 2
			} else {
				childIdx = (row - 1) / 2
			}
			childBlurb := l.grid[col-1][childIdx]

			// draw hook projecting from left edge of parent
			l.connectors = append(l.connectors, &Connector{
				Points: []Point{
					// Start just to left of blurb
					{X: b.LeftPos - l.opts.LineGap, Y: b.SideHookY()},

					// Move left by HookLength
					{X: b.LeftPos - l.opts.LineGap - l.opts.HookLength, Y: b.SideHookY()},

					// Move vertically to hook of child
					{X: b.LeftPos - l.opts.LineGap - l.opts.HookLength, Y: childBlurb.SideHookY()},

					// Move left by HookLength
					{X: b.LeftPos - l.opts.LineGap - l.opts.HookLength - l.opts.Hspace, Y: childBlurb.SideHookY()},
				},
			})

		}
	}
	return l
}

func (ch *AncestorChart) countGenerations(p *AncestorPerson) int {
	if p.Father == nil && p.Mother == nil {
		return 1
	}

	var g int
	if p.Father != nil {
		g = ch.countGenerations(p.Father)
	}
	if p.Mother != nil {
		m := ch.countGenerations(p.Mother)
		if m > g {
			g = m
		}
	}

	return 1 + g
}

type AncestorLayout struct {
	opts       AncestorLayoutOptions
	width      Pixel
	height     Pixel
	title      string
	notes      []string
	blurbs     map[int]*Blurb
	grid       [][]*Blurb // col, row
	rows       int
	connectors []*Connector
}

func (l *AncestorLayout) Width() Pixel { return l.width }

func (l *AncestorLayout) Height() Pixel { return l.height }

func (l *AncestorLayout) Margin() Pixel { return l.opts.Margin }

func (l *AncestorLayout) Title() TextElement {
	return TextElement{
		Text:       l.title,
		FontSize:   l.opts.TitleFontSize,
		LineHeight: l.opts.TitleLineHeight,
	}
}

func (l *AncestorLayout) Notes() []TextElement {
	tes := make([]TextElement, len(l.notes))

	for i := range l.notes {
		tes[i] = TextElement{
			Text:       l.notes[i],
			FontSize:   l.opts.NoteFontSize,
			LineHeight: l.opts.NoteLineHeight,
		}
	}
	return tes
}

func (l *AncestorLayout) Blurbs() []*Blurb {
	bs := make([]*Blurb, 0, len(l.blurbs))
	for _, b := range l.blurbs {
		bs = append(bs, b)
	}
	return bs
}

func (l *AncestorLayout) Connectors() []*Connector {
	return l.connectors
}

func (l *AncestorLayout) Debug() bool { return l.opts.Debug }

func (l *AncestorLayout) addPerson(p *AncestorPerson, col int, row int, child *Blurb) *Blurb {
	b := l.newBlurb(p.ID, p.Details, col, row, child)

	for len(l.grid) <= col {
		l.grid = append(l.grid, make([]*Blurb, colPopulation(len(l.grid)+1)))
	}

	l.grid[col][row] = b

	// father goes on next column, previous row
	if p.Father != nil {
		l.addPerson(p.Father, col+1, (row * 2), b)
	}

	// mother goes on next column, next row
	if p.Mother != nil {
		l.addPerson(p.Mother, col+1, (row*2)+1, b)
	}

	return b
}

func (l *AncestorLayout) newBlurb(id int, texts []string, col int, row int, child *Blurb) *Blurb {
	// texts = l.wrapTexts(texts)
	b := &Blurb{
		ID:                  id,
		Col:                 col,
		Row:                 col,
		AbsolutePositioning: true,
		// Parent: parent,
		// TopHookOffset:     l.opts.Hspace * 2,
		SideHookOffset:    (l.opts.HeadingLineHeight * 2) / 3,
		LeftNeighbour:     child,
		HeadingFontSize:   l.opts.HeadingFontSize,
		DetailFontSize:    l.opts.DetailFontSize,
		HeadingLineHeight: l.opts.HeadingLineHeight,
		DetailLineHeight:  l.opts.DetailLineHeight,
	}

	if len(texts) > 0 {
		b.HeadingText = texts[0]
		b.Height = b.HeadingLineHeight
		b.Width = textWidth([]rune(b.HeadingText), b.HeadingFontSize)

		if len(texts) > 1 {

			b.DetailTexts = l.wrapDetailTexts(texts[1:])
			b.Height += b.DetailLineHeight * Pixel(len(b.DetailTexts))

			for i := range b.DetailTexts {
				wl := textWidth([]rune(b.DetailTexts[i]), b.DetailFontSize)
				if wl > b.Width {
					b.Width = wl
				}
			}
		}

	}

	l.blurbs[id] = b

	return b
}

func (l *AncestorLayout) titleDimensions() (Pixel, Pixel) {
	if l.title == "" && len(l.notes) == 0 {
		return 0, 0
	}

	var h, w Pixel

	if l.title != "" {
		h += l.opts.TitleLineHeight
		w = textWidth([]rune(l.title), l.opts.TitleFontSize)
	}

	if len(l.notes) != 0 {
		h += l.opts.NoteLineHeight * Pixel(len(l.notes))
		for i := 0; i < len(l.notes); i++ {
			wl := textWidth([]rune(l.notes[i]), l.opts.NoteFontSize)
			if wl > w {
				w = wl
			}
		}
	}

	return h, w
}

func (l *AncestorLayout) wrapDetailTexts(texts []string) []string {
	if len(texts) == 0 {
		return []string{}
	}
	wrapped := make([]string, 0, len(texts))
	for i := 0; i < len(texts); i++ {
		wl := textWidth([]rune(texts[i]), l.opts.DetailFontSize)
		if wl <= l.opts.DetailWrapWidth {
			wrapped = append(wrapped, texts[i])
			continue
		}

		words := strings.Fields(texts[i])
		if len(words) == 0 {
			wrapped = append(wrapped, "")
			continue
		}

		// wrap the liine
		var line string
		for w := 0; w < len(words); w++ {
			candidate := line
			if len(line) != 0 {
				candidate += " "
			}
			candidate += words[w]
			wl := textWidth([]rune(candidate), l.opts.DetailFontSize)
			if wl >= l.opts.DetailWrapWidth {
				if len(line) == 0 {
					// this is a single long word
					wrapped = append(wrapped, candidate)
					line = ""
				} else {
					// add current line and start a new one
					wrapped = append(wrapped, line)
					line = words[w]
				}
				continue
			}

			line = candidate
		}
		wrapped = append(wrapped, line)
	}

	return wrapped
}

// colPopulation returns the expected population of each column
func colPopulation(col int) int {
	return 1 << col
}
