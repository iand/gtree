package gtree

import (
	"bytes"
	"fmt"
	"html"
	"text/template"
)

// ButterflyChart represents a horizontal butterfly chart, where the parents of the root
// person are positioned in the centre. Successive paternal generations extend to the
// left and maternal generations extend to the right. Ancestors in each generation are
// aligned vertically, visually depicting the lineage from the root person to their
// ancestors.
type ButterflyChart struct {
	TitleLine1 string
	TitleLine2 string
	Note       string
	Root       *ButterflyPerson
}

// ButterflyPerson represents an individual in the butterfly chart, including their ID, details, and their parents.
type ButterflyPerson struct {
	ID            int
	Forenames     string
	Surname       string
	DetailLine1   string
	DetailLine2   string
	MarriageLine1 string // only used for male
	MarriageLine2 string
	Father        *ButterflyPerson
	Mother        *ButterflyPerson
}

// ButterflyLayoutOptions defines various layout parameters for rendering the butterfly chart.
type ButterflyLayoutOptions struct {
	Debug bool

	// LineWidth Pixel // width of any drawn lines
	// Margin    Pixel // margin to add to entire drawing
	// Hspace    Pixel // the horizontal space to leave between blurbs in different generations
	// Vspace    Pixel // the vertical space to leave between blurbs in the same generation
	// LineGap   Pixel // the distance to leave between a connecting line and any text

	// HookLength Pixel // the length of the line drawn from the parent or a child to the vertical line that joins them

	// TitleStyle   TextStyleOption // TitleStyle is the style of the font to use for the title of the chart.
	// NoteStyle    TextStyleOption // NoteStyle is the style of the font to use for the notes of the chart.
	// HeadingStyle TextStyleOption // HeadingStyle is the style of the font to use for the first line of each blurb.
	// DetailStyle  TextStyleOption // DetailStyle is the style of the font to use for the subsequent lines of each blurb after the first.

	// DetailWrapWidth Pixel // DetailWrapWidth is the maximum width of detail text before wrapping to a new line.

	// Horizontals are x coordinates by generation
	// if paternal then use 5-gen to find coord
	// if maternal then use 6+gen to find coord
	Horizontals [12]float64

	// Verticals are y coordinates by generation, 6 generations
	Verticals [6][]float64

	// BoxWidths are widths of each box, by generation
	BoxWidths [6]float64

	// BoxHeights are heights of each box, by generation
	BoxHeights [6]float64

	// AscenderHorizontals are the horizontal positions of each ascender, by generation
	// if paternal then use 4-gen to find coord
	// if maternal then use 5+gen to find coord
	AscenderHorizontals [10]float64

	// AscenderVerticals are the vertical psotions of each ascender by generation, 5 generations
	AscenderVerticals       [5][]float64
	AscenderVerticalOffset  [5]float64
	AscenderVerticalSpacing [5]float64
}

// DefaultButterflyLayoutOptions returns the default layout options for rendering the butterfly chart.
func DefaultButterflyLayoutOptions() *ButterflyLayoutOptions {
	opts := &ButterflyLayoutOptions{}

	opts.BoxWidths = [6]float64{82, 82, 68, 56, 48, 48}
	opts.BoxHeights = [6]float64{48, 48, 40, 32, 27, 18}

	// horizontals are x coordinates by generation
	// if paternal then use 5-gen to find coord
	// if maternal then use 6+gen to find coord
	// opts.Horizontals = [12]float64{
	// 	// 54.878864, // paternal gen 5
	// 	// 132.57796, // paternal gen 4
	// 	// 225.09232, // paternal gen 3
	// 	// 327.79849, // paternal gen 2
	// 	// 464.64127, // paternal gen 1
	// 	// 551.19366, // paternal gen 0
	// 	// 645.55109, // maternal gen 0
	// 	// 727.18555, // maternal gen 1
	// 	// 862.97327, // maternal gen 2
	// 	// 966.63647, // maternal gen 3
	// 	// 1060.2081, // maternal gen 4
	// 	// 1137.0238, // maternal gen 5
	// }

	opts.Horizontals = [12]float64{
		28,       // paternal gen 5
		108 - 2,  // paternal gen 4
		202 - 4,  // paternal gen 3
		302 - 6,  // paternal gen 2
		438 - 6,  // paternal gen 1
		524 - 6,  // paternal gen 0
		617 - 6,  // maternal gen 0
		703 - 9,  // maternal gen 1
		839 - 4,  // maternal gen 2
		939 - 4,  // maternal gen 3
		1033 - 4, // maternal gen 4
		1113,     // maternal gen 5
	}

	// generation 0
	// just one vertical for p or m
	// opts.Verticals[0] = []float64{413.35709}
	opts.Verticals[0] = []float64{396.5}

	// generation 1
	// two verticals for pp and pm, or mp and mm
	// opts.Verticals[1] = []float64{209.27454, 618.19}
	opts.Verticals[1] = []float64{
		196.5,
		596.5,
	}

	// generation 2
	// four verticals for ppp, ppm, pmp and pmm
	// opts.Verticals[2] = []float64{
	// 	108.44363, 312.19366,
	// 	517.90063, 721.05359,
	// }
	opts.Verticals[2] = []float64{
		100.5,
		300.5,
		500.5,
		700.5,
	}

	// generation 3
	// eight verticals
	// opts.Verticals[3] = []float64{
	// 	57.63866, 159.97037, 261.98456, 364.12625,
	// 	466.68274, 569.81989, 671.15948, 773.16248,
	// }

	// generation 3
	// eight verticals
	opts.Verticals[3] = []float64{
		54.5,
		154.5,
		254.5,
		354.5,
		454.5,
		554.5,
		654.5,
		754.5,
	}

	// // generation 4
	// // sixteen verticals
	// opts.Verticals[4] = []float64{
	// 	34.647652, 76.860626, 136.7316, 178.61171,
	// 	238.86554, 289.78601, 341.00723, 392.07809,
	// 	442.43933, 493.61765, 545.13873, 596.36145,
	// 	639.51832, 697.98175, 749.57556, 792.73086,
	// }

	// generation 4
	// sixteen verticals
	opts.Verticals[4] = []float64{
		32,
		82,
		132,
		182,
		232,
		282,
		332,
		382,
		432,
		482,
		532,
		582,
		632,
		682,
		732,
		782,
	}

	// generation 5
	// thirty-two verticals
	// opts.Verticals[5] = []float64{
	// 	24.760853, 47.891567, 76.860626, 100.14864, 126.75223, 150.24899, 178.61171, 201.96101,
	// 	0, 0, 280.29141, 303.32248, 331.0275, 354.45291, 382.09274, 405.46417,
	// 	433.36111, 456.58325, 484.2146, 507.5788, 0, 0, 0, 0,
	// 	0, 0, 688.51782, 711.88928, 0, 0, 0, 0,
	// }

	// margin of 25 top and bottom
	// each box has height 18
	// 5 between boxes for couple pairs
	// 9 between couple pairs
	opts.Verticals[5] = []float64{
		25,
		48,
		75,
		98,
		125,
		148,
		175,
		198,
		225,
		248,
		275,
		298,
		325,
		348,
		375,
		398,
		425,
		448,
		475,
		498,
		525,
		548,
		575,
		598,
		625,
		648,
		675,
		698,
		725,
		748,
		775,
		798,
	}

	// these are distorted since they are subject to the scale factor used
	opts.AscenderHorizontals = [10]float64{
		83 - 2,   // paternal gen 4
		180 - 4,  // paternal gen 3
		345 - 6,  // paternal gen 2
		380 - 6,  // paternal gen 1
		528 - 6,  // paternal gen 0
		694 - 6,  // maternal gen 0
		836 - 6,  // maternal gen 1
		1192 - 6, // maternal gen 2
		1100 - 4, // maternal gen 3
		1078 - 4, // maternal gen 4
	}

	opts.AscenderVerticalOffset[0] = 210
	opts.AscenderVerticalSpacing[0] = 0

	opts.AscenderVerticalOffset[1] = 313
	opts.AscenderVerticalSpacing[1] = 400

	opts.AscenderVerticalOffset[2] = 119
	opts.AscenderVerticalSpacing[2] = 153

	opts.AscenderVerticalOffset[3] = 82
	opts.AscenderVerticalSpacing[3] = 91

	opts.AscenderVerticalOffset[4] = 57
	opts.AscenderVerticalSpacing[4] = 51

	return opts
}

// boxBounds returns the bounds of a boxLleft, top, width, height
func (opts *ButterflyLayoutOptions) boxBounds(id string) (float64, float64, float64, float64) {
	left := 0.0

	generation := len(id) - 1
	if id[0] == 'p' {
		left = opts.Horizontals[5-generation]
	} else {
		left = opts.Horizontals[6+generation]
	}

	vidx := 0
	for i := 1; i < len(id); i++ {
		vidx *= 2
		if id[i] == 'm' {
			vidx += 1
		}
	}

	top := opts.Verticals[generation][vidx]

	return left, top, opts.BoxWidths[generation], opts.BoxHeights[generation]
}

