package gtree

type Pixel int

type TextElement struct {
	Text       string
	FontSize   Pixel
	LineHeight Pixel
}

type Layout interface {
	Height() Pixel
	Width() Pixel
	Margin() Pixel
	Title() TextElement
	Notes() []TextElement
	Blurbs() []*Blurb
	Connectors() []*Connector
	Debug() bool
}

type Point struct {
	X Pixel
	Y Pixel
}

type Connector struct {
	Points []Point
}

func distance(a, b *Blurb) int {
	return int((a.X() - b.X()) * (a.X() - b.X()))
}

func rightDistance(a, b *Blurb) int {
	if a.X() > b.X() {
		return 0
	}
	return int((a.X() - b.X()) * (a.X() - b.X()))
}

type Blurb struct {
	ID                int
	HeadingText       string
	HeadingFontSize   Pixel
	HeadingLineHeight Pixel
	DetailTexts       []string
	DetailFontSize    Pixel
	DetailLineHeight  Pixel

	// Text          []string
	CentreText          bool  // true if the text for this blurb is better presented as centred
	Width               Pixel // Width is the horizontal extent of the Blurb
	AbsolutePositioning bool  // when true, the position of the blurb is controlled by TopPos and LeftPos, otherwise it is calclated relative to neighbours
	TopPos              Pixel // TopPos is the absolute vertical position of the upper edge of the Blurb
	LeftPos             Pixel // LeftPos is the absolute horizontal position of the left edge of the Blurb
	Height              Pixel // Height is the vertical extent of the Blurb
	Col                 int   // column the blurb appears in for layouts that use columns
	Row                 int   // row the blurb appears in for layouts that use rows
	LeftPad             Pixel // required padding to left of blurb to separate families
	LeftShift           Pixel // optional padding that shifts blurb to right for alignment
	NoShift             bool  // when true the left shift will not be changed
	KeepWith            []*Blurb
	KeepRightOf         []*Blurb
	LeftNeighbour       *Blurb // the blurb to the left of this one, when non-nil will be used for horizontal positioning
	Parent              *Blurb
	LeftStop            *Blurb // the blurb whose center must not be passed when shifting left
	RightStop           *Blurb // the blurb whose center must not be passed when shifting right
	TopHookOffset       Pixel  // TopHookOffset is the offset from the left of the blurb where any dropped connecting line should finish (ensures it is within the bounds of the name, even if subsequent detail lines are longer)
	SideHookOffset      Pixel  // SideHookOffset is the offset from the top of the blurb where any connecting line should finish
}

// X returns the horizontal position of the centre of the Blurb
func (b *Blurb) X() Pixel {
	if b.AbsolutePositioning {
		return b.LeftPos + b.Width/2
	}
	left := Pixel(0)
	if b.LeftNeighbour != nil {
		left = b.LeftNeighbour.Right()
	}
	left += b.LeftPad
	left += b.LeftShift
	return left + b.Width/2
}

// Y returns the vertical position of the centre of the Blurb
func (b *Blurb) Y() Pixel {
	return b.TopPos + b.Height/2
}

// Left returns the horizontal position of the leftmost edge of the Blurb
func (b *Blurb) Left() Pixel {
	if b.AbsolutePositioning {
		return b.LeftPos
	}
	return b.X() - b.Width/2
}

// Right returns the horizontal position of the rightmost edge of the Blurb
func (b *Blurb) Right() Pixel {
	if b.AbsolutePositioning {
		return b.LeftPos + b.Width
	}
	return b.X() + b.Width/2
}

// Bottom returns the vertical position of the lower edge of the Blurb
func (b *Blurb) Bottom() Pixel {
	return b.TopPos + b.Height
}

func (b *Blurb) TopHookX() Pixel {
	return b.Left() + b.TopHookOffset
}

func (b *Blurb) SideHookY() Pixel {
	return b.TopPos + b.SideHookOffset
}

func textWidth(t []rune, fontSize Pixel) Pixel {
	w := Pixel(0)
	for _, r := range t {
		rw, ok := runeWidths[r]
		if ok {
			w += rw
		} else {
			w += fontSize
		}
	}
	if fontSize == 16 {
		return w
	}
	return Pixel(float64(w)*float64(fontSize)/16 + 0.5)
}

/*
** Table of scale-factor estimates for variable-width characters.
** Actual character widths vary by font.  These numbers are only
** guesses.  And this table only provides data for ASCII.
**
** 100 means normal width.
 */
var runeWidths = map[rune]Pixel{
	' ':  5,
	'!':  6,
	'"':  7,
	'#':  13,
	'$':  10,
	'%':  15,
	'&':  14,
	'\'': 4,
	'(':  6,
	')':  6,
	'*':  8,
	'+':  13,
	',':  5,
	'-':  5,
	'.':  5,
	'/':  5,
	'0':  10,
	'1':  10,
	'2':  10,
	'3':  10,
	'4':  10,
	'5':  10,
	'6':  10,
	'7':  10,
	'8':  10,
	'9':  10,
	':':  5,
	';':  5,
	'<':  13,
	'=':  13,
	'>':  13,
	'?':  9,
	'@':  16,
	'A':  12,
	'B':  12,
	'C':  12,
	'D':  13,
	'E':  12,
	'F':  11,
	'G':  13,
	'H':  14,
	'I':  6,
	'J':  6,
	'K':  12,
	'L':  11,
	'M':  16,
	'N':  14,
	'O':  13,
	'P':  11,
	'Q':  13,
	'R':  12,
	'S':  11,
	'T':  11,
	'U':  13,
	'V':  12,
	'W':  16,
	'X':  11,
	'Y':  12,
	'Z':  11,
	'[':  6,
	'\\': 5,
	']':  6,
	'^':  13,
	'_':  8,
	'`':  8,
	'a':  10,
	'b':  10,
	'c':  9,
	'd':  10,
	'e':  9,
	'f':  6,
	'g':  10,
	'h':  10,
	'i':  5,
	'j':  5,
	'k':  10,
	'l':  5,
	'm':  16,
	'n':  10,
	'o':  10,
	'p':  10,
	'q':  10,
	'r':  8,
	's':  8,
	't':  6,
	'u':  10,
	'v':  9,
	'w':  14,
	'x':  9,
	'y':  9,
	'z':  8,
	'{':  10,
	'|':  5,
	'}':  10,
	'~':  13,
}
