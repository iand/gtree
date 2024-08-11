package gtree

import (
	"fmt"
	"log/slog"
	"math"
	prand "math/rand"
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
	Details  []string
	Families []*DescendantFamily
}

// DescendantFamily represents a family unit, including the spouse and their children.
type DescendantFamily struct {
	Other    *DescendantPerson
	Details  []string
	Children []*DescendantPerson
}

// LayoutOptions defines various layout parameters for rendering the descendant chart.
type LayoutOptions struct {
	Debug bool // Debug indicates whether to emit logging and debug information.

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
	l.align()
	l.reflow()

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
	b := l.newBlurb(p.ID, p.Details, row, parent)

	var prevSpouseWithChildren *Blurb
	var lastChildOfPrevFamily *Blurb

	for fi := range p.Families {
		relText := "="
		if len(p.Families) > 1 {
			relText += fmt.Sprintf(" (%d)", fi+1)
		}
		relDetails := []string{relText}
		relDetails = append(relDetails, p.Families[fi].Details...)
		rel := l.newBlurb(-p.Families[fi].Other.ID, relDetails, row, nil)
		rel.CentreText = true

		// Attempt to keep with spouse relation marker
		b.KeepWith = append(b.KeepWith, rel)
		rel.KeepWith = append(rel.KeepWith, b)

		sp := l.addPerson(p.Families[fi].Other, row, nil)
		sp.NoShift = true

		sp.KeepWith = append(sp.KeepWith, rel)
		rel.KeepWith = append(rel.KeepWith, sp)

		if len(p.Families[fi].Children) > 0 {
			prevSpouseWithChildren = sp
			if lastChildOfPrevFamily != nil {
				// Attempt to keep relation marker right of last child in previous family to avoid merging of descent lines
				rel.KeepRightOf = append(rel.KeepRightOf, lastChildOfPrevFamily)
			}

		}

		var prevChild *Blurb
		for ci := range p.Families[fi].Children {
			c := l.addPerson(p.Families[fi].Children[ci], row+1, rel)

			// Attempt to keep with relation marker
			c.KeepWith = append(c.KeepWith, rel)
			rel.KeepWith = append(rel.KeepWith, c)

			// Attempt to keep relation marker right of first child if there are multiple childen
			if ci == 0 && len(p.Families[fi].Children) > 1 {
				rel.KeepRightOf = append(rel.KeepRightOf, c)
			}

			if prevChild != nil {
				// Attempt to keep with previous child
				c.KeepWith = append(c.KeepWith, prevChild)
			}
			prevChild = c

			// Attempt to keep with grandparent marker, to encourage tree to look centred
			if parent != nil {
				c.KeepWith = append(c.KeepWith, parent)
				parent.KeepWith = append(parent.KeepWith, c)
			}

			if b.LeftStop == nil {
				b.LeftStop = c
			}
			b.RightStop = c

			if sp.LeftStop == nil {
				sp.LeftStop = c
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
func (l *DescendantLayout) newBlurb(id int, texts []string, row int, parent *Blurb) *Blurb {
	texts = wrapText(texts, l.opts.DetailWrapWidth, l.opts.DetailStyle.FontSize)
	b := &Blurb{
		ID: id,
		// Text:              texts,
		Row:            row,
		Parent:         parent,
		TopHookOffset:  l.opts.Hspace * 2,
		SideHookOffset: l.opts.HeadingStyle.LineHeight / 2,
		HeadingStyle:   l.opts.HeadingStyle,
		DetailStyle:    l.opts.DetailStyle,
	}
	if len(texts) > 0 {
		b.HeadingText = texts[0]
		b.Height = b.HeadingStyle.LineHeight
		b.Width = textWidth([]rune(b.HeadingText), b.HeadingStyle.FontSize)

		if len(texts) > 1 {
			b.DetailTexts = texts[1:]
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

	for len(l.rows) <= row {
		l.rows = append(l.rows, []*Blurb{})
	}
	l.rows[row] = append(l.rows[row], b)

	return b
}

// align aligns the blurbs and rows in the layout, ensuring proper spacing.
func (l *DescendantLayout) align() {
	// spread rows evenly
	top := Pixel(0)
	for _, bs := range l.rows {
		rowHeight := Pixel(0)
		for i := range bs {
			bs[i].TopPos = top
			if i > 0 {
				bs[i].LeftPad = l.opts.Hspace
				bs[i].LeftNeighbour = bs[i-1]

				// add a little more padding if neighbours have parents that are different
				if bs[i].ID > 0 && (bs[i].Parent != nil || bs[i-1].Parent != nil) && (bs[i].Parent != bs[i-1].Parent) {
					bs[i].LeftPad += l.opts.Hspace
				}
			}

			rowHeight = max(rowHeight, bs[i].Height)
		}
		top += rowHeight + l.generationDrop
	}

	// get parents roughly aligned over their children
	for iter := 0; iter < 3; iter++ {
		for r := len(l.rows) - 1; r >= 0; r-- {
			bs := l.rows[r]
			for i := range bs {
				if !bs[i].NoShift && bs[i].LeftStop != nil && bs[i].LeftStop.X() > bs[i].X() {
					bs[i].LeftShift += bs[i].LeftStop.X() - bs[i].X()
				}
				if bs[i].RightStop != nil && bs[i].X() > bs[i].RightStop.X() {
					bs[i].RightStop.LeftShift += bs[i].X() - bs[i].RightStop.X()
				}
			}
		}
	}
}

// jiggle randomly shifts a blurb in the layout, returning a function to undo the shift.
func (l *DescendantLayout) jiggle() func() {
	// pick a blurb at random

	var b *Blurb
	for b == nil || b.NoShift {
		row := prand.Intn(len(l.rows))
		n := prand.Intn(len(l.rows[row]))

		b = l.rows[row][n]

	}
	savedShift := b.LeftShift

	delta := Pixel(0)
	for delta == 0 || b.LeftShift+delta < 0 || (b.LeftStop != nil && b.X() > b.LeftStop.X() && b.X()+delta < b.LeftStop.X()) || (b.RightStop != nil && b.X() < b.RightStop.X() && b.X()+delta > b.RightStop.X()) {
		delta = Pixel((0.5 - (prand.Float64() * prand.Float64())) * float64(l.opts.Hspace))
	}

	b.LeftShift += delta
	return func() { b.LeftShift = savedShift }
}

// reflow adjusts the layout using simulated annealing to optimise the fitness of the layout.
func (l *DescendantLayout) reflow() {
	iters := 30000
	temp := float64(iters) * 10
	for i := 0; i < iters; i++ {
		fitnessBefore := l.fitness()
		undo := l.jiggle()
		fitnessAfter := l.fitness()

		// keep this change if the new fitness is lower
		diff := fitnessAfter - fitnessBefore
		if diff <= 0 {
			continue
		}

		t := temp / float64(i+1)

		// otherwise there is an ever decreasing chance of keeping a worse fitness
		prob := math.Exp(-float64(diff) / t)
		if prand.Float64() <= prob {
			continue
		} else {
			undo()
		}

	}

	l.centreBlurbs()

	// This is top-down layout
	l.connectors = []*Connector{}
	for _, b := range l.blurbs {
		if b.Parent != nil {
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

// centreBlurbs centres the blurbs within the layout.
func (l *DescendantLayout) centreBlurbs() {
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

// fitness calculates the fitness of the layout, used for layout optimisation.
func (l *DescendantLayout) fitness() int {
	total := 0
	for _, b := range l.blurbs {
		for _, kw := range b.KeepWith {
			total += distance(b, kw)
		}
		for _, kw := range b.KeepRightOf {
			total += rightDistance(b, kw) * 10 // enforce strongly
		}
	}

	return total
}