// RenderSVG renders butterfly chart as SVG based on the provided options.
func (ch *ButterflyChart) RenderSVG(opts *ButterflyLayoutOptions) (string, error) {
	width := "1189mm"
	height := "841mm"

	buf := new(bytes.Buffer)

	if opts.Debug {
		ch.generateDebug(opts)
	}

	fmt.Fprintf(buf, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>")
	fmt.Fprintln(buf)
	fmt.Fprintf(buf, `<svg viewBox="0 0 1189 841" width="%s" height="%s" xmlns="http://www.w3.org/2000/svg">`, width, height)
	fmt.Fprintln(buf)

	ch.renderBackground(opts, buf)

	type entry struct {
		id     string
		person *ButterflyPerson
	}

	stack := []*entry{}

	if ch.Root.Father != nil {
		stack = append(stack, &entry{id: "p", person: ch.Root.Father})
	}

	if ch.Root.Mother != nil {
		stack = append(stack, &entry{id: "m", person: ch.Root.Mother})
	}

	for len(stack) > 0 {
		e := stack[0]
		stack = stack[1:]

		ch.renderPersonBlurb(opts, buf, e.id, e.person)
		if e.person == nil {
			if len(e.id) < 6 {
				stack = append(stack, &entry{id: e.id + "p", person: nil})
				stack = append(stack, &entry{id: e.id + "m", person: nil})
				ch.renderAscender(opts, buf, e.id, true)
			}
			continue
		}
		if e.person.MarriageLine1 != "" {
			ch.renderMarriage(opts, buf, e.id)
		}
		if len(e.id) < 6 {
			stack = append(stack, &entry{id: e.id + "p", person: e.person.Father})
			stack = append(stack, &entry{id: e.id + "m", person: e.person.Mother})

			if e.person.Father != nil || e.person.Mother != nil {
				ch.renderAscender(opts, buf, e.id, false)
			} else {
				ch.renderAscender(opts, buf, e.id, true)
			}

		}

	}

	ch.renderTitle(opts, buf)

	fmt.Fprintln(buf, "</svg>")

	return buf.String(), nil
}

func (r *ButterflyChart) renderPersonBlurb(opts *ButterflyLayoutOptions, buf *bytes.Buffer, id string, person *ButterflyPerson) {
	type paramtype struct {
		X             float64
		Y             float64
		Id            string
		Line1text     string
		Line2text     string
		Line3text     string
		Line4text     string
		Line1height   float64
		Line1fontsize float64
		Line1y        float64
		Line2height   float64
		Line2fontsize float64
		Line2y        float64
		Line3height   float64
		Line3fontsize float64
		Line3y        float64
		Line4height   float64
		Line4fontsize float64
		Line4y        float64
		CentreX       float64
		TextOffsetY   float64

		StrokeColor         string
		BoxWidth            float64
		BoxHeight           float64
		BoxDecorationSize   float64
		BoxDecorationPathTL string
		BoxDecorationPathTR string
		BoxDecorationPathBL string
		BoxDecorationPathBR string
	}

	var params paramtype
	var tmpl string
	var tmplText string

	tmplBox := `<rect
       style="fill:#ffffff;fill-opacity:1;stroke:{{- .StrokeColor -}};stroke-width:0.8;stroke-linecap:butt;stroke-linejoin:miter;stroke-miterlimit:11.1;stroke-dasharray:none;stroke-opacity:1"
       id="person-{{- .Id -}}-box"
       width="{{- .BoxWidth -}}"
       height="{{- .BoxHeight -}}"
       x="0"
       y="0"
        />

	    <path
	       style="fill:none;fill-opacity:1;stroke:{{- .StrokeColor -}};stroke-width:0.499999;stroke-linecap:butt;stroke-linejoin:miter;stroke-miterlimit:11.1;stroke-dasharray:none;stroke-opacity:1"
	       d="{{- .BoxDecorationPathTL -}}"  id="person-{{- .Id -}}-box-tl"/>

	    <path
	       style="fill:none;fill-opacity:1;stroke:{{- .StrokeColor -}};stroke-width:0.499999;stroke-linecap:butt;stroke-linejoin:miter;stroke-miterlimit:11.1;stroke-dasharray:none;stroke-opacity:1"
	       d="{{- .BoxDecorationPathTR -}}"  id="person-{{- .Id -}}-box-tr"/>

	    <path
	       style="fill:none;fill-opacity:1;stroke:{{- .StrokeColor -}};stroke-width:0.499999;stroke-linecap:butt;stroke-linejoin:miter;stroke-miterlimit:11.1;stroke-dasharray:none;stroke-opacity:1"
	       d="{{- .BoxDecorationPathBL -}}"  id="person-{{- .Id -}}-box-bl"/>

	    <path
	       style="fill:none;fill-opacity:1;stroke:{{- .StrokeColor -}};stroke-width:0.499999;stroke-linecap:butt;stroke-linejoin:miter;stroke-miterlimit:11.1;stroke-dasharray:none;stroke-opacity:1"
	       d="{{- .BoxDecorationPathBR -}}" id="person-{{- .Id -}}-box-br" />`

	tmplFourLines := `<text
       xml:space="preserve"
       style="font-family:Georgia;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none"
       id="person-{{- .Id -}}-text"><tspan
         style="font-size:{{- .Line1fontsize -}}px;line-height:{{- .Line1height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line1y -}}">{{- .Line1text -}}</tspan><tspan
         style="font-size:{{- .Line2fontsize -}}px;line-height:{{- .Line2height -}};font-weight:bold;text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line2y -}}">{{- .Line2text -}}</tspan><tspan
         style="font-size:{{- .Line3fontsize -}}px;line-height:{{- .Line3height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line3y -}}">{{- .Line3text -}}</tspan><tspan
         style="font-size:{{- .Line4fontsize -}}px;line-height:{{- .Line4height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line4y -}}">{{- .Line4text -}}</tspan></text>`

	tmplThreeLines := `<text
       xml:space="preserve"
       style="font-family:Georgia;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none"
       id="person-{{- .Id -}}-text"><tspan
         style="font-size:{{- .Line1fontsize -}}px;line-height:{{- .Line1height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line1y -}}">{{- .Line1text -}}</tspan><tspan
         style="font-size:{{- .Line2fontsize -}}px;line-height:{{- .Line2height -}};font-weight:bold;text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line2y -}}">{{- .Line2text -}}</tspan><tspan
         style="font-size:{{- .Line3fontsize -}}px;line-height:{{- .Line3height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line3y -}}">{{- .Line3text -}}</tspan></text>`

	generation := len(id) - 1

	left, top, width, height := opts.boxBounds(id)
	params.X = left
	params.Y = top
	params.BoxWidth = width
	params.BoxHeight = height

	params.CentreX = params.BoxWidth / 2

	var line1Len int
	var line2Len int
	var line3Len int
	var line4Len int
	if person != nil {
		line1Len = len(person.Forenames)
		line2Len = len(person.Surname)
		line3Len = len(person.DetailLine1)
		line4Len = len(person.DetailLine2)
	}

	switch generation {
	case 0: // parents, largest
		tmplText = tmplFourLines

		params.Line1height = 1.25
		params.Line1fontsize = scaleFont(9.87778, 14, line1Len)
		params.Line2height = 1.25
		params.Line2fontsize = scaleFont(9.87778, 10, line2Len)
		params.Line3height = 1.22
		params.Line3fontsize = scaleFont(6.7, 22, max(line3Len, line4Len))
		params.Line4height = params.Line3height
		params.Line4fontsize = params.Line3fontsize

		params.TextOffsetY = 14
		params.BoxDecorationSize = 6

	case 1: // grandparents
		tmplText = tmplFourLines

		params.Line1height = 1.25
		params.Line1fontsize = scaleFont(9.87778, 14, line1Len)
		params.Line2height = 1.25
		params.Line2fontsize = scaleFont(9.87778, 10, line2Len)
		params.Line3height = 1.2
		params.Line3fontsize = scaleFont(6.7, 22, max(line3Len, line4Len))
		params.Line4height = params.Line3height
		params.Line4fontsize = params.Line3fontsize

		params.TextOffsetY = 14
		params.BoxDecorationSize = 6

		tmplText = `<text
       xml:space="preserve"
       style="font-family:Georgia;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none"
       id="person-{{- .Id -}}-text"><tspan
         style="font-size:{{- .Line1fontsize -}}px;line-height:{{- .Line1height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line1y -}}">{{- .Line1text -}}</tspan><tspan
         style="font-size:{{- .Line2fontsize -}}px;line-height:{{- .Line2height -}};font-weight:bold;text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line2y -}}">{{- .Line2text -}}</tspan><tspan
         style="font-size:{{- .Line3fontsize -}}px;line-height:{{- .Line3height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line3y -}}">{{- .Line3text -}}</tspan><tspan
         style="font-size:{{- .Line4fontsize -}}px;line-height:{{- .Line4height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line4y -}}">{{- .Line4text -}}</tspan></text>`

	case 2: // great grandparents
		tmplText = tmplFourLines

		params.Line1height = 1.25
		params.Line1fontsize = scaleFont(8.46667, 14, line1Len)
		params.Line2height = 1.25
		params.Line2fontsize = scaleFont(8.46667, 10, line2Len)
		params.Line3height = 1.2
		params.Line3fontsize = scaleFont(5.3, 24, max(line3Len, line4Len))
		params.Line4height = params.Line3height
		params.Line4fontsize = params.Line3fontsize

		params.CentreX = 34
		params.TextOffsetY = 12
		params.BoxDecorationSize = 5

		tmplText = `<text
       xml:space="preserve"
       style="font-family:Georgia;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none"
       id="person-{{- .Id -}}-text"><tspan
         style="font-size:{{- .Line1fontsize -}}px;line-height:{{- .Line1height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line1y -}}">{{- .Line1text -}}</tspan><tspan
         style="font-size:{{- .Line2fontsize -}}px;line-height:{{- .Line2height -}};font-weight:bold;text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line2y -}}">{{- .Line2text -}}</tspan><tspan
         style="font-size:{{- .Line3fontsize -}}px;line-height:{{- .Line3height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line3y -}}">{{- .Line3text -}}</tspan><tspan
         style="font-size:{{- .Line4fontsize -}}px;line-height:{{- .Line4height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line4y -}}">{{- .Line4text -}}</tspan></text>`

	case 3: // great great great grandparents
		tmplText = tmplFourLines

		params.Line1height = 1.25
		params.Line1fontsize = scaleFont(7.76111, 9, line1Len)
		params.Line2height = 1.25
		params.Line2fontsize = scaleFont(7.76111, 9, line2Len)
		params.Line3height = 1.2
		params.Line3fontsize = scaleFont(5, 20, max(line3Len, line4Len))
		params.Line4height = params.Line3height
		params.Line4fontsize = params.Line3fontsize

		params.CentreX = 28
		params.TextOffsetY = 8
		params.BoxDecorationSize = 4

		tmplText = `<text
       xml:space="preserve"
       style="font-family:Georgia;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none"
       id="person-{{- .Id -}}-text"><tspan
         style="font-size:{{- .Line1fontsize -}}px;line-height:{{- .Line1height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line1y -}}">{{- .Line1text -}}</tspan><tspan
         style="font-size:{{- .Line2fontsize -}}px;line-height:{{- .Line2height -}};font-weight:bold;text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line2y -}}">{{- .Line2text -}}</tspan><tspan
         style="font-size:{{- .Line3fontsize -}}px;line-height:{{- .Line3height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line3y -}}">{{- .Line3text -}}</tspan><tspan
         style="font-size:{{- .Line4fontsize -}}px;line-height:{{- .Line4height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line4y -}}">{{- .Line4text -}}</tspan></text>`

	case 4: // great great great great grandparents
		tmplText = tmplFourLines

		params.Line1height = 1.25
		params.Line1fontsize = scaleFont(5.6, 10, line1Len)
		params.Line2height = 1.25
		params.Line2fontsize = scaleFont(5.6, 10, line2Len)
		params.Line3height = 1.2
		params.Line3fontsize = scaleFont(4.58611, 18, max(line3Len, line4Len))
		params.Line4height = params.Line3height
		params.Line4fontsize = params.Line3fontsize

		params.CentreX = 24
		params.TextOffsetY = 6
		params.BoxDecorationSize = 3

		tmplText = `<text
       xml:space="preserve"
       style="font-family:Georgia;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none"
       id="person-{{- .Id -}}-text"><tspan
         style="font-size:{{- .Line1fontsize -}}px;line-height:{{- .Line1height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line1y -}}">{{- .Line1text -}}</tspan><tspan
         style="font-size:{{- .Line2fontsize -}}px;line-height:{{- .Line2height -}};font-weight:bold;text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line2y -}}">{{- .Line2text -}}</tspan><tspan
         style="font-size:{{- .Line3fontsize -}}px;line-height:{{- .Line3height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line3y -}}">{{- .Line3text -}}</tspan><tspan
         style="font-size:{{- .Line4fontsize -}}px;line-height:{{- .Line4height -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .CentreX -}}" y="{{- .Line4y -}}">{{- .Line4text -}}</tspan></text>`

	case 5: // great great great great great grandparents
		tmplText = tmplThreeLines

		params.Line1height = 1.25
		params.Line1fontsize = scaleFont(4.58611, 13, line1Len)
		params.Line2height = 1.25
		params.Line2fontsize = scaleFont(4.58611, 13, line2Len) // does about 13 uppercase characters at width 48
		params.Line3height = 1.25
		params.Line3fontsize = scaleFont(4.58611, 13, max(line3Len, line4Len))
		params.Line4height = params.Line3height
		params.Line4fontsize = params.Line3fontsize

		params.CentreX = 24
		params.TextOffsetY = 5
		params.BoxDecorationSize = 3

	default:
		panic(fmt.Sprintf("unsupported person generation: %d", generation))
	}

	params.Id = id
	params.StrokeColor = "#808080"
	if person != nil {
		params.StrokeColor = "#000000"
		params.Line1text = html.EscapeString(person.Forenames)
		params.Line2text = html.EscapeString(person.Surname)
		params.Line3text = html.EscapeString(person.DetailLine1)
		params.Line4text = html.EscapeString(person.DetailLine2)
	}

	params.BoxDecorationPathTL = fmt.Sprintf("M 1,%f v -%f h %f", params.BoxDecorationSize+1, params.BoxDecorationSize, params.BoxDecorationSize)
	params.BoxDecorationPathTR = fmt.Sprintf("M %f,%f v -%f h -%f", params.BoxWidth-1, params.BoxDecorationSize+1, params.BoxDecorationSize, params.BoxDecorationSize)
	params.BoxDecorationPathBL = fmt.Sprintf("M 1,%f v %f h %f", params.BoxHeight-params.BoxDecorationSize-1, params.BoxDecorationSize, params.BoxDecorationSize)
	params.BoxDecorationPathBR = fmt.Sprintf("M %f,%f v %f h -%f", params.BoxWidth-1, params.BoxHeight-params.BoxDecorationSize-1, params.BoxDecorationSize, params.BoxDecorationSize)

	params.Line1y = params.TextOffsetY
	params.Line2y = params.Line1y + params.Line2height*params.Line2fontsize
	params.Line3y = params.Line2y + params.Line3height*params.Line3fontsize
	params.Line4y = params.Line3y + params.Line4height*params.Line4fontsize

	tmpl = `<g id="person-{{- .Id -}}" transform="translate({{- .X -}}, {{- .Y -}})">` + "\n"
	tmpl += tmplBox
	tmpl += tmplText
	tmpl += `</g>` + "\n"

	t, err := template.New("person").Parse(tmpl)
	if err != nil {
		panic(err)
	}
	err = t.Execute(buf, params)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(buf)
}

func (r *ButterflyChart) renderAscender(opts *ButterflyLayoutOptions, buf *bytes.Buffer, id string, disabled bool) {
	var tmpl string

	tmplBrace := `<text
	   transform="scale({{- .ScaleX -}},{{- .ScaleY -}})"
       xml:space="preserve"
       style="font-size:{{- .FontSize -}}px;fill:#ffffff;stroke:{{- .StrokeColor -}};stroke-width:0.999998;stroke-linecap:round;stroke-miterlimit:11.1;stroke-opacity:1"
       x="{{- .X -}}"
       y="{{- .Y -}}"
       >{{- .Char -}}</text>`

	type paramtype struct {
		Id          string
		Y           float64
		X           float64
		Char        string
		FontSize    float64
		ScaleX      float64
		ScaleY      float64
		StrokeColor string
	}
	var params paramtype

	params.Id = id
	params.ScaleX = 1
	params.ScaleY = 1
	params.StrokeColor = "#000000"
	if disabled {
		params.StrokeColor = "#808080"
	}

	// generation is the generation that the ascender starts from (i.e. the pointy side)
	// ranges from 0 to 4
	generation := len(id) - 1

	vidx := 0
	for i := 1; i < len(id); i++ {
		vidx *= 2
		if id[i] == 'm' {
			vidx += 1
		}
	}

	params.Y = opts.AscenderVerticalOffset[generation] + opts.AscenderVerticalSpacing[generation]*float64(vidx)
	if id[0] == 'p' {
		params.Char = "}"
		params.X = opts.AscenderHorizontals[4-generation]
	} else {
		params.X = opts.AscenderHorizontals[5+generation]
		params.Char = "{"
	}
	switch generation {
	case 0:
		sign := ""
		if id[0] == 'm' {
			sign = "-"
		}
		tmpl = fmt.Sprintf(`<g transform="matrix(%s1,0,0,1,{{- .X -}},{{- .Y -}})" id="ascender-{{- .Id -}}"
       style="fill:#ffffff;stroke:{{- .StrokeColor -}};stroke-width:1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none;stroke-dashoffset:0;stroke-opacity:1">
       <path
       d="m 0,0 c 0,0 20.81637,-2.68514 20.55624,23.11572 -0.34191,33.91298 -66.55112,63.66026 -65.48777,118.60097 1.06335,54.94069 34.44668,65.43914 34.44668,65.43914 l 0.10998,10.83706 c 0,0 -33.79698,10.58733 -34.86032,65.52803 -1.06335,54.94069 64.96376,82.75079 65.30567,116.66377 0.26013,25.80086 -20.55573,25.05292 -20.55573,25.05292 l 0.25994,-11.82099 c 0,0 10.48178,0.2953 10.45711,-17.38413 -0.0308,-22.08563 -65.49828,-52.07936 -64.50054,-114.77707 -0.30269,-34.50049 17.30916,-60.37152 31.82751,-68.77588 -14.49094,-8.20589 -31.82202,-34.45417 -31.52333,-68.49736 -0.99774,-62.6977 63.20695,-94.13618 63.01756,-115.8948 -0.13814,-15.87041 -9.01288,-16.2664 -9.01288,-16.2664 z"
     /></g>`, sign)

	case 1:
		sign := ""
		if id[0] == 'p' {
			sign = "-"
		}

		tmpl = fmt.Sprintf(`<g transform="matrix(%s0.68276199,0,0,1.2464835,{{- .X -}},{{- .Y -}})" id="ascender-{{- .Id -}}"
       style="font-size:185.97px;fill:#ffffff;stroke:{{- .StrokeColor -}};stroke-width:1.08397;stroke-linecap:round;stroke-miterlimit:11.1">
      <path
         d="m 0,0 v 7.21984 h -5.6299 c -15.0737,0 -21.9733,-2.16911 -27.119,-6.64886 -5.0851,-4.47975 -8.5652,-10.80255 -8.5652,-24.18125 v -21.70256 c 0,-9.1411 -1.6345,-15.46723 -4.9035,-18.97838 -3.269,-3.51115 -10.8505,-6.9352 -19.4465,-6.90005 l -5.9193,0.0242 -0.03,-8.9218 5.9848,-0.004 c 8.6568,-0.007 16.2026,-4.17511 19.4111,-7.62572 3.269,-3.51115 4.9035,-9.77674 4.9035,-18.79677 v -21.79336 c 0,-13.37871 5.4191,-20.39008 10.5042,-24.80929 5.1457,-4.47975 10.0585,-6.10685 25.1322,-6.10685 h 5.6299 l -0.025,8.02977 h -6.1748 c -8.5357,0 -14.7584,2.47927 -17.3615,5.1429 -2.6031,2.66363 -5.4388,9.4671 -5.4388,18.00283 v 22.5198 c 0,9.50433 -1.3924,16.40556 -4.1771,20.70369 -2.7242,4.29814 -5.1553,7.30778 -11.8144,8.82121 6.7197,1.6345 9.1811,4.49696 11.9052,8.79509 2.7242,4.29814 4.0863,11.1691 4.0863,20.61289 v 22.5198 c 0,8.53573 2.0913,15.64155 4.6944,18.30518 2.6031,2.66363 9.6432,5.77285 18.1789,5.77218 z"
         />
    </g>`, sign)

	case 2:
		tmpl = tmplBrace
		params.FontSize = 97.9
		params.ScaleX = 0.76
		params.ScaleY = 1.31

	case 3:
		tmpl = tmplBrace
		params.FontSize = 63.4
		params.ScaleX = 0.91
		params.ScaleY = 1.10
	case 4:
		tmpl = tmplBrace
		params.FontSize = 38.5
		params.ScaleX = 1.01
		params.ScaleY = 0.98

	default:
		panic(fmt.Sprintf("unsupported ascender generation: %d", generation))

	}

	t, err := template.New("ascender").Parse(tmpl)
	if err != nil {
		panic(err)
	}
	err = t.Execute(buf, params)
	if err != nil {
		panic(err)
	}
}

func (r *ButterflyChart) renderMarriage(opts *ButterflyLayoutOptions, buf *bytes.Buffer, maleId string) {
	type paramtype struct {
		Id        string
		TextLine1 string
		TextLine2 string

		X          float64
		UpperY     float64
		LowerY     float64
		LineLength float64

		FontSize   float64
		LineHeight float64
		TextLine1Y float64
		TextLine2Y float64
	}
	var params paramtype

	params.TextLine1 = "m. 26 Aug 1905"
	params.TextLine2 = "Ireland"

	var tmpl string
	tmplSimple := `<path
       style="fill:#4d4d4d;stroke:#000000;stroke-width:1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:1,3;stroke-dashoffset:0;stroke-opacity:1"
       d="m {{- .X -}},{{- .UpperY -}} v {{- .LineLength -}}"
       id="marriage-{{- .Id -}}-line1" />
    <path
       style="fill:#4d4d4d;stroke:#000000;stroke-width:1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:1,3;stroke-dashoffset:0;stroke-opacity:1"
       d="m {{- .X -}},{{- .LowerY -}} v -{{- .LineLength -}}"
       id="marriage-{{- .Id -}}-line2"/>
	<text
       xml:space="preserve"
       style="font-family:Georgia;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:none"
       id="person-{{- .Id -}}-text"><tspan
         style="font-size:{{- .FontSize -}}px;line-height:{{- .LineHeight -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .X -}}" y="{{- .TextLine1Y -}}">{{- .TextLine1 -}}</tspan><tspan
         style="font-size:{{- .FontSize -}}px;line-height:{{- .LineHeight -}};text-align:center;text-anchor:middle;opacity:1;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.1;stroke-dasharray:none"
         x="{{- .X -}}" y="{{- .TextLine2Y -}}">{{- .TextLine2 -}}</tspan></text>`

	// ranges from 0 to 3
	generation := len(maleId) - 1

	left, top, width, height := opts.boxBounds(maleId)

	params.X = left + width/2
	params.UpperY = top + height

	femaleId := maleId[:len(maleId)-1] + "m"
	_, ftop, _, _ := opts.boxBounds(femaleId)
	params.LowerY = ftop

	switch generation {
	case 0:
	case 1:
		tmpl = tmplSimple
		params.LineLength = 155
		if maleId[0] == 'p' {
			params.X -= 20
		} else {
			params.X += 20
		}
		params.LineHeight = 1.25
		params.FontSize = 7.76111
		params.TextLine1Y = params.UpperY + (params.LowerY-params.UpperY)/2 - params.LineHeight*params.FontSize/2
		params.TextLine2Y = params.TextLine1Y + params.LineHeight*params.FontSize

	case 2:
		tmpl = tmplSimple
		params.LineLength = 65

		params.LineHeight = 1.25
		params.FontSize = 7.76111
		params.TextLine1Y = params.UpperY + (params.LowerY-params.UpperY)/2 - params.LineHeight*params.FontSize/2
		params.TextLine2Y = params.TextLine1Y + params.LineHeight*params.FontSize

	case 3:
		tmpl = tmplSimple
		params.LineLength = 24

		params.LineHeight = 1.25
		params.FontSize = 7.76111
		params.TextLine1Y = params.UpperY + (params.LowerY-params.UpperY)/2 - params.LineHeight*params.FontSize/2
		params.TextLine2Y = params.TextLine1Y + params.LineHeight*params.FontSize

	default:
		panic(fmt.Sprintf("unsupported marriage generation: %d", generation))

	}

	t, err := template.New("marriage").Parse(tmpl)
	if err != nil {
		panic(err)
	}
	err = t.Execute(buf, params)
	if err != nil {
		panic(err)
	}
}

func (ch *ButterflyChart) renderBackground(opts *ButterflyLayoutOptions, buf *bytes.Buffer) {
	// White background
	fmt.Fprintln(buf, `<rect id="background" width="100%" height="100%" fill="white"/>`)

	// Floral design
	buf.WriteString(`<g id="florals" style="display:inline;fill:#b3b3b3">
    <g
       transform="matrix(-0.04371296,0,0,0.04371296,1173.7377,9.0243058)"
       fill="#000000"
       stroke="none"
       id="floral-tr"
       style="display:inline;fill:#e6e6e6;stroke-width:0.807032">
      <path
         d="m 6803,11943 c 20,-2 54,-2 75,0 20,2 3,4 -38,4 -41,0 -58,-2 -37,-4 z"
         id="path11459"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 7012,11926 c 822,-120 1438,-732 1569,-1561 17,-103 17,-407 0,-510 -41,-260 -142,-536 -282,-770 -73,-121 -219,-319 -307,-415 l -60,-65 -23,20 -24,20 20,-24 20,-23 -100,-96 c -253,-241 -602,-475 -1021,-685 -311,-155 -982,-436 -989,-414 -2,7 -11,55 -21,107 -52,286 -170,622 -315,900 -400,762 -1080,1327 -1825,1516 -413,104 -794,95 -1168,-27 -523,-171 -927,-568 -1096,-1078 -71,-213 -84,-297 -84,-546 0,-192 2,-233 22,-324 88,-403 292,-746 582,-981 30,-24 8,1 -49,55 -128,122 -218,233 -299,368 -390,646 -297,1471 226,2018 254,266 572,432 967,504 142,26 414,31 573,11 1158,-148 2152,-1184 2371,-2472 l 18,-103 -122,-49 -123,-49 -42,116 c -109,300 -297,631 -511,901 -103,131 -362,391 -490,493 -354,283 -731,456 -1114,513 -145,22 -408,14 -543,-15 -274,-59 -500,-175 -678,-347 -242,-237 -377,-539 -389,-876 l -4,-123 9,100 c 26,281 82,454 216,658 61,93 238,270 334,334 296,198 661,269 1055,207 423,-66 856,-307 1240,-689 329,-327 571,-701 733,-1135 77,-204 83,-177 -55,-235 -411,-170 -812,-375 -1068,-545 -38,-25 -75,-50 -82,-54 -7,-3 8,35 32,86 67,144 174,488 208,673 2,11 7,-17 11,-62 15,-157 -14,-334 -73,-452 -19,-37 -27,-61 -19,-61 6,0 49,18 94,40 103,50 183,125 222,207 25,54 27,67 26,183 0,109 -5,146 -39,290 -34,145 -39,181 -39,296 l -1,132 -25,-18 c -37,-27 -241,-236 -300,-309 -151,-184 -304,-459 -354,-636 -51,-179 -53,-347 -6,-485 17,-52 19,-65 8,-75 -7,-6 -58,-50 -114,-97 -284,-245 -525,-561 -695,-913 -94,-195 -145,-332 -205,-553 l -17,-63 -28,49 c -36,64 -162,197 -263,278 -45,36 -169,124 -276,197 -107,72 -219,148 -248,170 -30,21 -56,37 -59,34 -6,-6 41,-413 63,-542 39,-236 103,-436 185,-582 63,-112 182,-229 279,-276 73,-36 189,-72 198,-62 12,11 13,156 3,223 -31,201 -172,511 -325,720 -63,85 -25,64 54,-31 114,-135 204,-287 295,-501 l 47,-110 2,-245 c 2,-193 8,-277 25,-391 68,-449 228,-857 494,-1255 243,-365 660,-727 1084,-939 804,-402 1772,-416 2570,-36 320,152 583,339 839,594 251,252 458,544 590,833 30,66 31,74 16,82 -26,13 -259,111 -266,111 -3,0 -6,-5 -6,-11 0,-18 -128,-266 -187,-361 -265,-429 -671,-791 -1138,-1014 -255,-122 -505,-195 -805,-235 -126,-17 -507,-18 -640,-1 -506,63 -955,246 -1350,550 -204,157 -473,457 -625,697 -139,220 -301,605 -340,809 -18,91 -20,366 -5,501 58,502 284,977 632,1325 191,191 366,318 599,434 254,127 519,207 869,265 211,35 379,56 386,48 9,-8 -25,-242 -56,-386 -103,-488 -321,-934 -638,-1311 -98,-117 -321,-336 -422,-414 -103,-81 -104,-82 -78,-119 12,-17 43,-52 70,-79 l 47,-47 69,58 c 209,174 373,347 538,566 128,171 218,323 391,665 195,383 284,616 368,963 18,77 35,140 37,142 2,1 73,10 158,20 486,53 805,115 1093,215 636,219 1068,646 1178,1165 57,272 23,558 -96,809 -34,72 -130,226 -140,226 -7,0 -6,-2 53,-93 264,-411 253,-955 -28,-1372 -314,-465 -883,-749 -1708,-851 -152,-18 -496,-54 -498,-51 -1,1 6,58 16,127 39,272 49,539 29,808 -6,81 -11,151 -12,156 -1,4 106,54 238,110 355,151 497,216 730,335 675,345 1140,723 1452,1183 l 66,96 66,16 c 36,9 161,31 277,50 230,37 337,68 418,120 129,83 210,236 249,476 28,169 49,560 30,560 -10,0 -199,-139 -313,-229 -222,-175 -377,-358 -479,-565 -43,-86 -84,-196 -77,-204 9,-8 140,59 226,116 101,67 249,212 308,301 53,81 54,69 0,-28 -91,-167 -215,-302 -373,-406 -78,-51 -294,-175 -306,-175 -3,0 29,69 71,153 117,235 183,449 215,690 46,353 -24,740 -194,1075 -249,489 -710,844 -1253,966 -64,14 -151,28 -194,31 l -79,5 z M 5746,6995 c 9,-242 -9,-481 -57,-762 -11,-65 -14,-73 -34,-73 -22,0 -22,3 -28,198 -8,250 -38,459 -98,695 -11,44 -18,81 -16,84 3,2 51,24 108,48 l 104,44 7,-27 c 4,-15 11,-108 14,-207 z m -338,-12 c 50,-223 75,-480 70,-703 l -3,-135 -50,-7 c -27,-4 -124,-17 -215,-28 -520,-64 -920,-191 -1249,-395 -370,-229 -675,-563 -871,-954 -43,-86 -113,-267 -144,-371 -22,-75 -25,-80 -25,-45 -1,22 8,99 19,170 105,684 439,1278 965,1719 201,169 402,306 664,457 199,113 767,387 807,389 6,0 21,-44 32,-97 z m 249,-910 c -3,-10 -11,-43 -17,-73 -18,-79 -30,-95 -24,-30 9,100 14,120 30,120 10,0 14,-6 11,-17 z"
         id="path11461"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 3799,10642 c -170,-89 -151,-343 30,-407 84,-29 169,-10 233,52 143,139 43,383 -157,383 -38,0 -67,-8 -106,-28 z"
         id="path11463"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 4498,10311 c -46,-15 -83,-42 -108,-79 -20,-29 -25,-48 -25,-98 0,-55 4,-67 31,-101 39,-48 87,-73 144,-73 57,0 105,25 144,73 27,34 31,46 31,101 0,50 -5,69 -25,99 -42,61 -130,97 -192,78 z"
         id="path11465"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 4962,9790 c -87,-53 -95,-157 -17,-230 22,-21 37,-25 85,-25 63,0 96,20 126,77 44,84 -29,198 -126,198 -19,0 -50,-9 -68,-20 z"
         id="path11467"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 5360,9343 c -8,-3 -26,-15 -38,-27 -54,-48 -67,-117 -32,-173 72,-119 244,-81 258,56 5,61 -21,111 -72,135 -33,16 -86,20 -116,9 z"
         id="path11469"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 7123,8804 c -49,-49 -10,-134 61,-134 55,0 93,57 75,110 -20,56 -91,69 -136,24 z"
         id="path11471"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 6903,8444 c -18,-9 -39,-31 -48,-51 -33,-69 18,-153 92,-153 70,0 113,43 113,113 0,39 -21,72 -60,92 -36,19 -57,19 -97,-1 z"
         id="path11473"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 11072,8227 c -65,-69 34,-163 102,-97 24,24 25,74 1,100 -25,28 -76,26 -103,-3 z"
         id="path11475"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 10578,8223 c 7,-3 16,-2 19,1 4,3 -2,6 -13,5 -11,0 -14,-3 -6,-6 z"
         id="path11477"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 10625,8210 c 11,-5 27,-9 35,-9 9,0 8,4 -5,9 -11,5 -27,9 -35,9 -9,0 -8,-4 5,-9 z"
         id="path11479"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 10730,8178 c 14,-6 63,-29 109,-50 287,-131 519,-381 636,-683 144,-373 118,-751 -77,-1107 -242,-443 -696,-762 -1308,-918 -356,-91 -718,-126 -1485,-145 -725,-18 -1129,-64 -1500,-172 -302,-87 -596,-233 -806,-398 -93,-73 -249,-230 -314,-315 -367,-479 -414,-1083 -121,-1535 76,-118 242,-281 358,-353 479,-296 1090,-220 1479,185 89,93 173,216 224,328 31,68 33,75 17,83 -26,14 -123,45 -127,41 -1,-2 -16,-34 -33,-71 -165,-366 -542,-608 -947,-608 -517,0 -956,381 -1035,899 -19,124 -8,330 24,456 66,256 180,456 375,659 371,383 891,599 1641,680 213,23 390,34 825,51 726,28 1080,67 1413,156 473,125 816,316 1098,612 377,394 507,915 349,1394 -89,268 -267,510 -487,660 -76,53 -260,146 -308,157 l -25,6 z"
         id="path11481"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 6539,8135 c -28,-15 -59,-70 -59,-105 0,-54 65,-120 118,-120 71,0 122,54 122,131 0,33 -6,47 -34,75 -28,28 -41,34 -77,34 -24,-1 -56,-7 -70,-15 z"
         id="path11483"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 11374,8081 c -87,-53 -50,-183 52,-183 48,0 81,25 95,70 14,48 -6,94 -51,117 -38,20 -58,19 -96,-4 z"
         id="path11485"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 6111,7884 c -43,-36 -58,-89 -38,-137 18,-42 68,-77 112,-77 44,0 94,35 112,77 32,77 -27,163 -112,163 -32,0 -51,-7 -74,-26 z"
         id="path11487"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="M 522,7423 C 242,6837 88,6276 18,5590 -1,5402 2,4863 24,4670 46,4468 66,4334 96,4180 361,2845 1136,1668 2250,908 2633,646 3107,411 3543,266 4523,-60 5568,-85 6581,194 c 219,60 494,158 704,251 139,61 315,149 315,157 0,9 -41,91 -58,116 -7,11 -47,-4 -183,-71 C 6791,368 6188,202 5543,145 c -222,-20 -660,-19 -868,0 -646,62 -1225,224 -1788,502 -384,189 -703,396 -1037,672 -140,115 -495,469 -609,606 C 333,3018 -91,4405 65,5775 c 68,596 242,1198 487,1689 22,43 37,80 35,82 -2,2 -31,-53 -65,-123 z"
         id="path11489"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 8263,7134 c -54,-27 -77,-77 -74,-158 1,-39 66,-102 113,-111 50,-9 121,23 148,67 73,119 -61,264 -187,202 z"
         id="path11491"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 7384,6986 c -43,-19 -64,-53 -64,-104 0,-121 172,-144 210,-28 28,85 -65,168 -146,132 z"
         id="path11493"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 1970,6926 c 0,-2 8,-10 18,-17 15,-13 16,-12 3,4 -13,16 -21,21 -21,13 z"
         id="path11495"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 12050,6581 c -19,-10 -80,-33 -135,-51 -328,-105 -370,-124 -439,-200 -35,-39 -71,-98 -64,-106 2,-2 26,13 54,32 74,54 182,102 244,108 l 55,6 -51,-18 c -91,-32 -196,-102 -286,-192 -49,-47 -88,-92 -88,-99 0,-12 22,-13 124,-8 205,11 350,51 431,120 45,37 83,115 117,235 16,59 40,126 53,150 13,23 23,42 22,42 -1,-1 -18,-9 -37,-19 z"
         id="path11497"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 9865,6485 c -17,-41 -79,-132 -128,-189 -28,-32 -141,-136 -252,-231 -283,-243 -369,-342 -410,-470 -20,-66 -17,-197 7,-253 l 14,-32 21,22 c 29,32 80,115 143,233 29,55 70,125 91,156 40,60 134,151 180,174 25,14 23,7 -27,-62 -75,-105 -131,-201 -175,-303 -36,-84 -65,-192 -55,-208 9,-16 216,8 293,33 289,97 384,322 338,805 -10,102 -21,221 -24,265 -5,58 -10,74 -16,60 z"
         id="path11499"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 462,5745 c 0,-16 2,-22 5,-12 2,9 2,23 0,30 -3,6 -5,-1 -5,-18 z"
         id="path11501"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 453,5650 c 0,-30 2,-43 4,-27 2,15 2,39 0,55 -2,15 -4,2 -4,-28 z"
         id="path11503"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="M 436,5388 C 417,4872 504,4276 671,3776 1002,2783 1674,1911 2540,1348 3903,462 5598,339 7075,1018 c 444,204 837,462 1605,1052 684,525 966,720 1351,933 487,269 958,417 1499,472 202,21 647,16 830,-9 113,-16 368,-62 393,-71 13,-6 51,135 39,146 -10,10 -277,57 -432,76 -195,24 -664,24 -860,0 -700,-86 -1253,-300 -1929,-747 C 9342,2719 9130,2564 8781,2291 7845,1561 7598,1390 7145,1160 6642,904 6136,754 5544,683 5336,659 4777,658 4574,683 3909,763 3303,961 2750,1280 1529,1985 680,3245 491,4635 c -34,249 -42,377 -42,668 0,152 -2,277 -3,277 -2,0 -6,-87 -10,-192 z"
         id="path11505"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 4209,5516 c -113,-40 -176,-153 -149,-265 50,-208 343,-233 426,-36 21,51 21,133 -2,180 -26,56 -81,103 -137,120 -61,18 -90,18 -138,1 z"
         id="path11507"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 3702,4979 c -96,-48 -128,-169 -69,-258 49,-74 140,-100 223,-62 136,62 136,260 0,322 -54,24 -101,24 -154,-2 z"
         id="path11509"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 902,4769 c -131,-84 -67,-289 90,-289 50,0 97,26 132,73 29,39 29,125 0,164 -54,73 -152,96 -222,52 z"
         id="path11511"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 11335,4176 c -102,-44 -122,-172 -40,-245 134,-117 322,64 208,200 -46,54 -109,71 -168,45 z"
         id="path11513"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 10736,4078 c -53,-31 -78,-80 -73,-139 5,-57 30,-93 85,-120 97,-48 206,21 206,131 0,113 -122,185 -218,128 z"
         id="path11515"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 1230,3994 c -138,-60 -178,-235 -80,-347 51,-57 97,-79 170,-79 127,0 220,95 220,224 -1,91 -51,169 -132,202 -47,20 -132,20 -178,0 z"
         id="path11517"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 10200,3858 c -42,-29 -60,-61 -60,-111 0,-57 25,-96 77,-119 77,-34 146,-6 177,73 13,31 14,49 6,78 -23,88 -127,129 -200,79 z"
         id="path11519"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 9739,3531 c -55,-55 -30,-148 45,-167 74,-18 140,48 122,122 -19,75 -112,100 -167,45 z"
         id="path11521"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 1815,3197 c -92,-37 -151,-95 -184,-181 -27,-72 -27,-130 0,-202 75,-199 330,-258 478,-112 70,70 86,108 86,213 0,105 -18,148 -87,215 -73,70 -207,100 -293,67 z"
         id="path11523"
         style="fill:#e6e6e6;stroke-width:0.807032" />
      <path
         d="m 9372,3175 c -39,-33 -48,-74 -27,-118 30,-63 104,-76 157,-28 28,25 33,36 32,72 -3,86 -97,129 -162,74 z"
         id="path11525"
         style="fill:#e6e6e6;stroke-width:0.807032" />
    </g>
    <g
       transform="matrix(0.04466596,0,0,-0.04466596,14.76115,832.37505)"
       fill="#000000"
       stroke="none"
       id="floral-bl"
       style="display:inline;fill:#e6e6e6;stroke-width:0.789814">
      <path
         d="m 6803,11943 c 20,-2 54,-2 75,0 20,2 3,4 -38,4 -41,0 -58,-2 -37,-4 z"
         id="path14444"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 7012,11926 c 822,-120 1438,-732 1569,-1561 17,-103 17,-407 0,-510 -41,-260 -142,-536 -282,-770 -73,-121 -219,-319 -307,-415 l -60,-65 -23,20 -24,20 20,-24 20,-23 -100,-96 c -253,-241 -602,-475 -1021,-685 -311,-155 -982,-436 -989,-414 -2,7 -11,55 -21,107 -52,286 -170,622 -315,900 -400,762 -1080,1327 -1825,1516 -413,104 -794,95 -1168,-27 -523,-171 -927,-568 -1096,-1078 -71,-213 -84,-297 -84,-546 0,-192 2,-233 22,-324 88,-403 292,-746 582,-981 30,-24 8,1 -49,55 -128,122 -218,233 -299,368 -390,646 -297,1471 226,2018 254,266 572,432 967,504 142,26 414,31 573,11 1158,-148 2152,-1184 2371,-2472 l 18,-103 -122,-49 -123,-49 -42,116 c -109,300 -297,631 -511,901 -103,131 -362,391 -490,493 -354,283 -731,456 -1114,513 -145,22 -408,14 -543,-15 -274,-59 -500,-175 -678,-347 -242,-237 -377,-539 -389,-876 l -4,-123 9,100 c 26,281 82,454 216,658 61,93 238,270 334,334 296,198 661,269 1055,207 423,-66 856,-307 1240,-689 329,-327 571,-701 733,-1135 77,-204 83,-177 -55,-235 -411,-170 -812,-375 -1068,-545 -38,-25 -75,-50 -82,-54 -7,-3 8,35 32,86 67,144 174,488 208,673 2,11 7,-17 11,-62 15,-157 -14,-334 -73,-452 -19,-37 -27,-61 -19,-61 6,0 49,18 94,40 103,50 183,125 222,207 25,54 27,67 26,183 0,109 -5,146 -39,290 -34,145 -39,181 -39,296 l -1,132 -25,-18 c -37,-27 -241,-236 -300,-309 -151,-184 -304,-459 -354,-636 -51,-179 -53,-347 -6,-485 17,-52 19,-65 8,-75 -7,-6 -58,-50 -114,-97 -284,-245 -525,-561 -695,-913 -94,-195 -145,-332 -205,-553 l -17,-63 -28,49 c -36,64 -162,197 -263,278 -45,36 -169,124 -276,197 -107,72 -219,148 -248,170 -30,21 -56,37 -59,34 -6,-6 41,-413 63,-542 39,-236 103,-436 185,-582 63,-112 182,-229 279,-276 73,-36 189,-72 198,-62 12,11 13,156 3,223 -31,201 -172,511 -325,720 -63,85 -25,64 54,-31 114,-135 204,-287 295,-501 l 47,-110 2,-245 c 2,-193 8,-277 25,-391 68,-449 228,-857 494,-1255 243,-365 660,-727 1084,-939 804,-402 1772,-416 2570,-36 320,152 583,339 839,594 251,252 458,544 590,833 30,66 31,74 16,82 -26,13 -259,111 -266,111 -3,0 -6,-5 -6,-11 0,-18 -128,-266 -187,-361 -265,-429 -671,-791 -1138,-1014 -255,-122 -505,-195 -805,-235 -126,-17 -507,-18 -640,-1 -506,63 -955,246 -1350,550 -204,157 -473,457 -625,697 -139,220 -301,605 -340,809 -18,91 -20,366 -5,501 58,502 284,977 632,1325 191,191 366,318 599,434 254,127 519,207 869,265 211,35 379,56 386,48 9,-8 -25,-242 -56,-386 -103,-488 -321,-934 -638,-1311 -98,-117 -321,-336 -422,-414 -103,-81 -104,-82 -78,-119 12,-17 43,-52 70,-79 l 47,-47 69,58 c 209,174 373,347 538,566 128,171 218,323 391,665 195,383 284,616 368,963 18,77 35,140 37,142 2,1 73,10 158,20 486,53 805,115 1093,215 636,219 1068,646 1178,1165 57,272 23,558 -96,809 -34,72 -130,226 -140,226 -7,0 -6,-2 53,-93 264,-411 253,-955 -28,-1372 -314,-465 -883,-749 -1708,-851 -152,-18 -496,-54 -498,-51 -1,1 6,58 16,127 39,272 49,539 29,808 -6,81 -11,151 -12,156 -1,4 106,54 238,110 355,151 497,216 730,335 675,345 1140,723 1452,1183 l 66,96 66,16 c 36,9 161,31 277,50 230,37 337,68 418,120 129,83 210,236 249,476 28,169 49,560 30,560 -10,0 -199,-139 -313,-229 -222,-175 -377,-358 -479,-565 -43,-86 -84,-196 -77,-204 9,-8 140,59 226,116 101,67 249,212 308,301 53,81 54,69 0,-28 -91,-167 -215,-302 -373,-406 -78,-51 -294,-175 -306,-175 -3,0 29,69 71,153 117,235 183,449 215,690 46,353 -24,740 -194,1075 -249,489 -710,844 -1253,966 -64,14 -151,28 -194,31 l -79,5 z M 5746,6995 c 9,-242 -9,-481 -57,-762 -11,-65 -14,-73 -34,-73 -22,0 -22,3 -28,198 -8,250 -38,459 -98,695 -11,44 -18,81 -16,84 3,2 51,24 108,48 l 104,44 7,-27 c 4,-15 11,-108 14,-207 z m -338,-12 c 50,-223 75,-480 70,-703 l -3,-135 -50,-7 c -27,-4 -124,-17 -215,-28 -520,-64 -920,-191 -1249,-395 -370,-229 -675,-563 -871,-954 -43,-86 -113,-267 -144,-371 -22,-75 -25,-80 -25,-45 -1,22 8,99 19,170 105,684 439,1278 965,1719 201,169 402,306 664,457 199,113 767,387 807,389 6,0 21,-44 32,-97 z m 249,-910 c -3,-10 -11,-43 -17,-73 -18,-79 -30,-95 -24,-30 9,100 14,120 30,120 10,0 14,-6 11,-17 z"
         id="path14446"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 3799,10642 c -170,-89 -151,-343 30,-407 84,-29 169,-10 233,52 143,139 43,383 -157,383 -38,0 -67,-8 -106,-28 z"
         id="path14448"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 4498,10311 c -46,-15 -83,-42 -108,-79 -20,-29 -25,-48 -25,-98 0,-55 4,-67 31,-101 39,-48 87,-73 144,-73 57,0 105,25 144,73 27,34 31,46 31,101 0,50 -5,69 -25,99 -42,61 -130,97 -192,78 z"
         id="path14450"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 4962,9790 c -87,-53 -95,-157 -17,-230 22,-21 37,-25 85,-25 63,0 96,20 126,77 44,84 -29,198 -126,198 -19,0 -50,-9 -68,-20 z"
         id="path14452"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 5360,9343 c -8,-3 -26,-15 -38,-27 -54,-48 -67,-117 -32,-173 72,-119 244,-81 258,56 5,61 -21,111 -72,135 -33,16 -86,20 -116,9 z"
         id="path14454"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 7123,8804 c -49,-49 -10,-134 61,-134 55,0 93,57 75,110 -20,56 -91,69 -136,24 z"
         id="path14456"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 6903,8444 c -18,-9 -39,-31 -48,-51 -33,-69 18,-153 92,-153 70,0 113,43 113,113 0,39 -21,72 -60,92 -36,19 -57,19 -97,-1 z"
         id="path14458"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 11072,8227 c -65,-69 34,-163 102,-97 24,24 25,74 1,100 -25,28 -76,26 -103,-3 z"
         id="path14460"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 10578,8223 c 7,-3 16,-2 19,1 4,3 -2,6 -13,5 -11,0 -14,-3 -6,-6 z"
         id="path14462"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 10625,8210 c 11,-5 27,-9 35,-9 9,0 8,4 -5,9 -11,5 -27,9 -35,9 -9,0 -8,-4 5,-9 z"
         id="path14464"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 10730,8178 c 14,-6 63,-29 109,-50 287,-131 519,-381 636,-683 144,-373 118,-751 -77,-1107 -242,-443 -696,-762 -1308,-918 -356,-91 -718,-126 -1485,-145 -725,-18 -1129,-64 -1500,-172 -302,-87 -596,-233 -806,-398 -93,-73 -249,-230 -314,-315 -367,-479 -414,-1083 -121,-1535 76,-118 242,-281 358,-353 479,-296 1090,-220 1479,185 89,93 173,216 224,328 31,68 33,75 17,83 -26,14 -123,45 -127,41 -1,-2 -16,-34 -33,-71 -165,-366 -542,-608 -947,-608 -517,0 -956,381 -1035,899 -19,124 -8,330 24,456 66,256 180,456 375,659 371,383 891,599 1641,680 213,23 390,34 825,51 726,28 1080,67 1413,156 473,125 816,316 1098,612 377,394 507,915 349,1394 -89,268 -267,510 -487,660 -76,53 -260,146 -308,157 l -25,6 z"
         id="path14466"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 6539,8135 c -28,-15 -59,-70 -59,-105 0,-54 65,-120 118,-120 71,0 122,54 122,131 0,33 -6,47 -34,75 -28,28 -41,34 -77,34 -24,-1 -56,-7 -70,-15 z"
         id="path14468"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 11374,8081 c -87,-53 -50,-183 52,-183 48,0 81,25 95,70 14,48 -6,94 -51,117 -38,20 -58,19 -96,-4 z"
         id="path14470"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 6111,7884 c -43,-36 -58,-89 -38,-137 18,-42 68,-77 112,-77 44,0 94,35 112,77 32,77 -27,163 -112,163 -32,0 -51,-7 -74,-26 z"
         id="path14472"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="M 522,7423 C 242,6837 88,6276 18,5590 -1,5402 2,4863 24,4670 46,4468 66,4334 96,4180 361,2845 1136,1668 2250,908 2633,646 3107,411 3543,266 4523,-60 5568,-85 6581,194 c 219,60 494,158 704,251 139,61 315,149 315,157 0,9 -41,91 -58,116 -7,11 -47,-4 -183,-71 C 6791,368 6188,202 5543,145 c -222,-20 -660,-19 -868,0 -646,62 -1225,224 -1788,502 -384,189 -703,396 -1037,672 -140,115 -495,469 -609,606 C 333,3018 -91,4405 65,5775 c 68,596 242,1198 487,1689 22,43 37,80 35,82 -2,2 -31,-53 -65,-123 z"
         id="path14474"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 8263,7134 c -54,-27 -77,-77 -74,-158 1,-39 66,-102 113,-111 50,-9 121,23 148,67 73,119 -61,264 -187,202 z"
         id="path14476"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 7384,6986 c -43,-19 -64,-53 -64,-104 0,-121 172,-144 210,-28 28,85 -65,168 -146,132 z"
         id="path14478"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 1970,6926 c 0,-2 8,-10 18,-17 15,-13 16,-12 3,4 -13,16 -21,21 -21,13 z"
         id="path14480"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 12050,6581 c -19,-10 -80,-33 -135,-51 -328,-105 -370,-124 -439,-200 -35,-39 -71,-98 -64,-106 2,-2 26,13 54,32 74,54 182,102 244,108 l 55,6 -51,-18 c -91,-32 -196,-102 -286,-192 -49,-47 -88,-92 -88,-99 0,-12 22,-13 124,-8 205,11 350,51 431,120 45,37 83,115 117,235 16,59 40,126 53,150 13,23 23,42 22,42 -1,-1 -18,-9 -37,-19 z"
         id="path14482"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 9865,6485 c -17,-41 -79,-132 -128,-189 -28,-32 -141,-136 -252,-231 -283,-243 -369,-342 -410,-470 -20,-66 -17,-197 7,-253 l 14,-32 21,22 c 29,32 80,115 143,233 29,55 70,125 91,156 40,60 134,151 180,174 25,14 23,7 -27,-62 -75,-105 -131,-201 -175,-303 -36,-84 -65,-192 -55,-208 9,-16 216,8 293,33 289,97 384,322 338,805 -10,102 -21,221 -24,265 -5,58 -10,74 -16,60 z"
         id="path14484"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 462,5745 c 0,-16 2,-22 5,-12 2,9 2,23 0,30 -3,6 -5,-1 -5,-18 z"
         id="path14486"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 453,5650 c 0,-30 2,-43 4,-27 2,15 2,39 0,55 -2,15 -4,2 -4,-28 z"
         id="path14488"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="M 436,5388 C 417,4872 504,4276 671,3776 1002,2783 1674,1911 2540,1348 3903,462 5598,339 7075,1018 c 444,204 837,462 1605,1052 684,525 966,720 1351,933 487,269 958,417 1499,472 202,21 647,16 830,-9 113,-16 368,-62 393,-71 13,-6 51,135 39,146 -10,10 -277,57 -432,76 -195,24 -664,24 -860,0 -700,-86 -1253,-300 -1929,-747 C 9342,2719 9130,2564 8781,2291 7845,1561 7598,1390 7145,1160 6642,904 6136,754 5544,683 5336,659 4777,658 4574,683 3909,763 3303,961 2750,1280 1529,1985 680,3245 491,4635 c -34,249 -42,377 -42,668 0,152 -2,277 -3,277 -2,0 -6,-87 -10,-192 z"
         id="path14490"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 4209,5516 c -113,-40 -176,-153 -149,-265 50,-208 343,-233 426,-36 21,51 21,133 -2,180 -26,56 -81,103 -137,120 -61,18 -90,18 -138,1 z"
         id="path14492"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 3702,4979 c -96,-48 -128,-169 -69,-258 49,-74 140,-100 223,-62 136,62 136,260 0,322 -54,24 -101,24 -154,-2 z"
         id="path14494"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 902,4769 c -131,-84 -67,-289 90,-289 50,0 97,26 132,73 29,39 29,125 0,164 -54,73 -152,96 -222,52 z"
         id="path14496"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 11335,4176 c -102,-44 -122,-172 -40,-245 134,-117 322,64 208,200 -46,54 -109,71 -168,45 z"
         id="path14498"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 10736,4078 c -53,-31 -78,-80 -73,-139 5,-57 30,-93 85,-120 97,-48 206,21 206,131 0,113 -122,185 -218,128 z"
         id="path14500"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 1230,3994 c -138,-60 -178,-235 -80,-347 51,-57 97,-79 170,-79 127,0 220,95 220,224 -1,91 -51,169 -132,202 -47,20 -132,20 -178,0 z"
         id="path14502"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 10200,3858 c -42,-29 -60,-61 -60,-111 0,-57 25,-96 77,-119 77,-34 146,-6 177,73 13,31 14,49 6,78 -23,88 -127,129 -200,79 z"
         id="path14504"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 9739,3531 c -55,-55 -30,-148 45,-167 74,-18 140,48 122,122 -19,75 -112,100 -167,45 z"
         id="path14506"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 1815,3197 c -92,-37 -151,-95 -184,-181 -27,-72 -27,-130 0,-202 75,-199 330,-258 478,-112 70,70 86,108 86,213 0,105 -18,148 -87,215 -73,70 -207,100 -293,67 z"
         id="path14508"
         style="fill:#e6e6e6;stroke-width:0.789814" />
      <path
         d="m 9372,3175 c -39,-33 -48,-74 -27,-118 30,-63 104,-76 157,-28 28,25 33,36 32,72 -3,86 -97,129 -162,74 z"
         id="path14510"
         style="fill:#e6e6e6;stroke-width:0.789814" />
    </g>
  </g>`)

	// Note
	if ch.Note != "" {
		buf.WriteString(fmt.Sprintf(`<text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.2;font-family:Georgia;-inkscape-font-specification:Georgia;text-align:center;text-anchor:middle;fill:none;stroke:#000000;stroke-width:0.1;stroke-linecap:round;stroke-miterlimit:11.1;stroke-dasharray:0.1, 0.1"
       x="972.01276"
       y="834.27167"
       id="text18595"><tspan
         id="tspan18593"
         style="fill:#000000;stroke-width:0.1"
         x="972.01276"
         y="834.27167">%s</tspan></text>`, ch.Note))
	}
}

func (ch *ButterflyChart) renderTitle(opts *ButterflyLayoutOptions, buf *bytes.Buffer) {
	if ch.TitleLine1 == "" && ch.TitleLine2 == "" {
		return
	}
	buf.WriteString(fmt.Sprintf(`<text
       xml:space="preserve"
       style="font-style:normal;font-size:22.57777778px;font-family:Georgia;-inkscape-font-specification:Georgia;text-align:center;text-anchor:middle;fill:#e6e6e6;stroke:#000000;stroke-width:0.4;stroke-linecap:round;stroke-miterlimit:11.1;font-weight:normal;font-stretch:normal;font-variant:normal"
       x="597.32202"
       y="26.130615"
       id="text17898"><tspan
         style="font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;font-family:Georgia;fill:#000000;stroke-width:0.4;font-size:22.57777778px"
         x="597.32202"
         y="25"
         id="tspan18173">%s</tspan><tspan
         style="font-style:normal;font-variant:normal;font-weight:normal;font-stretch:normal;font-family:Georgia;fill:#000000;stroke-width:0.4;font-size:22.57777778px"
         x="597.32202"
         y="50"
         id="tspan20937">%s</tspan></text>`, ch.TitleLine1, ch.TitleLine2))

	buf.WriteString(`<path
       style="fill:none;stroke:#333333;stroke-width:0.499999;stroke-linecap:round;stroke-miterlimit:11.1"
       d="m 365.43622,65 h 467.3201"
       id="path21809" />`)
}

func (ch *ButterflyChart) generateDebug(opts *ButterflyLayoutOptions) {
	dummyPerson := func() *ButterflyPerson {
		return &ButterflyPerson{
			Forenames:   "First Names",
			Surname:     "SURNAME",
			DetailLine1: "b. 26 Aug 1905, Ireland",
			DetailLine2: "d. 9 Sep 1988, Essex",
		}
	}

	root := dummyPerson()

	type entry struct {
		Generation int
		Person     *ButterflyPerson
	}

	stack := []*entry{
		{
			Generation: -1,
			Person:     root,
		},
	}

	for len(stack) > 0 {
		e := stack[0]
		stack = stack[1:]

		e.Person.Father = dummyPerson()
		e.Person.Mother = dummyPerson()
		if e.Generation < 3 {
			e.Person.Father.MarriageLine1 = "m. 9 Sep 1936"
			e.Person.Father.MarriageLine2 = "Northumberland"
		}

		if e.Generation < 5 {
			stack = append(stack, &entry{Generation: e.Generation + 1, Person: e.Person.Father})
			stack = append(stack, &entry{Generation: e.Generation + 1, Person: e.Person.Mother})
		}
	}

	ch.Root = root
	ch.Root.Father.Mother.Father = nil

	ch.TitleLine1 = "Ancestors of"
	ch.TitleLine2 = "Test Person and Test Person"
	ch.Note = "Created by Ian Davis on 14 March, 2025"
}

func scaleFont(baseFontSize float64, baseCharlen int, charlen int) float64 {
	if charlen <= baseCharlen {
		return baseFontSize
	}

	return baseFontSize * float64(baseCharlen) / float64(charlen)
}
