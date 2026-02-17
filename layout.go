package gtree

// Layout defines an interface for chart layouts, providing methods to retrieve dimensions, text elements,
// and layout components such as blurbs and connectors.
type Layout interface {
	Height() Pixel
	Width() Pixel
	Margin() Pixel
	Legend() *Blurb
	Blurbs() []*Blurb
	Connectors() []*Connector
	Debug() bool
	Background() string // raw svg to use as a background
	BackgroundColor() string
}

// Connector represents a connection between two or more points in the layout, typically used to draw lines between blurbs.
type Connector struct {
	Points []Point
}

type Alignment int

const (
	AlignmentLeft   Alignment = 0
	AlignmentCenter Alignment = 1
	AlignmentRight  Alignment = 2
)

// Blurb represents a visual element in the layout, typically used to display information about a person in a chart.
type Blurb struct {
	Texts     []TextSection
	Alignment Alignment

	X        Pixel   // the vertical position of the centre of the Blurb
	Y        Pixel   // the horizontal position of the centre of the Blurb
	Width    Pixel   // the horizontal extent of the Blurb
	Height   Pixel   // the vertical extent of the Blurb
	Rotation float64 //
}

// Pixel represents a unit of measurement used for digital dimensions, such as font sizes, margins, and positions.
type Pixel int

type TextSection struct {
	Lines []string
	Style TextStyle
}

// Point represents a coordinate in the layout, defined by its X (horizontal) and Y (vertical) position.
type Point struct {
	X Pixel
	Y Pixel
}

// Extent represents a size in the layout, defined by its width and height
type Extent struct {
	Width  Pixel
	Height Pixel
}

type TextStyle struct {
	FontSize   Pixel // FontSize is the size of the font to use for the text of each blurb.
	FontFamily string
	LineHeight Pixel // LineHeight is the vertical distance between lines of text of the same style.
	Font       *Font
	Color      string // Color is the color of the text. The default is black #000000.
}

func ScaleStyle(t *TextStyle, factor float64) {
	t.FontSize = Pixel(float64(t.FontSize) * factor)
	t.LineHeight = Pixel(float64(t.LineHeight) * factor)
}

func ScaleBlurb(b *Blurb, factor float64) {
	for i := range b.Texts {
		t := b.Texts[i]
		ScaleStyle(&t.Style, factor)
		b.Texts[i] = t
	}
}

func MeasureBlurb(b *Blurb) {
	b.Height = 0
	b.Width = 0

	for _, t := range b.Texts {
		b.Height += Pixel(float64(t.Style.LineHeight) * float64(len(t.Lines)))
		for i := range t.Lines {
			wl := Pixel(float64(t.Style.MeasureWidth(t.Lines[i])))
			if wl > b.Width {
				b.Width = wl
			}
		}
	}
}

func LeftAlignedLegend(topLeft Point, title string, titleStyle TextStyle, notes []string, noteStyle TextStyle) *Blurb {
	legend := newLegend(title, titleStyle, notes, noteStyle)
	legend.X = topLeft.X
	legend.Y = topLeft.Y + titleStyle.LineHeight/2
	return legend
}

func CenterAlignedLegend(topCenter Point, title string, titleStyle TextStyle, notes []string, noteStyle TextStyle) *Blurb {
	legend := newLegend(title, titleStyle, notes, noteStyle)
	legend.X = topCenter.X
	legend.Y = topCenter.Y + titleStyle.LineHeight/2
	legend.Alignment = AlignmentCenter
	return legend
}

func CenterAlignedTitle(topCenter Point, title string, preTitle string, postTitle string, titleStyle TextStyle, subTitleStyle TextStyle) *Blurb {
	b := &Blurb{
		Texts:     []TextSection{},
		Alignment: AlignmentCenter,
	}

	if preTitle != "" {
		b.Texts = append(b.Texts, TextSection{
			Lines: []string{preTitle},
			Style: subTitleStyle,
		})
	}
	if title != "" {
		b.Texts = append(b.Texts, TextSection{
			Lines: []string{title},
			Style: titleStyle,
		})
	}
	if postTitle != "" {
		b.Texts = append(b.Texts, TextSection{
			Lines: []string{postTitle},
			Style: subTitleStyle,
		})
	}

	b.X = topCenter.X
	b.Y = topCenter.Y + titleStyle.LineHeight/2

	MeasureBlurb(b)
	return b
}

func RightAlignedNotes(topRight Point, notes []string, noteStyle TextStyle) *Blurb {
	b := &Blurb{
		Texts: []TextSection{
			{
				Lines: notes,
				Style: noteStyle,
			},
		},
		Alignment: AlignmentRight,
	}
	MeasureBlurb(b)
	b.X = topRight.X - b.Width
	b.Y = topRight.Y - b.Height/2
	return b
}

func newLegend(title string, titleStyle TextStyle, notes []string, noteStyle TextStyle) *Blurb {
	legend := &Blurb{
		Texts: []TextSection{
			{
				Lines: []string{title},
				Style: titleStyle,
			},
		},
		Alignment: AlignmentLeft,
	}

	if len(notes) > 0 {
		legend.Texts = append(legend.Texts, TextSection{
			Lines: notes,
			Style: noteStyle,
		})
	}
	MeasureBlurb(legend)
	return legend
}
