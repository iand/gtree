package gtree

import (
	"fmt"
	"log/slog"
	"math"
	prand "math/rand"
	"strings"
	"sync"
)

type DescendantChart struct {
	Title string
	Notes []string
	Root  *DescendantPerson
}

type DescendantPerson struct {
	ID       int
	Details  []string
	Families []*DescendantFamily
}

type DescendantFamily struct {
	Other    *DescendantPerson
	Details  []string
	Children []*DescendantPerson
}

type LayoutOptions struct {
	FontSize         Pixel // size of the font to use for the main text of each blurb (first line)
	DetailFontSize   Pixel // size of the font to use for the detail text of each blurb (subsequent lines)
	TextLineHeight   Pixel // vertical distance between lines of text in heading of a blurb
	DetailLineHeight Pixel // vertical distance between lines of text in detail of a blurb
	DetailWrapWidth  Pixel // maximum width of detail text before wrapping to a new line
	Hspace           Pixel // horizontal spacing between blurbs within the same family
	LineWidth        Pixel
	Margin           Pixel // margin to add to entire drawing
	FamilyDrop       Pixel // length of line to draw down from parents to the children group line
	ChildDrop        Pixel // the length of the line drawn from the children group line to a child
	LineGap          Pixel // the distance to leave between a connecting line and any text
	Debug            bool  // emit logging and debug information
	TitleFontSize    Pixel // size of the font to use for the title of the chart
	NoteFontSize     Pixel // size of the font to use for the notes of the chart
	TitleLineHeight  Pixel // vertical distance to use for spacing the title of the chart
	NoteLineHeight   Pixel // vertical distance to use for spacing the notes of the chart
}

func DefaultLayoutOptions() *LayoutOptions {
	return &LayoutOptions{
		FontSize:         20,
		DetailFontSize:   16,
		TitleFontSize:    40,
		NoteFontSize:     20,
		TextLineHeight:   22,
		DetailLineHeight: 18,
		TitleLineHeight:  42,
		NoteLineHeight:   22,
		DetailWrapWidth:  18 * 16,
		Hspace:           16,
		LineWidth:        2,
		Margin:           16,
		FamilyDrop:       48,
		ChildDrop:        16,
		LineGap:          8,
	}
}

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
	initOnce   sync.Once
}

func (l *DescendantLayout) Width() Pixel { return l.width }

func (l *DescendantLayout) Height() Pixel { return l.height }

func (l *DescendantLayout) Margin() Pixel { return l.opts.Margin }

func (l *DescendantLayout) Title() TextElement {
	return TextElement{
		Text:       l.title,
		FontSize:   l.opts.TitleFontSize,
		LineHeight: l.opts.TitleLineHeight,
	}
}

func (l *DescendantLayout) Notes() []TextElement {
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

func (l *DescendantLayout) Blurbs() []*Blurb {
	bs := make([]*Blurb, 0, len(l.blurbs))
	for _, b := range l.blurbs {
		bs = append(bs, b)
	}
	return bs
}

func (l *DescendantLayout) Connectors() []*Connector {
	return l.connectors
}

func (l *DescendantLayout) Debug() bool { return l.opts.Debug }

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

func (l *DescendantLayout) newBlurb(id int, texts []string, row int, parent *Blurb) *Blurb {
	texts = l.wrapTexts(texts)
	b := &Blurb{
		ID: id,
		// Text:              texts,
		Row:               row,
		Parent:            parent,
		TopHookOffset:     l.opts.Hspace * 2,
		SideHookOffset:    l.opts.TextLineHeight / 2,
		HeadingFontSize:   l.opts.FontSize,
		DetailFontSize:    l.opts.DetailFontSize,
		HeadingLineHeight: l.opts.TextLineHeight,
		DetailLineHeight:  l.opts.DetailLineHeight,
	}
	if len(texts) > 0 {
		b.HeadingText = texts[0]
		b.Height = b.HeadingLineHeight
		b.Width = textWidth([]rune(b.HeadingText), b.HeadingFontSize)

		if len(texts) > 1 {
			b.DetailTexts = texts[1:]
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

	for len(l.rows) <= row {
		l.rows = append(l.rows, []*Blurb{})
	}
	l.rows[row] = append(l.rows[row], b)

	return b
}

func (l *DescendantLayout) wrapTexts(texts []string) []string {
	if len(texts) == 0 {
		return []string{}
	}
	wrapped := make([]string, 0, len(texts))
	wrapped = append(wrapped, texts[0])

	for i := 1; i < len(texts); i++ {
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

	th, _ := l.titleDimensions()
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

func (l *DescendantLayout) titleDimensions() (Pixel, Pixel) {
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
