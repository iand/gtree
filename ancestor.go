package gtree

import (
	"fmt"
	"log/slog"
	"strings"
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
	ID       int
	Headings []string
	Details  []string
	Father   *AncestorPerson
	Mother   *AncestorPerson
}

// AncestorLayoutOptions defines various layout parameters for rendering the ancestor chart.
type AncestorLayoutOptions struct {
	Debug bool

	LineWidth Pixel // width of any drawn lines
	Margin    Pixel // margin to add to entire drawing
	Hspace    Pixel // the horizontal space to leave between nodes in different generations
	Vspace    Pixel // the vertical space to leave between nodes in the same generation
	LineGap   Pixel // the distance to leave between a connecting line and any text

	HookLength Pixel // the length of the line drawn from the parent or a child to the vertical line that joins them

	TitleStyle   TextStyleOption // TitleStyle is the style of the font to use for the title of the chart.
	NoteStyle    TextStyleOption // NoteStyle is the style of the font to use for the notes of the chart.
	HeadingStyle TextStyleOption // HeadingStyle is the style of the font to use for the first line of each node.
	DetailStyle  TextStyleOption // DetailStyle is the style of the font to use for the subsequent lines of each node after the first.

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

		TitleStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   40,
			LineHeight: 42,
			Color:      "#000",
		},
		NoteStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   24,
			LineHeight: 26,
			Color:      "#000",
		},
		HeadingStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   28,
			LineHeight: 30,
			Color:      "#000",
		},
		DetailStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   16,
			LineHeight: 18,
			Color:      "#000",
		},

		DetailWrapWidth: 18 * 16,
	}
}

// Layout generates the layout for the ancestor chart based on the provided options.
func (ch *AncestorChart) Layout(opts *AncestorLayoutOptions) (*AncestorLayout, error) {
	if opts == nil {
		opts = DefaultAncestorLayoutOptions()
	}

	l := new(AncestorLayout)
	l.opts = *opts
	l.title = ch.Title
	l.notes = ch.Notes
	l.nodes = make(map[int]*AlignedNode)

	ts, err := NewTextStyle(opts.TitleStyle)
	if err != nil {
		return nil, fmt.Errorf("setup default title style: %v", err)
	}
	l.titleStyle = ts

	ns, err := NewTextStyle(opts.NoteStyle)
	if err != nil {
		return nil, fmt.Errorf("setup default note style: %v", err)
	}
	l.noteStyle = ns

	hs, err := NewTextStyle(opts.HeadingStyle)
	if err != nil {
		return nil, fmt.Errorf("setup default heading style: %v", err)
	}
	l.headingStyle = hs

	ds, err := NewTextStyle(opts.DetailStyle)
	if err != nil {
		return nil, fmt.Errorf("setup default detail style: %v", err)
	}
	l.detailStyle = ds

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

		largestNodeHeight := Pixel(0)
		largestNodeWidth := Pixel(0)
		for _, b := range l.grid[col] {
			if b == nil {
				continue
			}
			if b.Height > largestNodeHeight {
				largestNodeHeight = b.Height
			}
			if b.Width > largestNodeWidth {
				largestNodeWidth = b.Width
			}
		}
		colWidths[col] = largestNodeWidth + l.opts.Hspace*2

		// Give each node equal vertical space
		colHeight := Pixel(pop) * largestNodeHeight

		// Add VSpace between each mother and father node
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

	// reposition nodes

	lowestTopPos := Pixel(200000)
	x := l.opts.Margin
	// number of vertical divisions is 2^col (col 0 has entire vertical space, col 1 splits it in two, col 2 splits in four)
	divisions := 1
	for col := range l.grid {
		spacing := gridHeight / Pixel(divisions)
		for row, b := range l.grid[col] {
			if b == nil {
				continue
			}
			b.LeftPos = x

			// centre the node in the division
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
	titleHeight, _ := titleDimensions(l.title, l.notes, l.titleStyle, l.noteStyle)

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
			childNode := l.grid[col-1][childIdx]

			// draw hook projecting from left edge of parent
			l.connectors = append(l.connectors, &Connector{
				Points: []Point{
					// Start just to left of node
					{X: b.LeftPos - l.opts.LineGap, Y: b.SideHookY()},

					// Move left by HookLength
					{X: b.LeftPos - l.opts.LineGap - l.opts.HookLength, Y: b.SideHookY()},

					// Move vertically to hook of child
					{X: b.LeftPos - l.opts.LineGap - l.opts.HookLength, Y: childNode.SideHookY()},

					// Move left by HookLength
					{X: b.LeftPos - l.opts.LineGap - l.opts.HookLength - l.opts.Hspace, Y: childNode.SideHookY()},
				},
			})

		}
	}

	l.legend = LeftAlignedLegend(
		Point{X: l.opts.Margin, Y: l.opts.Margin},
		l.title,
		l.titleStyle,
		l.notes,
		l.noteStyle)

	return l, nil
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
	nodes      map[int]*AlignedNode
	grid       [][]*AlignedNode // col, row
	rows       int
	connectors []*Connector

	titleStyle   TextStyle
	noteStyle    TextStyle
	headingStyle TextStyle
	detailStyle  TextStyle

	legend *Blurb
}

