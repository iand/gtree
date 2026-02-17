package gtree

import (
	"fmt"
	"log/slog"
)

// DescendantChart represents a chart of descendants, with the earliest ancestor (root person) at the top.
// Each successive generation is arranged in horizontal rows, with the next generation placed directly below
// the previous one. This layout visually depicts the lineage, with descendants expanding downward from
// the root person.
type DescendantChart struct {
	Title string
	Notes []string
	Root  *DescendantPerson
}

// DescendantPerson represents an individual in the descendant chart, including their ID, details, and families.
type DescendantPerson struct {
	ID       int
	Headings []string
	Details  []string
	Families []*DescendantFamily
	Tags     []string
}

// DescendantFamily represents a family unit, including the spouse and their children.
type DescendantFamily struct {
	Other    *DescendantPerson
	Details  []string
	Children []*DescendantPerson
}

// LayoutOptions defines various layout parameters for rendering the descendant chart.
type LayoutOptions struct {
	Debug      bool // Debug indicates whether to emit logging and debug information.
	Iterations int  // Number of iterations of adjustment to run

	Hspace     Pixel // Hspace is the horizontal spacing between nodes within the same family.
	LineWidth  Pixel // LineWidth is the width of the lines connecting nodes.
	Margin     Pixel // Margin is the margin added to the entire drawing.
	FamilyDrop Pixel // FamilyDrop is the length of the line drawn from parents to the children group line.
	ChildDrop  Pixel // ChildDrop is the length of the line drawn from the children group line to a child.
	LineGap    Pixel // LineGap is the distance between a connecting line and any text.

	TitleStyle   TextStyleOption // TitleStyle is the style of the font to use for the title of the chart.
	NoteStyle    TextStyleOption // NoteStyle is the style of the font to use for the notes of the chart.
	HeadingStyle TextStyleOption // HeadingStyle is the style of the font to use for the first line of each node.
	DetailStyle  TextStyleOption // DetailStyle is the style of the font to use for the subsequent lines of each node after the first.

	DetailWrapWidth Pixel // DetailWrapWidth is the maximum width of detail text before wrapping to a new line.

	BackgroundColor string // BackgroundColor is the color of the background, empty for transparent/no-background fill
}

// DefaultLayoutOptions returns the default layout options for rendering the descendant chart.
func DefaultLayoutOptions() *LayoutOptions {
	return &LayoutOptions{
		Iterations:      30000,
		DetailWrapWidth: 18 * 16,
		Hspace:          16,
		LineWidth:       2,
		Margin:          16,
		FamilyDrop:      48,
		ChildDrop:       16,
		LineGap:         8,
		TitleStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   40,
			LineHeight: 42,
			Color:      "#000",
		},
		NoteStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   20,
			LineHeight: 22,
			Color:      "#000",
		},
		HeadingStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   20,
			LineHeight: 22,
			Color:      "#000",
		},
		DetailStyle: TextStyleOption{
			FontNames:  []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"},
			FontSize:   16,
			LineHeight: 18,
			Color:      "#000",
		},
		BackgroundColor: "#FFF",
	}
}

type TextStyleOption struct {
	FontNames  []string // list of font names in priority order
	FontSize   Pixel    // FontSize is the size of the font to use for the text of each node.
	Color      string   // Color is the color of the text. The default is black #000000.
	LineHeight Pixel    // TODO: remove
}

// Layout generates the layout for the descendant chart based on the provided options.
func (ch *DescendantChart) Layout(opts *LayoutOptions) (*DescendantLayout, error) {
	if opts == nil {
		opts = DefaultLayoutOptions()
	}

	l := new(DescendantLayout)
	l.title = ch.Title
	l.notes = ch.Notes
	l.opts = *opts
	l.nodes = make(map[int]*AlignedNode)
	l.generationDrop = l.opts.LineWidth + l.opts.LineGap + l.opts.LineGap + l.opts.ChildDrop + l.opts.FamilyDrop

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

	l.addPerson(ch.Root, 0, nil)

	a := new(SpreadingDescendantArranger)
	a.Arrange(l)

	l.legend = LeftAlignedLegend(
		Point{X: l.opts.Margin, Y: l.opts.Margin},
		l.title,
		l.titleStyle,
		l.notes,
		l.noteStyle)

	return l, nil
}

