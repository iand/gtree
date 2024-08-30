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

	Hspace     Pixel // Hspace is the horizontal spacing between blurbs within the same family.
	LineWidth  Pixel // LineWidth is the width of the lines connecting blurbs.
	Margin     Pixel // Margin is the margin added to the entire drawing.
	FamilyDrop Pixel // FamilyDrop is the length of the line drawn from parents to the children group line.
	ChildDrop  Pixel // ChildDrop is the length of the line drawn from the children group line to a child.
	LineGap    Pixel // LineGap is the distance between a connecting line and any text.

	TitleStyle   TextStyle // TitleStyle is the style of the font to use for the title of the chart.
	NoteStyle    TextStyle // NoteStyle is the style of the font to use for the notes of the chart.
	HeadingStyle TextStyle // HeadingStyle is the style of the font to use for the first line of each blurb.
	DetailStyle  TextStyle // DetailStyle is the style of the font to use for the subsequent lines of each blurb after the first.

	DetailWrapWidth Pixel // DetailWrapWidth is the maximum width of detail text before wrapping to a new line.
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
		TitleStyle: TextStyle{
			FontSize:   40,
			LineHeight: 42,
			Color:      "#000",
		},
		NoteStyle: TextStyle{
			FontSize:   20,
			LineHeight: 22,
			Color:      "#000",
		},
		HeadingStyle: TextStyle{
			FontSize:   20,
			LineHeight: 22,
			Color:      "#000",
		},
		DetailStyle: TextStyle{
			FontSize:   16,
			LineHeight: 18,
			Color:      "#000",
		},
	}
}

// Layout generates the layout for the descendant chart based on the provided options.
func (ch *DescendantChart) Layout(opts *LayoutOptions) *DescendantLayout {
	if opts == nil {
		opts = DefaultLayoutOptions()
	}

	l := new(DescendantLayout)
	l.title = ch.Title
	l.notes = ch.Notes
	l.opts = *opts
	l.blurbs = make(map[int]*Blurb)
	l.generationDrop = l.opts.LineWidth + l.opts.LineGap + l.opts.LineGap + l.opts.ChildDrop + l.opts.FamilyDrop

	l.addPerson(ch.Root, 0, nil)

	a := new(SpreadingDescendantArranger)
	a.Arrange(l)

	return l
}

// DescendantLayout represents the layout of a descendant chart, including dimensions and layout options.
type DescendantLayout struct {
	title          string
	notes          []string
	width          Pixel
	height         Pixel
	generationDrop Pixel // distance between generations

	opts LayoutOptions

	blurbs     map[int]*Blurb
	connectors []*Connector
	rows       [][]*Blurb
}

// Width returns the width of the layout.
func (l *DescendantLayout) Width() Pixel { return l.width }

// Height returns the height of the layout.
func (l *DescendantLayout) Height() Pixel { return l.height }

// Margin returns the margin of the layout.
func (l *DescendantLayout) Margin() Pixel { return l.opts.Margin }

// Title returns the title element of the layout.
func (l *DescendantLayout) Title() TextElement {
	return TextElement{
		Text:  l.title,
		Style: l.opts.TitleStyle,
	}
}