var _ Layout = (*AncestorLayout)(nil)

// Width returns the width of the layout.
func (l *AncestorLayout) Width() Pixel { return l.width }

// Height returns the height of the layout.
func (l *AncestorLayout) Height() Pixel { return l.height }

// Margin returns the margin of the layout.
func (l *AncestorLayout) Margin() Pixel { return l.opts.Margin }

// Legend returns the legend element of the layout.
func (l *AncestorLayout) Legend() *Blurb {
	return l.legend
}

// Blurbs returns all the blurbs in the layout.
func (l *AncestorLayout) Blurbs() []*Blurb {
	bs := make([]*Blurb, 0, len(l.nodes))
	for _, n := range l.nodes {
		bs = append(bs, n.Blurb())
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
func (l *AncestorLayout) addPerson(p *AncestorPerson, col int, row int, child *AlignedNode) *AlignedNode {
	b := l.newNode(p.ID, p.Headings, p.Details, col, row, child)

	for len(l.grid) <= col {
		l.grid = append(l.grid, make([]*AlignedNode, colPopulation(len(l.grid)+1)))
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

// newNode creates a new node for the given person at the specified column and row.
func (l *AncestorLayout) newNode(id int, headings []string, details []string, col int, row int, child *AlignedNode) *AlignedNode {
	// texts = l.wrapTexts(texts)
	b := &AlignedNode{
		ID:                  id,
		Col:                 col,
		Row:                 col,
		AbsolutePositioning: true,

		HeadingTexts: TextSection{
			Lines: []string{},
			Style: l.headingStyle,
		},
		DetailTexts: TextSection{
			Lines: []string{},
			Style: l.detailStyle,
		},

		SideHookOffset: (l.opts.HeadingStyle.LineHeight * 2) / 3,
		LeftNeighbour:  child,
	}

	if len(headings) > 0 {
		b.HeadingTexts.Lines = headings
		b.Height = b.HeadingTexts.Style.LineHeight * Pixel(len(b.HeadingTexts.Lines))
	} else {
		b.HeadingTexts.Lines = append(b.HeadingTexts.Lines, details[0])
		b.Height = b.HeadingTexts.Style.LineHeight
		details = details[1:]
	}

	if len(details) > 0 {
		b.DetailTexts.Lines = details
		b.Height += b.DetailTexts.Style.LineHeight * Pixel(len(b.DetailTexts.Lines))
	}

	for i := range b.HeadingTexts.Lines {
		wl := b.HeadingTexts.Style.MeasureWidth(b.HeadingTexts.Lines[i])
		if wl > b.Width {
			b.Width = wl
		}
	}
	for i := range b.DetailTexts.Lines {
		wl := b.DetailTexts.Style.MeasureWidth(b.DetailTexts.Lines[i])
		if wl > b.Width {
			b.Width = wl
		}
	}

	l.nodes[id] = b

	return b
}

// colPopulation returns the expected population of each column
func colPopulation(col int) int {
	return 1 << col
}

// AlignedNode represents a visual element in the layout, typically used to display information about a person in a chart.
// It includes various properties to control its positioning, text content, and relationships with other blurbs.
type AlignedNode struct {
	ID           int
	HeadingTexts TextSection
	DetailTexts  TextSection
	Tags         []string

	// Text          []string
	CentreText          bool         // true if the text for this blurb is better presented as centred
	Width               Pixel        // Width is the horizontal extent of the Blurb
	AbsolutePositioning bool         // when true, the position of the blurb is controlled by TopPos and LeftPos, otherwise it is calculated relative to neighbours
	TopPos              Pixel        // TopPos is the absolute vertical position of the upper edge of the Blurb
	LeftPos             Pixel        // LeftPos is the absolute horizontal position of the left edge of the Blurb
	Height              Pixel        // Height is the vertical extent of the Blurb
	Col                 int          // column the blurb appears in for layouts that use columns
	Row                 int          // row the blurb appears in for layouts that use rows
	LeftPad             Pixel        // required padding to left of blurb to separate families
	NoShift             bool         // when true the left shift will not be changed
	KeepTightRight      *AlignedNode // the blurb to the right that this blurb should keep as close as possible to
	LeftNeighbour       *AlignedNode // the blurb to the left of this one, when non-nil will be used for horizontal positioning
	Parent              *AlignedNode
	TopHookOffset       Pixel // TopHookOffset is the offset from the left of the blurb where any dropped connecting line should finish (ensures it is within the bounds of the name, even if subsequent detail lines are longer)
	SideHookOffset      Pixel // SideHookOffset is the offset from the top of the blurb where any connecting line should finish

	FirstChild *AlignedNode
	LastChild  *AlignedNode
}

func (n *AlignedNode) Blurb() *Blurb {
	b := &Blurb{
		Texts: []TextSection{
			n.HeadingTexts,
			n.DetailTexts,
		},
		// HeadingTexts: n.HeadingTexts,
		// DetailTexts:  n.DetailTexts,
		Alignment: AlignmentLeft,
	}
	if n.CentreText {
		b.Alignment = AlignmentCenter
	}

	b.X = n.X()
	b.Y = n.Y()
	b.Width = b.Width
	b.Height = n.Height

	return b
}

// X returns the horizontal position of the centre of the Blurb
func (b *AlignedNode) X() Pixel {
	if b.AbsolutePositioning {
		return b.LeftPos + b.Width/2
	}
	left := Pixel(0)
	if b.LeftNeighbour != nil {
		left = b.LeftNeighbour.Right()
	}
	left += b.LeftPad
	return left + b.Width/2
}

// Y returns the vertical position of the centre of the Blurb
func (b *AlignedNode) Y() Pixel {
	return b.TopPos + b.Height/2
}

// Left returns the horizontal position of the leftmost edge of the Blurb
func (b *AlignedNode) Left() Pixel {
	if b.AbsolutePositioning {
		return b.LeftPos
	}
	return b.X() - b.Width/2
}

// Right returns the horizontal position of the rightmost edge of the Node
func (b *AlignedNode) Right() Pixel {
	if b.AbsolutePositioning {
		return b.LeftPos + b.Width
	}
	return b.X() + b.Width/2
}

// Bottom returns the vertical position of the lower edge of the Node
func (b *AlignedNode) Bottom() Pixel {
	return b.TopPos + b.Height
}

func (b *AlignedNode) TopHookX() Pixel {
	return b.Left() + b.TopHookOffset
}

func (b *AlignedNode) SideHookY() Pixel {
	return b.TopPos + b.SideHookOffset
}

func wrapText(texts []string, maxWidth Pixel, ts TextStyle) []string {
	if len(texts) == 0 {
		return []string{}
	}
	wrapped := make([]string, 0, len(texts))
	for i := 0; i < len(texts); i++ {
		wl := ts.MeasureWidth(texts[i])
		if wl <= maxWidth {
			wrapped = append(wrapped, texts[i])
			continue
		}

		words := strings.Fields(texts[i])
		if len(words) == 0 {
			wrapped = append(wrapped, "")
			continue
		}

		var line string
		for w := 0; w < len(words); w++ {
			candidate := line
			if len(line) != 0 {
				candidate += " "
			}
			candidate += words[w]
			wl := ts.MeasureWidth(candidate)
			if wl >= maxWidth {
				if len(line) == 0 {
					wrapped = append(wrapped, candidate)
					line = ""
				} else {
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

func titleDimensions(title string, notes []string, titleStyle TextStyle, noteStyle TextStyle) (Pixel, Pixel) {
	if title == "" && len(notes) == 0 {
		return 0, 0
	}

	var h, w Pixel

	if title != "" {
		h += titleStyle.LineHeight
		w = titleStyle.MeasureWidth(title)
	}

	if len(notes) != 0 {
		h += noteStyle.LineHeight * Pixel(len(notes))
		for i := 0; i < len(notes); i++ {
			wl := noteStyle.MeasureWidth(notes[i])
			if wl > w {
				w = wl
			}
		}
	}

	return h, w
}
