package gtree

import (
	"bytes"
	"fmt"
)

func SVG(lay Layout) (string, error) {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\n")
	fmt.Fprintf(buf, "<svg width=\"%s\" height=\"%s\" xmlns=\"http://www.w3.org/2000/svg\">\n", length(lay.Width()), length(lay.Height()))

	// White background
	fmt.Fprintln(buf, `<rect width="100%" height="100%" fill="white"/>`)

	var y Pixel
	title := lay.Title()
	if title.Text != "" {
		fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"alphabetic\" text-anchor=\"start\" font-size=\"%dpx\" letter-spacing=\"0\">%s</text>\n", length(lay.Margin()), length(lay.Margin()+title.LineHeight), title.FontSize, title.Text)
		y += title.LineHeight
	}

	notes := lay.Notes()
	for i := range notes {
		fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"alphabetic\" text-anchor=\"start\" font-size=\"%dpx\" letter-spacing=\"0\">%s</text>\n", length(lay.Margin()), length(lay.Margin()+notes[i].LineHeight+y), notes[i].FontSize, notes[i].Text)
		y += notes[i].LineHeight
	}

	// Draw blurbs
	for _, b := range lay.Blurbs() {
		_ = b
		if lay.Debug() {
			fmt.Fprintf(buf, "<!-- blurb %s (left=%d, top=%d, width=%d, height=%d) -->\n", b.HeadingText, b.Left(), b.TopPos, b.Width, b.Height)
			fmt.Fprintf(buf, "<rect x=\"%s\" y=\"%s\" width=\"%s\" height=\"%s\" fill=\"#eeeeee\"/>", length(b.Left()), length(b.TopPos), length(b.Width), length(b.Height))
		}
		textAnchor := "start"
		textx := length(b.Left())
		if b.CentreText {
			textAnchor = "middle"
			textx = length(b.X())
		}
		if len(b.DetailTexts) == 0 {
			fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"hanging\" text-anchor=\"%s\" font-size=\"%dpx\" letter-spacing=\"0\">%s</text>\n", textx, length(b.TopPos), textAnchor, b.HeadingFontSize, b.HeadingText)
		} else {
			fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"hanging\" text-anchor=\"%s\">\n", textx, length(b.TopPos), textAnchor)
			fmt.Fprintf(buf, "<tspan x=\"%s\" dy=\"%s\" font-size=\"%dpx\">%s</tspan>\n", textx, length(b.HeadingLineHeight), b.HeadingFontSize, b.HeadingText)
			for _, line := range b.DetailTexts {
				fmt.Fprintf(buf, "<tspan x=\"%s\" dy=\"%s\" font-size=\"%dpx\">%s</tspan>\n", textx, length(b.DetailLineHeight), b.DetailFontSize, line)
			}
			fmt.Fprintf(buf, "</text>\n")
		}
	}

	// Add lines
	for _, b := range lay.Connectors() {
		var data string
		for i, p := range b.Points {
			if i == 0 {
				data = fmt.Sprintf("M %s,%s", length(p.X), length(p.Y))
				continue
			}
			data += fmt.Sprintf(" L %s,%s", length(p.X), length(p.Y))
		}
		fmt.Fprintf(buf, "<path style=\"fill:none;fill-opacity:0.75000000;fill-rule:evenodd;stroke:#000000;stroke-width:2.3750000;stroke-linecap:butt;stroke-linejoin:miter;stroke-miterlimit:4.0000000;stroke-opacity:1.0000000\" d=\"%s\" />\n", data)
	}

	fmt.Fprintln(buf, "</svg>")

	return buf.String(), nil
}

func length(v Pixel) string {
	return fmt.Sprintf("%d", v)
}
