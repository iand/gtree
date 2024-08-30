package gtree

import "strings"

// Pixel represents a unit of measurement used for layout dimensions, such as font sizes, margins, and positions.
type Pixel int

// TextElement represents a piece of text with a specified font size and line height, used in various layout elements.
type TextElement struct {
	Text  string
	Style TextStyle
}

// Layout defines an interface for chart layouts, providing methods to retrieve dimensions, text elements,
// and layout components such as blurbs and connectors.
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

// Point represents a coordinate in the layout, defined by its X (horizontal) and Y (vertical) position.
type Point struct {
	X Pixel
	Y Pixel
}

// Connector represents a connection between two points in the layout, typically used to draw lines between blurbs.
type Connector struct {
	Points []Point
}

// Blurb represents a visual element in the layout, typically used to display information about a person in a chart.
// It includes various properties to control its positioning, text content, and relationships with other blurbs.
type Blurb struct {
	ID           int
	HeadingTexts TextSection
	DetailTexts  TextSection
	Tags         []string

	// Text          []string
	CentreText          bool   // true if the text for this blurb is better presented as centred
	Width               Pixel  // Width is the horizontal extent of the Blurb
	AbsolutePositioning bool   // when true, the position of the blurb is controlled by TopPos and LeftPos, otherwise it is calculated relative to neighbours
	TopPos              Pixel  // TopPos is the absolute vertical position of the upper edge of the Blurb
	LeftPos             Pixel  // LeftPos is the absolute horizontal position of the left edge of the Blurb
	Height              Pixel  // Height is the vertical extent of the Blurb
	Col                 int    // column the blurb appears in for layouts that use columns
	Row                 int    // row the blurb appears in for layouts that use rows
	LeftPad             Pixel  // required padding to left of blurb to separate families
	LeftShift           Pixel  // optional padding that shifts blurb to right for alignment
	NoShift             bool   // when true the left shift will not be changed
	KeepTightRight      *Blurb // the blurb to the right that this blurb should keep as close as possible to
	LeftNeighbour       *Blurb // the blurb to the left of this one, when non-nil will be used for horizontal positioning
	Parent              *Blurb
	TopHookOffset       Pixel // TopHookOffset is the offset from the left of the blurb where any dropped connecting line should finish (ensures it is within the bounds of the name, even if subsequent detail lines are longer)
	SideHookOffset      Pixel // SideHookOffset is the offset from the top of the blurb where any connecting line should finish

	FirstChild *Blurb
	LastChild  *Blurb
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

type TextStyle struct {
	FontSize   Pixel  // FontSize is the size of the font to use for the text of each blurb.
	LineHeight Pixel  // LineHeight is the vertical distance between lines of text of the same style.
	Color      string // Color is the color of the text. The default is black #000000.
}

type TextSection struct {
	Lines []string
	Style TextStyle
}

func wrapText(texts []string, maxWidth Pixel, fontSize Pixel) []string {
	if len(texts) == 0 {
		return []string{}
	}
	wrapped := make([]string, 0, len(texts))
	for i := 0; i < len(texts); i++ {
		wl := textWidth([]rune(texts[i]), fontSize)
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
			wl := textWidth([]rune(candidate), fontSize)
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
		w = textWidth([]rune(title), titleStyle.FontSize)
	}

	if len(notes) != 0 {
		h += noteStyle.LineHeight * Pixel(len(notes))
		for i := 0; i < len(notes); i++ {
			wl := textWidth([]rune(notes[i]), noteStyle.FontSize)
			if wl > w {
				w = wl
			}
		}
	}

	return h, w
}
