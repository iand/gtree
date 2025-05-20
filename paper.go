package gtree

// Millimetre represents a unit of measurement used for physical dimensions. The digital sizes
// are determined by scaling by a dpi (dots per inch)
type Millimetre float64

func (m Millimetre) Pixel(dpi int) Pixel {
	return Pixel(float64(m) * float64(dpi) / 25.4)
}

type PaperSize struct {
	Width  Millimetre // width in millimetres
	Height Millimetre // height in millimetres
}

func (ps PaperSize) ExtentAtDPI(dpi int) Extent {
	return Extent{
		Width:  Pixel(ps.Width.Pixel(dpi)),
		Height: Pixel(ps.Height.Pixel(dpi)),
	}
}

var (
	// ISO A Series Sizes
	PaperSizeA0Portrait  = PaperSize{Width: 841, Height: 1189}
	PaperSizeA0Landscape = PaperSize{Width: 1189, Height: 841}

	PaperSizeA1Portrait  = PaperSize{Width: 594, Height: 841}
	PaperSizeA1Landscape = PaperSize{Width: 841, Height: 594}

	PaperSizeA2Portrait  = PaperSize{Width: 420, Height: 594}
	PaperSizeA2Landscape = PaperSize{Width: 594, Height: 420}

	PaperSizeA3Portrait  = PaperSize{Width: 297, Height: 420}
	PaperSizeA3Landscape = PaperSize{Width: 420, Height: 297}

	PaperSizeA4Portrait  = PaperSize{Width: 210, Height: 297}
	PaperSizeA4Landscape = PaperSize{Width: 297, Height: 210}

	// North American Letter (8.5 × 11 inches)
	PaperSizeLetterPortrait  = PaperSize{Width: 215.9, Height: 279.4} // 8.5 × 11 inches
	PaperSizeLetterLandscape = PaperSize{Width: 279.4, Height: 215.9}
)
