package gtree

import (
	"fmt"
	"os"
	"sync"

	"github.com/flopp/go-findfont"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func DefaultSerifFontNames() []string {
	return []string{"Superclarendon", "Bookman Old Style", "URW Bookman", "URW Bookman L", "Georgia Pro", "Georgia", "serif"}
}

func NewTextStyle(opt TextStyleOption) (TextStyle, error) {
	ts := TextStyle{
		FontSize:   opt.FontSize,
		LineHeight: opt.LineHeight,
		Color:      opt.Color,
	}

	fd, err := loadFirstFont(opt.FontNames...)
	if err != nil {
		return ts, err
	}

	family, err := fd.Font.Name(nil, sfnt.NameIDFamily)
	if err != nil {
		return ts, err
	}

	dpi := 72.0
	pointSize := float64(opt.FontSize) * dpi / 72.0

	tf, err := opentype.NewFace(fd.Font, &opentype.FaceOptions{
		Size:    pointSize,
		DPI:     dpi,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return ts, fmt.Errorf("new face: %v", err)
	}

	ts.FontFamily = family
	ts.Font = &Font{
		Name:  family,
		Bytes: fd.Bytes,
		tf:    tf,
	}

	return ts, nil
}

func (ts *TextStyle) MeasureWidth(s string) Pixel {
	if ts.Font == nil {
		return fallbackTextWidth([]rune(s), ts.FontSize)
	}
	return ts.Font.MeasureWidth(s)
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

func fallbackTextWidth(t []rune, fontSize Pixel) Pixel {
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

type Font struct {
	Name  string
	Bytes []byte
	tf    font.Face
}

type FontData struct {
	Font  *opentype.Font
	Bytes []byte
}

var (
	fontCache   = map[string]*FontData{}
	fontCacheMu sync.Mutex
)

func loadFirstFont(fontNames ...string) (*FontData, error) {
	fontCacheMu.Lock()
	defer fontCacheMu.Unlock()

	for _, fname := range fontNames {
		if f, ok := fontCache[fname]; ok {
			return f, nil
		}

		fontPath, err := findfont.Find(fname)
		if err != nil {
			continue
		}

		// load the font with the freetype library
		data, err := os.ReadFile(fontPath)
		if err != nil {
			continue
		}
		tf, err := opentype.Parse(data)
		if err != nil {
			continue
		}

		fd := &FontData{
			Font:  tf,
			Bytes: data,
		}
		fontCache[fname] = fd
		return fd, nil
	}

	return nil, fmt.Errorf("no matching fonts found")
}

func (f *Font) MeasureWidth(s string) Pixel {
	var adv fixed.Int26_6
	prevC := rune(-1)
	for _, c := range s {
		if prevC >= 0 {
			adv += f.tf.Kern(prevC, c)
		}
		a, _ := f.tf.GlyphAdvance(c)
		adv += a
		prevC = c
	}
	return Pixel(adv.Ceil())
}
