package gtree

import (
	"log/slog"
)

// AncestorChart represents a horizontal ancestor chart, where the root person is
// positioned on the left, and successive generations extend to the right. Ancestors in
// each generation are aligned vertically, visually depicting the lineage from the root
// person to their ancestors.
type AncestorChart struct {
	Title string
	Notes []string
	Root  *AncestorPerson
}

// AncestorPerson represents an individual in the ancestor chart, including their ID, details, and their parents.
type AncestorPerson struct {
	ID      int
	Details []string
	Father  *AncestorPerson
	Mother  *AncestorPerson
}

// AncestorLayoutOptions defines various layout parameters for rendering the ancestor chart.
type AncestorLayoutOptions struct {
	Debug bool

	LineWidth Pixel // width of any drawn lines
	Margin    Pixel // margin to add to entire drawing
	Hspace    Pixel // the horizontal space to leave between blurbs in different generations
	Vspace    Pixel // the vertical space to leave between blurbs in the same generation
	LineGap   Pixel // the distance to leave between a connecting line and any text

	HookLength Pixel // the length of the line drawn from the parent or a child to the vertical line that joins them

	TitleStyle   TextStyle // TitleStyle is the style of the font to use for the title of the chart.
	NoteStyle    TextStyle // NoteStyle is the style of the font to use for the notes of the chart.
	HeadingStyle TextStyle // HeadingStyle is the style of the font to use for the first line of each blurb.
	DetailStyle  TextStyle // DetailStyle is the style of the font to use for the subsequent lines of each blurb after the first.

	DetailWrapWidth Pixel // DetailWrapWidth is the maximum width of detail text before wrapping to a new line.
}

// DefaultAncestorLayoutOptions returns the default layout options for rendering the ancestor chart.
func DefaultAncestorLayoutOptions() *AncestorLayoutOptions {
	return &AncestorLayoutOptions{
		LineWidth:  2,
		Margin:     16,
		Hspace:     12,
		Vspace:     4,
		LineGap:    8,
		HookLength: 12,

		TitleStyle: TextStyle{
			FontSize:   40,
			LineHeight: 42,
		},
		NoteStyle: TextStyle{
			FontSize:   20,
			LineHeight: 22,
		},
		HeadingStyle: TextStyle{
			FontSize:   20,
			LineHeight: 22,
		},
		DetailStyle: TextStyle{
			FontSize:   16,
			LineHeight: 18,
		},

		DetailWrapWidth: 18 * 16,
	}
}

// Layout generates the layout for the ancestor chart based on the provided options.
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
	titleHeight, _ := titleDimensions(l.title, l.notes, l.opts.TitleStyle, l.opts.NoteStyle)

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

// countGenerations counts the number of generations from the root person in the ancestor chart.
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

// AncestorLayout represents the layout of an ancestor chart, including dimensions and layout options.
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

// Width returns the width of the layout.
func (l *AncestorLayout) Width() Pixel { return l.width }

// Height returns the height of the layout.
func (l *AncestorLayout) Height() Pixel { return l.height }

// Margin returns the margin of the layout.
func (l *AncestorLayout) Margin() Pixel { return l.opts.Margin }

// Title returns the title element of the layout.
func (l *AncestorLayout) Title() TextElement {
	return TextElement{
		Text:  l.title,
		Style: l.opts.TitleStyle,
	}
}

// Notes returns the notes elements of the layout.
func (l *AncestorLayout) Notes() []TextElement {
	tes := make([]TextElement, len(l.notes))

	for i := range l.notes {
		tes[i] = TextElement{
			Text:  l.notes[i],
			Style: l.opts.NoteStyle,
		}
	}
	return tes
}

// Blurbs returns all the blurbs in the layout.
func (l *AncestorLayout) Blurbs() []*Blurb {
	bs := make([]*Blurb, 0, len(l.blurbs))
	for _, b := range l.blurbs {
		bs = append(bs, b)
	}
	return bs
}

// Connectors returns all the connectors in the layout.
func (l *AncestorLayout) Connectors() []*Connector {
	return l.connectors
}

// Debug reports whether the layout is in debug mode.
func (l *AncestorLayout) Debug() bool { return l.opts.Debug }

// addPerson adds a person and their parents to the layout at the specified column and row.
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

// newBlurb creates a new blurb for the given person at the specified column and row.
func (l *AncestorLayout) newBlurb(id int, texts []string, col int, row int, child *Blurb) *Blurb {
	// texts = l.wrapTexts(texts)
	b := &Blurb{
		ID:                  id,
		Col:                 col,
		Row:                 col,
		AbsolutePositioning: true,
		// Parent: parent,
		// TopHookOffset:     l.opts.Hspace * 2,
		SideHookOffset: (l.opts.HeadingStyle.LineHeight * 2) / 3,
		LeftNeighbour:  child,
		HeadingStyle:   l.opts.HeadingStyle,
		DetailStyle:    l.opts.DetailStyle,
	}

	if len(texts) > 0 {
		b.HeadingTexts = append(b.HeadingTexts, texts[0])
		b.Height = b.HeadingStyle.LineHeight
		b.Width = textWidth([]rune(b.HeadingTexts[0]), b.HeadingStyle.FontSize)

		if len(texts) > 1 {

			b.DetailTexts = wrapText(texts[1:], l.opts.DetailWrapWidth, l.opts.DetailStyle.FontSize)
			b.Height += b.DetailStyle.LineHeight * Pixel(len(b.DetailTexts))

			for i := range b.DetailTexts {
				wl := textWidth([]rune(b.DetailTexts[i]), b.DetailStyle.FontSize)
				if wl > b.Width {
					b.Width = wl
				}
			}
		}

	}

	l.blurbs[id] = b

	return b
}

// colPopulation returns the expected population of each column
func colPopulation(col int) int {
	return 1 << col
}
