package gtree

import (
	"bytes"
	"fmt"
)

// SVG generates an SVG (Scalable Vector Graphics) representation of the provided layout.
// It takes a Layout interface as input and returns a string containing the SVG markup, or an error if the generation fails.
//
// The SVG output includes:
// - The XML declaration and SVG root element with specified width and height based on the layout dimensions.
// - A white background covering the entire SVG canvas.
// - The title of the chart, if provided, rendered at the top of the SVG.
// - Any notes, rendered below the title, with appropriate spacing.
// - Blurbs representing individuals or family members, each with their associated text and optional background rectangle if debug mode is enabled.
// - Connectors, represented as paths, connecting blurbs according to their relationships.
//
// The function iterates over the layout elements (title, notes, blurbs, connectors), converts their properties to SVG-compatible attributes,
// and appends them to an internal buffer. Finally, it returns the complete SVG as a string.
func SVG(lay Layout) (string, error) {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\n")
	fmt.Fprintf(buf, "<svg width=\"%s\" height=\"%s\" xmlns=\"http://www.w3.org/2000/svg\">\n", length(lay.Width()), length(lay.Height()))

	// White background
	fmt.Fprintln(buf, `<rect width="100%" height="100%" fill="white"/>`)

	var y Pixel
	title := lay.Title()
	if title.Text != "" {
		fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"alphabetic\" text-anchor=\"start\" font-size=\"%dpx\" letter-spacing=\"0\">%s</text>\n", length(lay.Margin()), length(lay.Margin()+title.Style.LineHeight), title.Style.FontSize, title.Text)
		y += title.Style.LineHeight
	}

	notes := lay.Notes()
	for i := range notes {
		fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"alphabetic\" text-anchor=\"start\" font-size=\"%dpx\" letter-spacing=\"0\">%s</text>\n", length(lay.Margin()), length(lay.Margin()+notes[i].Style.LineHeight+y), notes[i].Style.FontSize, notes[i].Text)
		y += notes[i].Style.LineHeight
	}

	// Draw blurbs
	for _, b := range lay.Blurbs() {
		_ = b
		if lay.Debug() {
			fmt.Fprintf(buf, "<!-- blurb %s (left=%d, top=%d, width=%d, height=%d) -->\n", b.HeadingTexts.Lines[0], b.Left(), b.TopPos, b.Width, b.Height)
			fmt.Fprintf(buf, "<rect x=\"%s\" y=\"%s\" width=\"%s\" height=\"%s\" fill=\"#eeeeee\"/>", length(b.Left()), length(b.TopPos), length(b.Width), length(b.Height))
		}
		textAnchor := "start"
		textx := length(b.Left())
		if b.CentreText {
			textAnchor = "middle"
			textx = length(b.X())
		}
		fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"hanging\" text-anchor=\"%s\">\n", textx, length(b.TopPos), textAnchor)
		for _, line := range b.HeadingTexts.Lines {
			fmt.Fprintf(buf, "<tspan x=\"%s\" dy=\"%s\" font-size=\"%dpx\" fill=\"%s\">%s</tspan>\n", textx, length(b.HeadingTexts.Style.LineHeight), b.HeadingTexts.Style.FontSize, b.HeadingTexts.Style.Color, line)
		}
		for _, line := range b.DetailTexts.Lines {
			fmt.Fprintf(buf, "<tspan x=\"%s\" dy=\"%s\" font-size=\"%dpx\" fill=\"%s\">%s</tspan>\n", textx, length(b.DetailTexts.Style.LineHeight), b.DetailTexts.Style.FontSize, b.DetailTexts.Style.Color, line)
		}
		fmt.Fprintf(buf, "</text>\n")
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