// DescendantLayout represents the layout of a descendant chart, including dimensions and layout options.
type DescendantLayout struct {
	title          string
	notes          []string
	width          Pixel
	height         Pixel
	generationDrop Pixel // distance between generations

	opts LayoutOptions

	nodes        map[int]*AlignedNode
	connectors   []*Connector
	rows         [][]*AlignedNode
	titleStyle   TextStyle
	noteStyle    TextStyle
	headingStyle TextStyle
	detailStyle  TextStyle

	legend *Blurb
}

var _ Layout = (*DescendantLayout)(nil)

// Width returns the width of the layout.
func (l *DescendantLayout) Width() Pixel { return l.width }

// Height returns the height of the layout.
func (l *DescendantLayout) Height() Pixel { return l.height }

// Margin returns the margin of the layout.
func (l *DescendantLayout) Margin() Pixel { return l.opts.Margin }

// Title returns the legend element of the layout.
func (l *DescendantLayout) Legend() *Blurb {
	return l.legend
}

func (l *DescendantLayout) Background() string      { return "" }
func (l *DescendantLayout) BackgroundColor() string { return l.opts.BackgroundColor }

// Blurbs returns all the blurbs in the layout.
func (l *DescendantLayout) Blurbs() []*Blurb {
	bs := make([]*Blurb, 0, len(l.nodes))
	for _, n := range l.nodes {
		bs = append(bs, n.Blurb())
	}
	return bs
}

// Connectors returns all the connectors in the layout.
func (l *DescendantLayout) Connectors() []*Connector {
	return l.connectors
}

// Debug reports whether the layout is in debug mode.
func (l *DescendantLayout) Debug() bool { return l.opts.Debug }

// addPerson adds a person and their family to the layout at the specified row.
func (l *DescendantLayout) addPerson(p *DescendantPerson, row int, parent *AlignedNode) *AlignedNode {
	b := l.newNode(p.ID, p.Headings, p.Details, p.Tags, row, parent)

	for fi := range p.Families {
		relText := "="
		if len(p.Families) > 1 {
			relText += fmt.Sprintf(" (%d)", fi+1)
		}
		relDetails := []string{relText}
		relDetails = append(relDetails, p.Families[fi].Details...)

		var rel, sp *AlignedNode
		var famCentre *AlignedNode
		// var famRightmost *Node
		if p.Families[fi].Other != nil {
			rel = l.newNode(-p.Families[fi].Other.ID, []string{}, relDetails, []string{}, row, nil)
			rel.CentreText = true
			famCentre = rel

			// Attempt to keep with spouse relation marker if this is the first one
			if b.KeepTightRight == nil {
				b.KeepTightRight = rel
			}

			sp = l.addPerson(p.Families[fi].Other, row, nil)
			sp.NoShift = true

		} else {
			famCentre = b
		}

		// var prevChild *Node
		for ci := range p.Families[fi].Children {
			c := l.addPerson(p.Families[fi].Children[ci], row+1, famCentre)

			if rel != nil {

				if ci == 0 {
					rel.FirstChild = c
				}
				if ci == len(p.Families[fi].Children)-1 {
					rel.LastChild = c
				}

			} else {
				// Attempt to keep with parent

				if ci == 0 {
					b.FirstChild = c
				}
				if ci == len(p.Families[fi].Children)-1 {
					b.LastChild = c
				}

			}

		}
	}

	return b
}

