package gtree

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"strings"
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
func SVG(lay Layout, ps PaperSize) (string, error) {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\n")
	fmt.Fprintf(buf, "<svg viewBox=\"0 0 %s %s\" width=\"%gmm\" height=\"%gmm\" xmlns=\"http://www.w3.org/2000/svg\">\n", length(lay.Width()), length(lay.Height()), ps.Width, ps.Height)

	// embed fonts
	fonts := map[string][]byte{}
	for _, b := range lay.Blurbs() {
		for _, t := range b.Texts {
			if _, ok := fonts[t.Style.FontFamily]; ok {
				continue
			}
			fonts[t.Style.FontFamily] = t.Style.Font.Bytes
		}
	}
	if len(fonts) > 0 {
		fmt.Fprintf(buf, "<defs>\n")
		for family, data := range fonts {
			fmt.Fprintf(buf, "<style type=\"text/css\">\n")
			fmt.Fprintf(buf, "<![CDATA[\n")
			fmt.Fprintf(buf, "@font-face {\n")
			fmt.Fprintf(buf, "	font-family: '%s';\n", family)
			fmt.Fprintf(buf, "	src: url('data:application/x-font-ttf;base64,%s');\n", base64.StdEncoding.EncodeToString(data))
			fmt.Fprintf(buf, "	}\n")
			fmt.Fprintf(buf, "]]>\n")
			fmt.Fprintf(buf, "</style>\n")
		}
		fmt.Fprintf(buf, "</defs>\n")
	}

	// White background
	fmt.Fprintln(buf, `<rect width="100%" height="100%" fill="white"/>`)

	if lay.Background() != "" {
		fmt.Fprintln(buf, lay.Background())
	}

	// draw legend
	svgBlurb(buf, lay.Legend(), lay.Debug())

	// Draw blurbs
	for _, b := range lay.Blurbs() {
		svgBlurb(buf, b, lay.Debug())
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
		fmt.Fprintf(buf, "<path style=\"fill:none;fill-opacity:0.75000000;fill-rule:evenodd;stroke:#000000;stroke-width:1.0;stroke-linecap:butt;stroke-linejoin:miter;stroke-miterlimit:4.0000000;stroke-opacity:1.0000000\" d=\"%s\" />\n", data)
	}

	fmt.Fprintln(buf, "</svg>")

	return buf.String(), nil
}

func length(v Pixel) string {
	return fmt.Sprintf("%d", v)
}

func svgBlurb(buf *bytes.Buffer, b *Blurb, debug bool) {
	rotation := b.Rotation // in degrees
	transform := svgTransform(b.X, b.Y, 1, rotation)

	left := b.X - b.Width/2
	top := b.Y - b.Height/2

	if debug {
		fmt.Fprintf(buf, "<!-- blurb %s (left=%d, top=%d, width=%d, height=%d) -->\n", b.Texts[0].Lines[0], left, top, b.Width, b.Height)
		fmt.Fprintf(buf, "<rect x=\"%s\" y=\"%s\" width=\"%s\" height=\"%s\" fill=\"#eeeeee\"/>", length(left), length(top), length(b.Width), length(b.Height))
	}
	textAnchor := "start"
	if b.Alignment == AlignmentCenter {
		textAnchor = "middle"
	}
	fmt.Fprintf(buf, "<g transform=\"%s\">\n", transform)
	fmt.Fprintf(buf, "<text x=\"0\" y=\"%s\" dominant-baseline=\"auto\" text-anchor=\"%s\">\n", length(-b.Height/2), textAnchor)
	// fmt.Fprintf(buf, "<text x=\"%s\" y=\"%s\" dominant-baseline=\"hanging\" text-anchor=\"%s\">\n", textx, length(top), textAnchor)
	for _, t := range b.Texts {
		for _, line := range t.Lines {
			fmt.Fprintf(buf, "<tspan x=\"0\" dy=\"%s\" font-size=\"%dpx\" font-family=\"%s\" fill=\"%s\">%s</tspan>\n", length(t.Style.LineHeight), t.Style.FontSize, t.Style.FontFamily, t.Style.Color, html.EscapeString(line))
		}
	}
	fmt.Fprintf(buf, "</text>\n</g>\n")
}

func svgTransform(x, y Pixel, scale, rotateDeg float64) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("translate(%g %g)", float64(x), float64(y)))
	if rotateDeg != 0 {
		parts = append(parts, fmt.Sprintf("rotate(%g)", rotateDeg))
	}
	if scale != 1 {
		parts = append(parts, fmt.Sprintf("scale(%g)", scale))
	}
	return strings.Join(parts, " ")
}