// Notes returns the notes elements of the layout.
func (l *DescendantLayout) Notes() []TextElement {
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
func (l *DescendantLayout) Blurbs() []*Blurb {
	bs := make([]*Blurb, 0, len(l.blurbs))
	for _, b := range l.blurbs {
		bs = append(bs, b)
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
func (l *DescendantLayout) addPerson(p *DescendantPerson, row int, parent *Blurb) *Blurb {
	b := l.newBlurb(p.ID, p.Headings, p.Details, p.Tags, row, parent)

	var prevSpouseWithChildren *Blurb
	var lastChildOfPrevFamily *Blurb

	for fi := range p.Families {
		relText := "="
		if len(p.Families) > 1 {
			relText += fmt.Sprintf(" (%d)", fi+1)
		}
		relDetails := []string{relText}
		relDetails = append(relDetails, p.Families[fi].Details...)

		var rel, sp *Blurb
		var famCentre *Blurb
		var famRightmost *Blurb
		if p.Families[fi].Other != nil {
			rel = l.newBlurb(-p.Families[fi].Other.ID, []string{}, relDetails, []string{}, row, nil)
			rel.CentreText = true
			famCentre = rel

			// Attempt to keep with spouse relation marker
			b.KeepWith = append(b.KeepWith, rel)
			rel.KeepWith = append(rel.KeepWith, b)

			sp = l.addPerson(p.Families[fi].Other, row, nil)
			sp.NoShift = true

			sp.KeepWith = append(sp.KeepWith, rel)
			rel.KeepWith = append(rel.KeepWith, sp)
			famRightmost = sp
		} else {
			famCentre = b
			famRightmost = b
		}

		if len(p.Families[fi].Children) > 0 {
			prevSpouseWithChildren = famRightmost
			if lastChildOfPrevFamily != nil {
				// Attempt to keep relation marker right of last child in previous family to avoid merging of descent lines
				famCentre.KeepRightOf = append(famCentre.KeepRightOf, lastChildOfPrevFamily)
			}

		}

		var prevChild *Blurb
		for ci := range p.Families[fi].Children {
			c := l.addPerson(p.Families[fi].Children[ci], row+1, famCentre)
			if ci == 0 {
				b.FirstChild = c
			}
			if ci == len(p.Families[fi].Children)-1 {
				b.LastChild = c
			}

			if rel != nil {
				// Attempt to keep with relation marker
				c.KeepWith = append(c.KeepWith, rel)
				rel.KeepWith = append(rel.KeepWith, c)

				// Attempt to keep relation marker right of first child if there are multiple childen
				if ci == 0 && len(p.Families[fi].Children) > 1 {
					rel.KeepRightOf = append(rel.KeepRightOf, c)
				}
			} else {
				// Attempt to keep with parent
				c.KeepWith = append(c.KeepWith, b)
				b.KeepWith = append(b.KeepWith, c)

			}

			if prevChild != nil {
				// Attempt to keep with previous child
				c.KeepWith = append(c.KeepWith, prevChild)
			}
			prevChild = c

			// Attempt to keep with grandparent marker, to encourage tree to look centred
			if b.LeftStop == nil {
				b.LeftStop = c
			}
			b.RightStop = c

			if sp != nil && sp.LeftStop == nil {
				sp.LeftStop = c
			}
			if rel != nil && rel.LeftStop == nil {
				rel.LeftStop = c
			}

			// Attempt to keep child right of previous spouse with children to avoid merging of descent lines
			if fi > 0 && prevSpouseWithChildren != nil {
				c.KeepRightOf = append(c.KeepRightOf, prevSpouseWithChildren)
			}

			if ci == len(p.Families[fi].Children)-1 {
				lastChildOfPrevFamily = c
			}
		}
	}

	return b
}

// newBlurb creates a new blurb for the given person or family at the specified row.
func (l *DescendantLayout) newBlurb(id int, headings []string, texts []string, tags []string, row int, parent *Blurb) *Blurb {
	texts = wrapText(texts, l.opts.DetailWrapWidth, l.opts.DetailStyle.FontSize)
	b := &Blurb{
		ID:             id,
		Row:            row,
		Parent:         parent,
		TopHookOffset:  l.opts.Hspace * 2,
		SideHookOffset: l.opts.HeadingStyle.LineHeight / 2,
		HeadingTexts: TextSection{
			Lines: []string{},
			Style: l.opts.HeadingStyle,
		},
		DetailTexts: TextSection{
			Lines: []string{},
			Style: l.opts.DetailStyle,
		},
		Tags: tags,
	}

	if len(headings) > 0 {
		b.HeadingTexts.Lines = headings
		b.Height = b.HeadingTexts.Style.LineHeight * Pixel(len(b.HeadingTexts.Lines))
	} else {
		b.HeadingTexts.Lines = append(b.HeadingTexts.Lines, texts[0])
		b.Height = b.HeadingTexts.Style.LineHeight
		texts = texts[1:]
	}

	if len(texts) > 0 {
		b.DetailTexts.Lines = texts
		b.Height += b.DetailTexts.Style.LineHeight * Pixel(len(b.DetailTexts.Lines))
	}

	for i := range b.HeadingTexts.Lines {
		wl := textWidth([]rune(b.HeadingTexts.Lines[i]), b.HeadingTexts.Style.FontSize)
		if wl > b.Width {
			b.Width = wl
		}
	}
	for i := range b.DetailTexts.Lines {
		wl := textWidth([]rune(b.DetailTexts.Lines[i]), b.DetailTexts.Style.FontSize)
		if wl > b.Width {
			b.Width = wl
		}
	}

	l.blurbs[id] = b

	for len(l.rows) <= row {
		l.rows = append(l.rows, []*Blurb{})
	}
	l.rows[row] = append(l.rows[row], b)

	return b
}

type SpreadingDescendantArranger struct{}

func (a *SpreadingDescendantArranger) Arrange(l *DescendantLayout) {
	// spread rows vertically
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

	// spread blurbs in last row evenly
	left := Pixel(0)
	bs := l.rows[len(l.rows)-1]
	for i := range bs {
		if i > 0 {
			left += l.opts.Hspace
			if bs[i].Parent != bs[i-1].Parent {
				// extra space between families
				left += l.opts.Hspace * 2
			}
		}
		bs[i].LeftPos = left
		left += bs[i].Width
	}

	if len(l.rows) == 1 {
		return
	}

	// work up from bottom row spreading out blurbs so subtrees don't overlap
	for row := len(l.rows) - 2; row >= 0; row-- {
		minLeft := Pixel(0)
		bs := l.rows[row]
		for i := range bs {
			if i > 0 {
				minLeft += l.opts.Hspace
				if bs[i].Parent != bs[i-1].Parent {
					// extra space between families
					minLeft += l.opts.Hspace * 2
				}
			}
			if bs[i].FirstChild != nil {
				// centre over children
				w := bs[i].LastChild.Right() - bs[i].FirstChild.Left()

				// This is centre point over children
				x := bs[i].FirstChild.Left() + w/2

				// adjust to the left side of the blurb
				x -= bs[i].Width / 2

				if x < minLeft {
					for j := i; j < len(bs); j++ {
						a.shiftChildren(l, row+1, bs[j], minLeft-x)
					}
				} else {
					minLeft = x
				}

			}

			bs[i].LeftPos = minLeft
			minLeft += bs[i].Width

		}
	}

	// close up gaps by pulling across any early siblings that don't have children
	for row := range l.rows {
		bs := l.rows[row]
		for i := len(bs) - 1; i >= 1; i-- {
			if bs[i-1].FirstChild == nil && bs[i].Parent != nil && bs[i-1].Parent != nil && bs[i].Parent == bs[i-1].Parent && bs[i].Left()-bs[i-1].Right() > l.opts.Hspace {
				bs[i-1].LeftPos = bs[i].Left() - l.opts.Hspace - bs[i-1].Width
			}
		}
	}

	a.centreBlurbs(l)

	// This is top-down layout
	l.connectors = []*Connector{}
	for _, b := range l.blurbs {
		if b.Parent != nil {
			if b.Parent.ID > 0 && b.Parent.FirstChild == b.Parent.LastChild {
				l.connectors = append(l.connectors, &Connector{
					Points: []Point{
						// Start just above blurb
						{X: b.TopHookX(), Y: b.TopPos - l.opts.LineGap},
						// Move up to parent
						{X: b.TopHookX(), Y: b.Parent.Bottom() + l.opts.LineGap},
					},
				})
			} else {
				l.connectors = append(l.connectors, &Connector{
					Points: []Point{
						// Start just above blurb
						{X: b.TopHookX(), Y: b.TopPos - l.opts.LineGap},
						// Move up by ChildDrop
						{X: b.TopHookX(), Y: b.TopPos - l.opts.LineGap - l.opts.ChildDrop},
						// Move horizontally to centre of parent
						{X: b.Parent.X(), Y: b.TopPos - l.opts.LineGap - l.opts.ChildDrop},
						// Move up to centre of parent
						{X: b.Parent.X(), Y: b.Parent.Bottom() + l.opts.LineGap},
					},
				})
			}
		}
	}
}

func (a *SpreadingDescendantArranger) shiftChildren(l *DescendantLayout, row int, parent *Blurb, shift Pixel) {
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

// centreBlurbs centres the blurbs within the layout.
func (a *SpreadingDescendantArranger) centreBlurbs(l *DescendantLayout) {
	var minX, maxX, minY, maxY Pixel
	initialized := false

	for _, b := range l.blurbs {
		if l.opts.Debug {
			slog.Info("blurb position", "l", b.Left(), "r", b.Right(), "t", b.TopPos, "b", b.Bottom())
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

	th, _ := titleDimensions(l.title, l.notes, l.opts.TitleStyle, l.opts.NoteStyle)
	minY -= th

	for _, bs := range l.rows {
		for i := range bs {
			if i == 0 {
				bs[i].LeftPad -= minX
			}
			bs[i].TopPos -= minY
		}
	}

	l.width = maxX - minX
	l.height = maxY - minY
}