// newNode creates a new node for the given person or family at the specified row.
func (l *DescendantLayout) newNode(id int, headings []string, details []string, tags []string, row int, parent *AlignedNode) *AlignedNode {
	details = wrapText(details, l.opts.DetailWrapWidth, l.detailStyle)
	b := &AlignedNode{
		ID:             id,
		Row:            row,
		Parent:         parent,
		TopHookOffset:  l.opts.Hspace * 2,
		SideHookOffset: l.opts.HeadingStyle.LineHeight / 2,
		HeadingTexts: TextSection{
			Lines: []string{},
			Style: l.headingStyle,
		},
		DetailTexts: TextSection{
			Lines: []string{},
			Style: l.detailStyle,
		},
		Tags: tags,
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

	for len(l.rows) <= row {
		l.rows = append(l.rows, []*AlignedNode{})
	}
	l.rows[row] = append(l.rows[row], b)

	return b
}

type SpreadingDescendantArranger struct{}

func (a *SpreadingDescendantArranger) Arrange(l *DescendantLayout) {
	a.assignVerticalPositions(l)
	a.spreadLastRow(l)
	if len(l.rows) > 1 {
		a.centreParentsOverChildren(l)
	}
	a.centreNodes(l)
	a.buildConnectors(l)
}

// assignVerticalPositions sets the vertical position of each row and
// links horizontal neighbours within each row.
func (a *SpreadingDescendantArranger) assignVerticalPositions(l *DescendantLayout) {
	top := Pixel(0)
	for _, bs := range l.rows {
		rowHeight := Pixel(0)
		for i := range bs {
			bs[i].AbsolutePositioning = true
			bs[i].TopPos = top
			if i > 0 {
				bs[i].LeftNeighbour = bs[i-1]
			}
			rowHeight = max(rowHeight, bs[i].Height)
		}
		top += rowHeight + l.generationDrop
	}
}

// spreadLastRow places the leaf row nodes left to right with spacing,
// adding extra space between different families.
func (a *SpreadingDescendantArranger) spreadLastRow(l *DescendantLayout) {
	left := Pixel(0)
	bs := l.rows[len(l.rows)-1]
	for i := range bs {
		if i > 0 {
			left += l.opts.Hspace
			if bs[i].Parent != bs[i-1].Parent {
				left += l.opts.Hspace * 2
			}
		}
		bs[i].LeftPos = left
		left += bs[i].Width
	}
}

// centreParentsOverChildren works from the second-to-last row upward,
// centering each node that has children over its child span. After
// positioning each row it immediately applies tight-packing so that
// the row above sees final positions.
func (a *SpreadingDescendantArranger) centreParentsOverChildren(l *DescendantLayout) {
	for row := len(l.rows) - 2; row >= 0; row-- {
		bs := l.rows[row]
		a.spreadRow(l, bs, row)
		a.applyKeepTightRight(bs, l.opts.Hspace)
		a.closeGaps(bs, l.opts.Hspace)
	}
}

// spreadRow positions each node in a row, centering nodes that have
// children over their child span. If a node's children are too far
// left, they are shifted right to avoid overlap.
func (a *SpreadingDescendantArranger) spreadRow(l *DescendantLayout, bs []*AlignedNode, row int) {
	minLeft := Pixel(0)
	for i := range bs {
		if i > 0 {
			minLeft += l.opts.Hspace
			if bs[i].Parent != bs[i-1].Parent {
				minLeft += l.opts.Hspace * 2
			}
		}
		if bs[i].FirstChild != nil {
			childSpan := bs[i].LastChild.Right() - bs[i].FirstChild.Left()
			childCentre := bs[i].FirstChild.Left() + childSpan/2
			x := childCentre - bs[i].Width/2

			if x < minLeft {
				shift := minLeft - x
				for j := i; j < len(bs); j++ {
					a.shiftChildren(l, row+1, bs[j], shift)
				}
			} else {
				minLeft = x
			}
		}

		bs[i].LeftPos = minLeft
		minLeft += bs[i].Width
	}
}

// applyKeepTightRight positions each node that has a KeepTightRight
// reference so it sits immediately to the left of that neighbour.
// This pulls a person node tight against its relationship marker.
func (a *SpreadingDescendantArranger) applyKeepTightRight(bs []*AlignedNode, hspace Pixel) {
	for i := 0; i < len(bs)-2; i++ {
		if bs[i].KeepTightRight == nil || bs[i].KeepTightRight != bs[i+1] {
			continue
		}
		bs[i].LeftPos = bs[i+1].Left() - hspace - bs[i].Width
	}
}

// closeGaps pulls childless nodes toward their right sibling when
// they share the same parent and there is excess space between them.
func (a *SpreadingDescendantArranger) closeGaps(bs []*AlignedNode, hspace Pixel) {
	for i := len(bs) - 1; i >= 1; i-- {
		prev := bs[i-1]
		cur := bs[i]
		if prev.FirstChild != nil {
			continue
		}
		if prev.Parent == nil || cur.Parent == nil || prev.Parent != cur.Parent {
			continue
		}
		if cur.Left()-prev.Right() > hspace {
			prev.LeftPos = cur.Left() - hspace - prev.Width
		}
	}
}

// buildConnectors creates the vertical and horizontal lines connecting
// each child node to its parent.
func (a *SpreadingDescendantArranger) buildConnectors(l *DescendantLayout) {
	l.connectors = []*Connector{}
	for _, b := range l.nodes {
		if b.Parent == nil {
			continue
		}

		onlyChild := b.Parent.ID > 0 && b.Parent.FirstChild == b.Parent.LastChild
		if onlyChild {
			// Straight vertical line to parent
			l.connectors = append(l.connectors, &Connector{
				Points: []Point{
					{X: b.TopHookX(), Y: b.TopPos - l.opts.LineGap},
					{X: b.TopHookX(), Y: b.Parent.Bottom() + l.opts.LineGap},
				},
			})
		} else {
			// L-shaped connector via a horizontal group line
			l.connectors = append(l.connectors, &Connector{
				Points: []Point{
					{X: b.TopHookX(), Y: b.TopPos - l.opts.LineGap},
					{X: b.TopHookX(), Y: b.TopPos - l.opts.LineGap - l.opts.ChildDrop},
					{X: b.Parent.X(), Y: b.TopPos - l.opts.LineGap - l.opts.ChildDrop},
					{X: b.Parent.X(), Y: b.Parent.Bottom() + l.opts.LineGap},
				},
			})
		}
	}
}

// shiftChildren recursively shifts all descendants of parent on the
// given row and below by the specified amount.
func (a *SpreadingDescendantArranger) shiftChildren(l *DescendantLayout, row int, parent *AlignedNode, shift Pixel) {
	if parent.FirstChild == nil || row > len(l.rows)-1 {
		return
	}
	bs := l.rows[row]
	for i := range bs {
		if bs[i].Parent == parent {
			bs[i].LeftPos += shift
			a.shiftChildren(l, row+1, bs[i], shift)
		}
	}
}

// centreNodes centres the nodes within the layout.
func (a *SpreadingDescendantArranger) centreNodes(l *DescendantLayout) {
	var minX, maxX, minY, maxY Pixel
	initialized := false

	for _, b := range l.nodes {
		if l.opts.Debug {
			slog.Info("node position", "l", b.Left(), "r", b.Right(), "t", b.TopPos, "b", b.Bottom())
		}
		if !initialized {
			minX = b.Left()
			maxX = b.Right()
			minY = b.TopPos
			maxY = b.Bottom()
			initialized = true
			continue
		}
		minX = min(minX, b.Left())
		maxX = max(maxX, b.Right())
		minY = min(minY, b.TopPos)
		maxY = max(maxY, b.Bottom())
	}

	minX -= l.opts.Margin
	maxX += l.opts.Margin
	minY -= l.opts.Margin
	maxY += l.opts.Margin

	th, _ := titleDimensions(l.title, l.notes, l.titleStyle, l.noteStyle)
	minY -= th

	for _, bs := range l.rows {
		for i := range bs {
			if bs[i].AbsolutePositioning {
				bs[i].LeftPos -= minX
				bs[i].TopPos -= minY
			} else {
				if i == 0 {
					bs[i].LeftPad -= minX
				}
				bs[i].TopPos -= minY
			}
		}
	}

	l.width = maxX - minX
	l.height = maxY - minY
}
