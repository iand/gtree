package gtree

import (
	"fmt"
	"math"
)

// FanChart defines the top-level structure for a radial ancestor chart.
// It represents a single subject (Root) and their lineage going backward.
// The chart includes a title, optional notes, and a recursive ancestry tree.
type FanChart struct {
	Title     string     // Title displayed on the diagram.
	PreTitle  string     // A line of text to place above the title, uses subtitle style
	PostTitle string     // A line of text to place below the title, uses subtitle style
	Notes     []string   // Optional lines of annotation, shown at the edge of the layout.
	Root      *FanPerson // The subject of the chart (generation 0). All ancestors trace back from this person.
}

// FanPerson represents an individual in the fan chart, including their ID, details, and their parents.
// Information is provided as freeform lines of text, for headings and details.
type FanPerson struct {
	ID       int        // Unique ID used to identify this person within the chart.
	Headings []string   // Main lines of text (e.g. name), usually rendered in larger or bolder type.
	Details  []string   // Secondary lines (e.g. birth/death years, occupation).
	Father   *FanPerson // Pointer to the father, if known. nil if unknown.
	Mother   *FanPerson // Pointer to the mother, if known. nil if unknown.
}

// FanLayoutOptions optsures the geometric and stylistic aspects of the fan chart layout.
type FanLayoutOptions struct {
	MaxGenerations int // Maximum number of generations to layout.
	Margin         Millimetre
	RadiusStep     float64 // Radial distance (in pixels) between generations.
	BaseFontSize   float64 // Font size for the closest generation (generation 0).
	GenScaleFactor float64 // Scaling factor applied to font size per generation (e.g. 0.85).
	AngularSpanDeg float64 // Total angular spread of the fan in degrees (e.g. 120 for 120°).
	Debug          bool
	TitleStyle     TextStyleOption // TitleStyle is the style of the font to use for the title of the chart.
	SubTitleStyle  TextStyleOption // SubTitleStyle is the style of the font to use for the pre and post titles of the chart.
	NoteStyle      TextStyleOption // NoteStyle is the style of the font to use for the notes of the chart.
	HeadingStyle   TextStyleOption // HeadingStyle is the style of the font to use for the heading lines of each node.
	DetailStyle    TextStyleOption // DetailStyle is the style of the font to use for the subsequent lines of each node after the first.
	DPI            int

	BackgroundColor string // BackgroundColor is the color of the background, empty for transparent/no-background fill

	titleStyle     TextStyle
	subTitleStyle  TextStyle
	noteStyle      TextStyle
	headingStyle   TextStyle
	detailStyle    TextStyle
	maxAngularSpan float64   // angular span in radians
	genScales      []float64 // scale factors for each generation
}

// DefaultAncestorLayoutOptions returns the default layout options for rendering the ancestor chart.
func DefaultFanLayoutOptions() *FanLayoutOptions {
	return &FanLayoutOptions{
		MaxGenerations: 3,
		Margin:         5,
		RadiusStep:     180,
		BaseFontSize:   16,
		GenScaleFactor: 0.85,
		AngularSpanDeg: 210,
		DPI:            300,

		TitleStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   46,
			LineHeight: 46,
			Color:      "#000",
		},
		SubTitleStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   32,
			LineHeight: 32,
			Color:      "#000",
		},
		NoteStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   16,
			LineHeight: 20,
			Color:      "#000",
		},
		HeadingStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   28,
			LineHeight: 32,
			Color:      "#000",
		},
		DetailStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   16,
			LineHeight: 18,
			Color:      "#000",
		},

		BackgroundColor: "#FFF",
	}
}

func GenerateFanLayout(chart *FanChart, opts *FanLayoutOptions) (*FanLayout, error) {
	if opts == nil {
		opts = DefaultFanLayoutOptions()
	}

	margin := opts.Margin.Pixel(opts.DPI)

	var err error
	opts.titleStyle, err = NewTextStyle(opts.TitleStyle)
	if err != nil {
		return nil, fmt.Errorf("setup title style: %w", err)
	}

	opts.subTitleStyle, err = NewTextStyle(opts.SubTitleStyle)
	if err != nil {
		return nil, fmt.Errorf("setup sub title style: %w", err)
	}

	opts.noteStyle, err = NewTextStyle(opts.NoteStyle)
	if err != nil {
		return nil, fmt.Errorf("setup note style: %w", err)
	}

	opts.headingStyle, err = NewTextStyle(opts.HeadingStyle)
	if err != nil {
		return nil, fmt.Errorf("setup heading style: %w", err)
	}

	opts.detailStyle, err = NewTextStyle(opts.DetailStyle)
	if err != nil {
		return nil, fmt.Errorf("setup detail style: %w", err)
	}

	opts.maxAngularSpan = degreesToRadians(opts.AngularSpanDeg)

	// Set up default font scaling
	opts.genScales = make([]float64, opts.MaxGenerations+1)
	opts.genScales[0] = 1
	for gen := 1; gen <= opts.MaxGenerations; gen++ {
		opts.genScales[gen] = math.Pow(opts.GenScaleFactor, float64(gen))
	}

	// Compute positioned layout nodes from the input tree
	nodes, err := assignFanNodes(chart.Root, opts)
	if err != nil {
		return nil, fmt.Errorf("assign fan nodes: %w", err)
	}

	arrangeBlurbs(nodes, opts)

	blurbs := []*Blurb{}
	blurbMap := make(map[int]*Blurb)

	for _, node := range nodes {

		blurbs = append(blurbs, node.Blurb)
		blurbMap[node.Person.ID] = node.Blurb
	}

	connectors := []*Connector{}
	for _, node := range nodes {
		child := node.Person
		childBlurb := blurbMap[child.ID]

		if child.Father != nil && child.Mother != nil {
			fb := blurbMap[child.Father.ID]
			mb := blurbMap[child.Mother.ID]
			if fb != nil && mb != nil {
				start := Point{
					X: childBlurb.X,
					Y: childBlurb.Y,
				}

				mid := midpoint(fb.X, fb.Y, mb.X, mb.Y)
				length := Pixel(distance(start, mid))

				branchLen := Pixel(float64(length) * 0.1 * opts.genScales[node.Gen])

				end := projectFrom(start, mid, length-branchLen)

				if node.Gen >= 5 {
					angle := math.Atan2(float64(end.Y-start.Y), float64(end.X-start.X)) * 180 / math.Pi
					if angle < -90 {
						angle += 180
					} else if angle > 90 {
						angle -= 180
					}
					fb.Rotation = angle
					mb.Rotation = angle
				}

				intersect, ok := intersectLineWithBlurb(start, end, childBlurb)
				if !ok {
					intersect = Point{X: childBlurb.X, Y: childBlurb.Y - childBlurb.Height/2} // fallback
				}

				connectors = append(connectors, &Connector{
					Points: []Point{
						intersect,
						end,
					},
				})

				// branches for father and mother
				if branchLen > 0 {
					fcentre := Point{X: fb.X, Y: fb.Y}
					connectors = append(connectors, &Connector{
						Points: []Point{
							end,
							projectFrom(end, fcentre, branchLen),
						},
					})
					mcentre := Point{X: mb.X, Y: mb.Y}
					connectors = append(connectors, &Connector{
						Points: []Point{
							end,
							projectFrom(end, mcentre, branchLen),
						},
					})
				}
			}
		}
	}

	if opts.Debug {
		for _, n := range nodes {
			outlines := slotBoxOutline(n.Gen, n.Slot, opts.maxAngularSpan, opts.RadiusStep)
			connectors = append(connectors, outlines...)
		}
	}

	// Centre the diagram
	width, height, dx, dy := computeLayoutBounds(blurbs, connectors, margin)

	for _, b := range blurbs {
		b.X += dx
		b.Y += dy
	}

	for _, c := range connectors {
		for i := range c.Points {
			c.Points[i].X += dx
			c.Points[i].Y += dy
		}
	}

	rootBlurb := blurbMap[chart.Root.ID]

	legend := CenterAlignedTitle(
		Point{X: rootBlurb.X, Y: height - margin},
		chart.Title,
		chart.PreTitle,
		chart.PostTitle,
		opts.titleStyle,
		opts.subTitleStyle)
	legend.Y = height - margin

	// Ensure there is enough room between root person and title
	base := rootBlurb.X + rootBlurb.Height/2
	gap := legend.Y - base - opts.titleStyle.LineHeight
	if gap < 0 {
		legend.Y += -gap
		height += -gap
	}

	// place background above title
	bg := background(width, height-legend.Height)

	// place notes
	notes := RightAlignedNotes(
		Point{X: width, Y: height},
		chart.Notes,
		opts.noteStyle)

	blurbs = append(blurbs, notes)

	layout := &FanLayout{
		legend:          legend,
		blurbs:          blurbs,
		connectors:      connectors,
		extent:          Extent{Width: width, Height: height},
		margin:          margin,
		debug:           opts.Debug,
		background:      bg,
		backgroundColor: opts.BackgroundColor,
	}

	return layout, nil
}

type fanNode struct {
	Person *FanPerson
	Gen    int     // Generation (0 = root)
	Slot   int     // 0 is the first slot
	Angle  float64 // Polar angle (radians) from vertical
	Radius float64 // Radial distance from centre
	Blurb  *Blurb
}

func assignFanNodes(root *FanPerson, opts *FanLayoutOptions) ([]*fanNode, error) {
	var nodes []*fanNode

	var nextPlaceholderID int = -1
	makePlaceholder := func(name string, gen int) *FanPerson {
		p := &FanPerson{
			ID:       nextPlaceholderID,
			Headings: []string{name},
			Details:  nil,
			Father:   nil,
			Mother:   nil,
		}
		nextPlaceholderID--
		return p
	}

	generationNodes := make(map[int][]*fanNode)

	seen := make(map[int]bool)
	var walk func(p *FanPerson, slot, gen int) error
	walk = func(p *FanPerson, slot, gen int) error {
		if gen > opts.MaxGenerations {
			return nil
		}
		if seen[p.ID] {
			return fmt.Errorf("duplicate person encountered (id=%v)", p.ID)
		}
		seen[p.ID] = true

		n := &fanNode{
			Gen:    gen,
			Slot:   slot,
			Person: p,
		}

		b := &Blurb{
			Texts: []TextSection{
				{
					Lines: p.Headings,
					Style: opts.headingStyle,
				},
				{
					Lines: p.Details,
					Style: opts.detailStyle,
				},
			},

			Alignment: AlignmentCenter,
		}

		MeasureBlurb(b)
		n.Blurb = b

		nodes = append(nodes, n)
		generationNodes[n.Gen] = append(generationNodes[n.Gen], n)

		// If both parents are unknown then stop here
		if p.Father == nil && p.Mother == nil {
			return nil
		}

		if p.Father == nil {
			p.Father = makePlaceholder("Unknown", gen)
		}
		if err := walk(p.Father, slot*2, gen+1); err != nil {
			return err
		}

		if p.Mother == nil {
			p.Mother = makePlaceholder("Unknown", gen)
		}
		if err := walk(p.Mother, slot*2+1, gen+1); err != nil {
			return err
		}

		return nil
	}

	if err := walk(root, 0, 0); err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}

	return nodes, nil
}

func angularSpanForGen(gen int, opts *FanLayoutOptions) float64 {
	compression := 1 - float64(opts.MaxGenerations-gen)/15
	return opts.maxAngularSpan * compression
}

// arrangeBlurbs places each blurb at the centre of its slot.
func arrangeBlurbs(nodes []*fanNode, opts *FanLayoutOptions) {
	for _, n := range nodes {
		angularSpan := angularSpanForGen(n.Gen, opts)
		a, b, c, d := slotBounds(n.Gen, n.Slot, angularSpan, opts.RadiusStep)

		// Centre of the trapezoid (approximate as average of 4 corners)
		x := float64(a.X+d.X+b.X+c.X) / 4.0
		y := float64(a.Y+d.Y+b.Y+c.Y) / 4.0

		n.Blurb.X = Pixel(x)
		n.Blurb.Y = Pixel(y)
		ScaleBlurb(n.Blurb, opts.genScales[n.Gen])
		MeasureBlurb(n.Blurb)

	}
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

func polarToCartesian(radius, angle float64) Point {
	adjusted := math.Pi/2 - angle // adjust so zero points vertically
	x := radius * math.Cos(adjusted)
	y := -radius * math.Sin(adjusted)
	return Point{X: Pixel(x), Y: Pixel(y)}
}

func midpoint(x1, y1, x2, y2 Pixel) Point {
	return Point{
		X: (x1 + x2) / 2,
		Y: (y1 + y2) / 2,
	}
}

func distance(p1, p2 Point) float64 {
	dx := float64(p2.X - p1.X)
	dy := float64(p2.Y - p1.Y)
	return math.Hypot(dx, dy)
}

func projectFrom(p1, p2 Point, d Pixel) Point {
	dx := float64(p2.X - p1.X)
	dy := float64(p2.Y - p1.Y)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return p1 // if both points are the same, no direction to move
	}

	ux := dx / length
	uy := dy / length

	return Point{
		X: Pixel(float64(p1.X) + float64(d)*ux),
		Y: Pixel(float64(p1.Y) + float64(d)*uy),
	}
}

// FanLayout implements the gtree.Layout interface, representing the output of a radial ancestor layout.
type FanLayout struct {
	legend          *Blurb
	blurbs          []*Blurb
	connectors      []*Connector
	margin          Pixel
	debug           bool
	extent          Extent
	background      string
	backgroundColor string
}

func (f *FanLayout) Height() Pixel            { return f.extent.Height }
func (f *FanLayout) Width() Pixel             { return f.extent.Width }
func (f *FanLayout) Margin() Pixel            { return f.margin }
func (f *FanLayout) Legend() *Blurb           { return f.legend }
func (f *FanLayout) Blurbs() []*Blurb         { return f.blurbs }
func (f *FanLayout) Connectors() []*Connector { return f.connectors }
func (f *FanLayout) Debug() bool              { return f.debug }
func (f *FanLayout) Background() string       { return f.background }
func (f *FanLayout) BackgroundColor() string  { return f.backgroundColor }

var _ Layout = (*FanLayout)(nil)

func computeLayoutBounds(blurbs []*Blurb, connectors []*Connector, margin Pixel) (Pixel, Pixel, Pixel, Pixel) {
	const maxInt = int(^uint(0) >> 1)
	const minInt = -maxInt - 1
	minX, minY := Pixel(maxInt), Pixel(maxInt)
	maxX, maxY := Pixel(minInt), Pixel(minInt)

	// Include blurb bounds
	for _, b := range blurbs {
		left := b.X - b.Width/2
		right := b.X + b.Width/2
		top := b.Y - b.Height/2
		bottom := b.Y + b.Height/2

		if left < minX {
			minX = left
		}
		if right > maxX {
			maxX = right
		}
		if top < minY {
			minY = top
		}
		if bottom > maxY {
			maxY = bottom
		}
	}

	// Include connector points
	for _, c := range connectors {
		for _, pt := range c.Points {
			if pt.X < minX {
				minX = pt.X
			}
			if pt.X > maxX {
				maxX = pt.X
			}
			if pt.Y < minY {
				minY = pt.Y
			}
			if pt.Y > maxY {
				maxY = pt.Y
			}
		}
	}

	width := (maxX - minX) + 2*margin
	height := (maxY - minY) + 2*margin

	dx := -Pixel(minX) + margin
	dy := -Pixel(minY) + margin

	return width, height, dx, dy
}

func intersectLineWithBlurb(start, end Point, b *Blurb) (Point, bool) {
	// Blurb bounding box
	left := b.X - b.Width/2
	right := b.X + b.Width/2
	top := b.Y - b.Height/2
	bottom := b.Y + b.Height/2

	// Define edges as line segments
	edges := [][2]Point{
		{{X: left, Y: top}, {X: right, Y: top}},       // top
		{{X: right, Y: top}, {X: right, Y: bottom}},   // right
		{{X: right, Y: bottom}, {X: left, Y: bottom}}, // bottom
		{{X: left, Y: bottom}, {X: left, Y: top}},     // left
	}

	for _, edge := range edges {
		if inter, ok := segmentIntersection(end, start, edge[0], edge[1]); ok {
			return inter, true
		}
	}
	return Point{}, false
}

func segmentIntersection(p1, p2, q1, q2 Point) (Point, bool) {
	// Convert to float64 for precision
	ax, ay := float64(p1.X), float64(p1.Y)
	bx, by := float64(p2.X), float64(p2.Y)
	cx, cy := float64(q1.X), float64(q1.Y)
	dx, dy := float64(q2.X), float64(q2.Y)

	// Line AB represented as a1x + b1y = c1
	a1 := by - ay
	b1 := ax - bx
	c1 := a1*ax + b1*ay

	// Line CD represented as a2x + b2y = c2
	a2 := dy - cy
	b2 := cx - dx
	c2 := a2*cx + b2*cy

	// Determinant
	det := a1*b2 - a2*b1
	if math.Abs(det) < 1e-6 {
		return Point{}, false // parallel
	}

	// Intersection point
	x := (b2*c1 - b1*c2) / det
	y := (a1*c2 - a2*c1) / det

	// Check within both segments
	if between(ax, bx, x) && between(ay, by, y) &&
		between(cx, dx, x) && between(cy, dy, y) {
		return Point{X: Pixel(x), Y: Pixel(y)}, true
	}

	return Point{}, false
}

func between(a, b, x float64) bool {
	return (x >= math.Min(a, b)-1e-6) && (x <= math.Max(a, b)+1e-6)
}

// slotBounds computes the bounding box of a slot for a given generation and slot index.
// It returns the four corner points (A, B, C, D) in Cartesian coordinates.
// A ----- B      at rTop
// |       |
// D ----- C      at rBot
func slotBounds(gen, slot int, angularSpan, radiusStep float64) (Point, Point, Point, Point) {
	totalSlots := 1 << uint(gen)
	slotAngle := angularSpan / float64(totalSlots)

	theta0 := -angularSpan/2 + float64(slot)*slotAngle
	theta1 := theta0 + slotAngle

	rTop := float64(gen) * radiusStep
	rBot := float64(gen+1) * radiusStep

	a := polarToCartesian(rTop, theta0)
	b := polarToCartesian(rTop, theta1)
	c := polarToCartesian(rBot, theta1)
	d := polarToCartesian(rBot, theta0)

	return a, b, c, d
}

// slotBoxOutline generates connectors outlining the slot's bounding box.
func slotBoxOutline(gen, slot int, angularSpan, radiusStep float64) []*Connector {
	a, b, c, d := slotBounds(gen, slot, angularSpan, radiusStep)
	return []*Connector{
		{Points: []Point{a, b}},
		{Points: []Point{b, c}},
		{Points: []Point{c, d}},
		{Points: []Point{d, a}},
	}
}

func background(width, height Pixel) string {
	const w = 1475.051
	const h = 867.406

	fudge := 1.0

	sx := fudge * float64(width) / w
	sy := fudge * float64(height) / h
	tx := fudge * -80
	//
	return fmt.Sprintf(`<path
     id="tree"
     style="display:inline;stroke-width:0;fill:#E8EEED;" transform="translate(%g,0) scale(%g, %g)"
     d="m 803.75195,18.445312 c -1.37812,-0.169531 -1.91445,1.155085 -3.75195,7.455079 -2,7.199992 -2.50039,7.898829 -5.40039,8.798828 -2.9,0.799999 -3.09961,1.201567 -3.59961,6.101562 -0.5,5.199995 -0.50079,5.299221 -5.80078,7.699219 l -5.39844,2.5 3.59961,4.400391 c 3.7,4.399995 5.59962,5.29883 16.59961,7.298828 3.8,0.699999 4.79922,1.401174 6.19922,4.201172 3,5.899994 5.20157,7.50039 10.60156,7.40039 6.29999,-0.1 13.39923,-3.700398 21.19922,-10.90039 3.5,-3.299997 6.50039,-5.300781 6.90039,-4.800782 C 846.10039,59.799608 852,77.300393 852,79.400391 852,80.700389 852.99961,81 857.09961,81 c 4.99999,0 5.1,-7.79e-4 4.5,2.699219 -1.4,7.199993 -1.4,10.401166 0,4.701172 l 1.5,-6.201172 L 871.19922,81.5 c 4.39999,-0.3 8.2,-0.399219 8.5,-0.199219 0.4,0.5 -8.39922,7.299222 -13.19922,10.199219 -1.5,0.899999 -1.40078,1.400395 0.69922,5.400391 3,5.499999 10.40001,9.998829 18,10.798829 l 5.30078,0.5 4.59961,-6.29883 c 8.59999,-12.099987 11.9,-19.801177 10.5,-25.201171 -0.5,-2.299998 -0.79884,-2.299609 -7.29883,-1.59961 -3.79999,0.5 -8.20039,1.100391 -9.90039,1.400391 -2.9,0.5 -3.30078,0.200778 -4.30078,-2.699219 -0.9,-2.399997 -0.9,-3.401172 0,-3.701172 2.6,-0.799999 8.4,-6.899613 10,-10.599609 C 895.09961,57.300002 896,53.399216 896,50.699219 896,46.499223 895.80078,46 893.80078,46 c -2.4,0 -12.20039,2.300392 -14.90039,3.400391 -1.3,0.599999 -2.10039,-1.19962 -4.40039,-11.09961 -1.5,-6.499993 -3.39961,-13.301174 -4.09961,-15.201172 L 869,19.699219 l -6.59961,9.201172 c -3.6,5.099995 -7.10117,9.399609 -7.70117,9.599609 -0.7,0.1 -5.89923,0.800001 -11.69922,1.5 -5.79999,0.599999 -10.90039,1.200781 -11.40039,1.300781 -0.4,0.1 -1.69922,-2.201565 -2.69922,-5.101562 -1,-2.899997 -2.8,-6.59883 -4,-8.298828 l -2.09961,-3 -2.20117,2 c -3,2.799997 -3.60039,2.699605 -6.90039,-1.400391 -3,-3.699996 -5.89883,-5.900392 -9.29883,-6.900391 -0.2375,-0.075 -0.45156,-0.130078 -0.64844,-0.154297 z M 944.90039,38 c -0.4,0 -1.30117,1.000783 -2.20117,2.300781 C 939.89922,44.600777 938,52.000006 938,58.5 V 65 h -10 v 4.5 c 0,4.099996 -0.2,4.5 -2.5,4.5 -1.3,0 -4.3,0.69961 -6.5,1.599609 -3.1,1.299999 -4,2.200002 -4,4 0,2.599998 2.49922,4.500393 8.19922,6.400391 2.1,0.599999 4.00078,1.6 4.30078,2 1.1,1.699998 -2.59922,5.900001 -5.69922,6.5 -1.7,0.4 -4.90156,1.500001 -7.10156,2.5 l -4.09961,1.900391 v 6.999999 c -0.1,8.69999 -1.50039,10.60039 -4.40039,5.90039 -2.2,-3.5 -7,-8.40117 -7.5,-7.70117 -5.1,6.39999 -5.69922,7.80079 -5.69922,13.30078 0,3.2 -0.29922,5.60039 -0.69922,5.40039 -0.5,-0.2 -2.5,-1.20117 -4.5,-2.20117 l -3.70117,-1.90039 -3.79883,3.40039 c -2.1,1.9 -5.30156,4.10078 -7.10156,4.80078 -1.7,0.7 -3.19922,1.89961 -3.19922,2.59961 0,1.6 5.3004,9.00079 9.90039,13.80078 2.3,2.5 3.2,4 2.5,4.5 C 876.1004,151.70078 868.2,157.5 868.5,158 c 0.2,0.3 -0.5,0.59922 -1.5,0.69922 -1.4,0.1 -3.09961,-1.90001 -6.09961,-7.5 -3.7,-6.59999 -4.29961,-8.49962 -4.59961,-15.09961 -0.5,-7.89999 -0.30117,-8.49884 5.29883,-17.79883 1.8,-3.1 1.79961,-3.20078 -0.40039,-7.80078 -1.2,-2.5 -2.19922,-5.80039 -2.19922,-7.40039 0,-2.8 -0.30079,-2.9 -6.30078,-4.000001 C 849.19922,98.49961 844.50078,98 842.30078,98 838.00079,98 836.7,97.099996 835.5,93 c -0.4,-1.399999 -1.59961,-4.400784 -2.59961,-6.800781 L 831,81.900391 827.09961,84 c -2.2,1.099999 -6.40039,2.900001 -9.40039,4 -5.7,2.199998 -8.89961,4.999225 -13.09961,11.699219 C 802.09961,103.79921 802,104.60001 802,117.5 v 13.59961 l -3.09961,-2.19922 -3.09961,-2.20117 L 796.5,112 l 0.69922,-14.599609 -5.5,-1.300782 C 785.29923,94.699611 785.80039,95.899983 783.90039,78.5 783.00039,70.600008 782.4,68.300002 782,70.5 c -0.4,1.599998 -1.2,4.100001 -2,5.5 l -1.30078,2.400391 -5,-5.400391 c -2.7,-2.999997 -5.59883,-6.800002 -6.29883,-8.5 l -1.30078,-3 L 766,64.400391 v 2.90039 l -7.5,-0.601562 c -8.49999,-0.6 -8.7,-0.399211 -6,7.300781 0.8,2.299998 1.5,4.39961 1.5,4.599609 0,0.3 -5.30001,3.199613 -11,6.09961 -0.3,0.2 -0.39922,-3.299223 -0.19922,-7.699219 0.4,-6.799993 0.19883,-8 -1.20117,-8 -4.7,0 -6.49922,-1.599614 -7.19922,-6.099609 -0.9,-5.499995 -5.9004,-11.901175 -11.40039,-14.701172 -5.39999,-2.699998 -10,-2.899998 -10,-0.5 0,0.899999 -0.49961,4.501566 -1.09961,8.101562 L 710.80078,62.199219 704.09961,61.5 c -10.09999,-0.999999 -9.79883,-0.700007 -10.29883,-8 -0.3,-4.299996 -1.00117,-7.000392 -2.20117,-8.400391 -1.8,-1.999998 -1.89883,-1.999605 -5.79883,2.400391 l -4,4.400391 L 678,49.699219 674.09961,47.5 l -2.5,2.400391 c -2.1,1.999998 -3.1,2.299218 -7,1.699218 l -4.5,-0.699218 -1.79883,8 c -1.6,7.299992 -2,8.100001 -5,9.5 -3.6,1.699998 -3.70038,1.100401 1.09961,11.90039 L 657,86.199219 662.69922,85.800781 668.30078,85.5 l 1.29883,3.900391 c 0.8,2.199997 1.60078,4.100391 1.80078,4.40039 0.3,0.2 2.99961,-0.501173 6.09961,-1.701172 C 680.6,90.899611 684.59961,90 686.59961,90 c 1.9,0 4.90117,-0.600782 6.70117,-1.300781 3,-1.299999 8.89883,-7.098833 11.29883,-11.298828 1.1,-1.799999 1.19961,-1.799608 3.09961,-0.09961 1.1,0.999999 2.80078,3.700003 3.80078,6 2.8,6.799993 3.1,7.200001 6,7.5 5.39999,0.7 11.5,-0.900784 16.5,-4.300781 4.4,-2.999997 5.19961,-3.199217 5.59961,-1.699219 2,6.199994 4.50118,9.998443 10.20117,15.398439 l 6.5,6 5.09961,-0.59961 c 2.8,-0.4 6.99883,-0.99883 9.29883,-1.29883 4.59999,-0.6 10.30078,1.09961 10.30078,3.09961 0,0.7 -1.50078,2.70039 -3.30078,4.40039 l -3.39844,3.29883 L 764,114.80078 753.69922,114.5 l -4,-7.30078 C 747.39922,103.29922 745.3,100 745,100 c -0.3,0 -0.39922,2.69961 -0.19922,6.09961 0.5,8.19999 -0.4,12.20039 -3.5,15.90039 l -2.60156,3.09961 -3.09961,-5.5 c -3.5,-6.19999 -15.90039,-19.599611 -18.40039,-20.09961 -0.9,-0.1 -1.6,-0.800391 -1.5,-1.400391 C 715.89922,97.19961 706.20078,89 704.80078,89 c -0.2,0 -1.90156,3.299223 -3.60156,7.199219 L 698,103.5 l -6.09961,1.90039 -6.09961,1.90039 0.69922,5.59961 c 0.3,3.1 0.79961,6.49961 1.09961,7.59961 1,4.1 -2.49922,-1.6004 -4.19922,-6.90039 -2.6,-8.09999 -5.20117,-12.4 -8.20117,-14.000001 C 672.59922,98.299611 655,96.300782 655,97.300781 c 0,0.3 1.89922,3.798829 4.19922,7.798829 4.89999,8.39999 15,33.1 14,34 -0.4,0.3 -2.79844,0.39961 -5.39844,0.0996 l -4.70117,-0.39844 -0.29883,-5.30078 -0.30078,-5.40039 -6.80078,3 c -3.8,1.6 -7.29844,2.90039 -7.89844,2.90039 -0.5,0 1.09922,-2.40078 3.69922,-5.30078 5.09999,-5.8 5.10078,-5.59845 1.80078,-20.39844 l -1.70117,-7.20117 -5.29883,-2.798829 C 643.40078,96.800783 641,95.099218 641,94.699219 c 0,-0.5 3.2,-2.198439 7,-3.898438 7.89999,-3.499996 7.7,-2.9004 3.5,-12.40039 C 649,72.700396 647.40077,72 637.30078,72 631.80079,72 631.8,71.999221 632.5,74.699219 633.2,77.799216 636.80078,84 637.80078,84 c 0.4,0.1 1.79961,1.000001 3.09961,2 l 2.29883,2 -4.09961,5.5 c -5.3,6.89999 -9.09961,14.50078 -9.09961,17.80078 0,3.2 -0.60079,3.49883 -5.80078,3.79883 -2.3,0.2 -4.19922,0.39961 -4.19922,0.59961 0,0.9 3.89922,6.40039 8.19922,11.40039 4.39999,5.19999 6.80078,9.50117 6.80078,12.20117 0,0.6 -1.30039,-0.40078 -2.90039,-2.30078 l -2.90039,-3.5 -3.79883,3.09961 c -2.2,1.7 -4.59961,4.4 -5.59961,6 -1.5,2.7 -1.50117,3.3004 0.29883,8.90039 1.1,3.3 2.19961,8.99961 2.59961,12.59961 0.6,6.49999 0.70039,6.80039 4.40039,8.90039 2.2,1.2 3.90039,2.70078 3.90039,3.30078 0,0.7 -0.7,7.49923 -1.5,15.19922 -0.8,7.69999 -1.5,14.59922 -1.5,15.19922 0,3 -1.90078,0.79999 -5.30078,-6.5 -4.2,-9.09999 -6.89922,-18.39963 -9.69922,-34.09961 -2.8,-16.19998 -2.7,-20.19962 1,-25.59961 l 3,-4.40039 -3.5,-3.29883 c -3.8,-3.5 -5.99922,-9.80118 -6.69922,-18.70117 l -0.30078,-4.40039 6,-1.39844 c 7.09999,-1.6 10.5,-4.50001 13.5,-11.499999 l 2.09961,-5 -2.69922,-2.601562 c -3.2,-3.199997 -7.2,-11.300005 -8,-16.5 C 617.00039,70.699221 616.29922,69 615.69922,69 c -0.6,0 -5.80001,2.499612 -11.5,5.599609 l -10.5,5.5 -6.79883,-2.5 -6.70117,-2.5 0.5,11.701172 C 580.89922,93.200775 581.3,99 581.5,99.5 c 0.1,0.6 3.40079,2.80039 7.30078,4.90039 3.9,2.2 6.89922,4.19961 6.69922,4.59961 -0.3,0.4 -3.69922,0.10039 -7.69922,-0.59961 -3.99999,-0.8 -7.60039,-1.40039 -7.90039,-1.40039 -0.3,0 -0.19961,2.00039 0.40039,4.40039 1.5,6.79999 0.29922,11.20039 -4.30078,14.40039 -3.3,2.3 -3.79922,3.19922 -3.69922,6.19922 0.2,3.3 0.1,3.50078 -4.5,4.30078 l -4.70117,0.89844 4.59961,8.80078 c 2.5,4.8 4.40117,9.20039 4.20117,9.90039 -0.3,0.7 -4.3004,-1.50039 -9.90039,-5.40039 -13.79999,-9.59999 -21,-13.09961 -21,-10.09961 0,2.4 -5.2004,7.69961 -12.40039,12.59961 -3.6,2.5 -6.59961,4.99961 -6.59961,5.59961 0,1.5 13.50001,7.20117 20.5,8.70117 6.79999,1.4 9.49922,-0.20079 13.19922,-7.80078 1.5,-3 3.00078,-5.5 3.30078,-5.5 0.3,0 2.90078,2.40078 5.80078,5.30078 l 5.29883,5.39844 -2.59961,8.70117 c -1.4,4.9 -2.5,10.09922 -2.5,11.69922 0,4 -1.4,3.60039 -3.5,-1.09961 -1,-2.2 -2.50078,-4.60078 -3.30078,-5.30078 -1.1,-0.9 -19.19922,-2.89961 -19.19922,-2.09961 0,0.1 1.1,3.10078 2.5,6.80078 L 544,190 h -4.09961 c -2.2,0 -6,-0.49961 -8.5,-1.09961 L 527,187.80078 v -7.90039 C 527,172.2004 526.90039,171.9 524.40039,171 c -5.29999,-2 -10.59961,-2.89922 -10.09961,-1.69922 0.2,0.7 0.99922,3.69883 1.69922,6.79883 l 1.30078,5.70117 L 513,185.59961 c -7.49999,6.79999 -7.29922,6.8 -12.69922,4 -8.19999,-4.1 -13.2,-8.39961 -15.5,-13.09961 -1.2,-2.5 -1.90156,-4.7 -1.60156,-5 0.3,-0.4 2.90117,0.50039 5.70117,1.90039 2.8,1.4 5.9,2.59961 7,2.59961 2.9,0 6.40001,-3.59962 12,-12.09961 l 4.90039,-7.59961 L 509.90039,152 c -4,-5.89999 -4.60117,-10.09922 -2.20117,-14.69922 1.1,-2 2.40039,-3.8 2.90039,-4 0.5,-0.2 1.10117,3.40001 1.20117,8 0.2,4.5 0.49922,8.39961 0.69922,8.59961 1,0.9 10.20078,-2.19961 12.30078,-4.09961 2.6,-2.5 2.99883,-8.00118 0.79883,-13.20117 C 524.19961,129.19961 518.4,122 517,122 c -0.4,0 -1.1,1.49922 -1.5,3.19922 l -0.69922,3.30078 -0.90039,-4.19922 c -1,-5.29999 -4.50117,-13.30078 -5.70117,-13.30078 -0.6,0 -2.39961,1.6 -4.09961,3.5 -2.8,3.2 -4.0004,6.40001 -7.90039,20.5 -0.9,3.6 -1.39922,3.19921 -2.69922,-2.30078 -0.5,-2.5 -0.90078,-2.79883 -3.30078,-2.29883 -1.5,0.3 -5.09922,-0.30078 -8.19922,-1.30078 -6.49999,-2.3 -6.39961,-2.29882 -8.09961,7.20117 -1.1,5.9 -1.60117,6.99883 -3.70117,7.79883 -1.4,0.6 -3.09883,0.8 -3.79883,0.5 -0.6,-0.2 -1.90078,1.20117 -2.80078,3.20117 -1,2 -2.59922,4.79961 -3.69922,6.09961 l -2,2.5 -2.40039,-3 c -1.3,-1.6 -2.4,-3.30117 -2.5,-3.70117 0,-0.5 -2.20039,0.80078 -4.90039,2.80078 l -5,3.59961 -5.79883,-1.69922 c -3.2,-0.9 -8.60156,-2.79961 -12.10156,-4.09961 -3.5,-1.4 -6.49922,-2.40156 -6.69922,-2.10156 -0.7,0.6 4.3,18.50157 6.5,23.60156 1.1,2.5 1.90078,4.59922 1.80078,4.69922 -0.2,0.1 -2.50078,0.70039 -5.30078,1.40039 -2.7,0.7 -6.70078,1.69922 -8.80078,2.19922 l -3.69922,1 5.90039,4.90039 c 3.2,2.7 6.8,6.20078 8,7.80078 3.1,4.4 8.99961,17.60001 10.59961,24 1.7,6.49999 1.79961,7.19961 0.59961,9.09961 -1.6,2.5 -5.8,-1.19961 -7,-6.09961 l -1,-4.30078 0.5,4.69922 c 0.3,2.6 0.20078,6.30078 -0.19922,8.30078 -0.6,3.3 -0.79961,3.5 -4.59961,3.5 -3.99999,0 -4.00039,-7.8e-4 -3.40039,-3.30078 0.3,-1.7 0.89883,-6.29961 1.29883,-10.09961 l 0.60156,-6.90039 L 414.5,213.5 c -3.2,0.4 -7.00078,1.09961 -8.30078,1.59961 -1.7,0.7 -3.59844,0.40039 -6.89844,-1.09961 C 394.80079,212 393,212.10078 393,214.30078 c 0,0.7 -0.40039,1.79961 -0.90039,2.59961 -0.7,1.1 -1.60039,0.89922 -4.40039,-1.30078 -1.9,-1.5 -5.09961,-2.99883 -7.09961,-3.29883 L 377,211.69922 V 216 c 0,3.6 -0.19922,4.10039 -1.69922,3.40039 -3.79999,-1.6 -10.30078,-1.6 -12.80078,0 -1.4,0.9 -2.5,1.19922 -2.5,0.69922 0,-0.5 -1.19961,-2.29961 -2.59961,-4.09961 -2.6,-3 -6.20118,-5.00039 -11.70117,-6.40039 -2.7,-0.6 -2.7,-0.60039 -2.5,3.59961 0.3,4.09999 0.2,4.20078 -3.5,5.30078 l -3.89844,1.09961 3.39844,3.30078 c 4.39999,4.5 10.30039,7.09961 15.90039,7.09961 h 4.70117 L 361,236.5 c 1.3,7.09999 2.8004,9.10079 10.40039,14.30078 2.6,1.8 5.50039,4.59922 6.40039,6.19922 3.5,6.29999 12.89924,5.49921 28.69922,-2.30078 6.19999,-3.2 11.60039,-5.69922 11.90039,-5.69922 0.4,0 0.3,2.49961 0,5.59961 l -0.70117,5.70117 6.30078,-0.70117 6.19922,-0.69922 -0.39844,4 c -0.2,2.3 -0.30156,4.09961 -0.10156,4.09961 0.1,0 3.7,-2.2 8,-5 4.29999,-2.7 8.1,-5 8.5,-5 1.7,0 13.50039,21.59922 12.40039,22.69922 -0.8,0.8 -15.99962,-3.89844 -22.59961,-6.89844 C 416.30002,263.70079 402,258.3 402,260 c 0,1 7.59922,7.9 12.19922,11 2.1,1.4 3.7,2.90078 3.5,3.30078 -0.1,0.5 -5.59923,3.09883 -12.19922,5.79883 -6.59999,2.7 -15.1,6.3 -19,8 -10.89999,4.79999 -10.99961,4.80039 -8.09961,0.40039 1.4,-2.1 2.59961,-4.50039 2.59961,-5.40039 0,-2.6 -13.00039,-13.80039 -14.40039,-12.40039 -0.3,0.4 -0.59961,5.5 -0.59961,11.5 0,12.59999 1.39921,11.80156 -12.30078,6.60156 -4.2,-1.6 -7.69922,-3.30117 -7.69922,-3.70117 0,-0.5 2.70039,-2.5 5.90039,-4.5 3.3,-2 7.1,-4.99922 8.5,-6.69922 2.3,-2.7 2.59961,-3.90079 2.59961,-11.30078 C 363,254.09962 360.70039,244 358.90039,244 c -0.6,0 -2.40039,0.9 -3.90039,2 l -2.80078,2 -3.59961,-5 c -2.2,-3 -4.09883,-4.80039 -4.79883,-4.40039 -0.7,0.4 -0.80039,0.3 -0.40039,-0.5 1.1,-1.7 -1.0004,-2.10039 -13.90039,-2.40039 l -12,-0.19922 0.30078,3.19922 c 0.4,2.7 -0.10117,3.50039 -3.70117,5.90039 -4,2.6 -4.2,2.80039 -3.5,6.90039 l 0.59961,4.19922 -4,0.60156 c -2.1,0.4 -5.59961,1.39922 -7.59961,2.19922 l -3.69922,1.59961 2.69922,3.5 c 3.2,4.1 7.49961,6.60078 10.09961,5.80078 1.3,-0.4 1.5,-0.29961 0.5,0.40039 -1.1,0.7 -0.99922,0.99961 0.30078,1.59961 0.9,0.3 2.79961,0.59961 4.09961,0.59961 2.1,0 2.40039,0.4 2.40039,4 0,4.4 0.50079,4.59961 7.30078,3.09961 3.7,-0.9 3.99844,-0.8 5.89844,2.5 2.1,3.5 4.40039,4.20039 7.90039,2.40039 1.5,-0.9 2.79961,-0.30039 7.09961,3.09961 2.9,2.2 7.90117,5.39961 11.20117,7.09961 5.79999,2.9 6.39962,3 16.09961,2.5 7.39999,-0.4 12.70079,-1.29922 20.30078,-3.69922 5.7,-1.8 10.39961,-3.10039 10.59961,-2.90039 0.2,0.2 0.60039,2.9 0.90039,6 l 0.5,5.59961 -3.90039,1.20117 c -2.1,0.7 -3.90039,1.59922 -3.90039,2.19922 0,1.5 4.40039,8.00078 7.90039,11.80078 3.3,3.6 3.89883,5.09961 1.79883,5.09961 -1.5,0 -5.89961,4.49922 -7.59961,7.69922 -1.1,2.1 -0.69921,2.50078 4.80078,5.30078 6.79999,3.4 10.79923,3.7 17.19922,1.5 L 424.40039,335 428,338.5 c 2,1.9 3.99961,3.5 4.59961,3.5 0.5,0 3.60078,-3.2 6.80078,-7 3.2,-3.9 6.19922,-7 6.69922,-7 1.4,0 8.9,9.4 11,14 4.79999,9.99999 7.5,20.20002 10.5,39 l 0.59961,3.5 -4.39844,-4.09961 C 451.60079,368.8004 437.19999,358.79999 425,353.5 c -13.29999,-5.89999 -18.30001,-7.00039 -31.5,-6.90039 -10.59999,0 -13.60079,0.40117 -25.80078,3.70117 -9.89999,2.7 -15.80001,3.69922 -21,3.69922 -14.49999,0 -12.69842,-3.00001 6.10156,-9.5 0.8,-0.3 -7.8e-4,-1.59961 -2.30078,-3.59961 -1.9,-1.7 -3.5,-3.7 -3.5,-4.5 0,-3.1 8.90079,-4.60117 13.30078,-2.20117 1.2,0.6 6.49923,1.20078 11.69922,1.30078 5.19999,0.1 10.39961,0.50039 11.59961,0.90039 1.8,0.7 2.40078,0.1 4.30078,-4.5 2.8,-6.49999 3.69922,-14.50118 2.19922,-19.20117 -1.9,-5.6 -5.79962,-8.19961 -16.09961,-10.59961 -5,-1.2 -9.50078,-2.09961 -9.80078,-2.09961 -0.4,0 -0.59844,2.20078 -0.39844,4.80078 l 0.39844,4.79883 -5.59961,2.70117 c -3.1,1.5 -8.09961,3.29844 -11.09961,3.89844 l -5.40039,1.10156 3.90039,4.5 c 2.1,2.5 3.70039,4.89883 3.40039,5.29883 -0.4,0.7 -10.20117,-1.2 -12.20117,-2.5 -1.5,-0.9 -0.19922,-4.29962 3.80078,-10.09961 3.6,-5.19999 3.79922,-7.49922 0.69922,-10.19922 -2.1,-2 -9.89883,-6.30078 -11.29883,-6.30078 -0.4,0 -0.40117,4.29961 -0.20117,9.59961 l 0.5,9.59961 -5.29883,-0.59961 c -4.99999,-0.6 -5.69961,-0.39961 -10.09961,2.90039 -2.7,2 -7.60156,6.29922 -11.10156,9.69922 l -6.39844,6.20117 1.59961,3.40039 c 2.4,4.9 2.10039,5.3 -3.09961,5 -4.49999,-0.3 -4.8,-0.50078 -4.5,-2.80078 0.4,-2.9 2.49884,-6.09923 10.29883,-15.69922 7.19999,-8.79999 8.99961,-12.40079 9.59961,-19.30078 0.8,-8.09999 -3,-18.20078 -6,-16.30078 -0.5,0.3 -0.69844,0.10156 -0.39844,-0.39844 0.4,-0.6 -7.8e-4,-2.70117 -0.80078,-4.70117 -0.8,-1.9 -1.49961,-4.79883 -1.59961,-6.29883 l -0.0996,-2.80078 -1,3.19922 c -1.1,3.4 -4.10078,4.90039 -5.80078,2.90039 -0.5,-0.7 -0.70039,-1.49883 -0.40039,-1.79883 0.3,-0.3 -0.69922,-1.50117 -2.19922,-2.70117 -2.1,-1.6 -3.50078,-2 -5.80078,-1.5 -4.2,0.8 -5.5,2.30001 -7,8 l -1.19922,4.80078 -14.20117,-0.30078 L 251,287.80078 l 0.90039,2.5 c 0.5,1.3 3.5,7.39962 6.5,13.59961 5.6,11.09999 5.59923,11.09961 11.69922,14.09961 11.49999,5.49999 11.60116,6.79922 0.70117,6.69922 -4.59999,0 -10.4,0.30117 -13,0.70117 -3.4,0.6 -5.90117,0.39883 -9.20117,-0.70117 -2.5,-0.9 -4.59961,-1.49883 -4.59961,-1.29883 0,0.1 1.30039,3.79922 2.90039,8.19922 4.9,13.69999 4.79882,17.30078 -0.20117,11.80078 -2.5,-2.7 -5,-7.7004 -8.5,-17.40039 l -2.09961,-5.5 v 5.69922 c -0.1,3.1 0.60039,8.90039 1.40039,12.90039 0.9,4 1.29961,7.39961 1.09961,7.59961 -0.2,0.3 -2.39883,-0.59883 -4.79883,-1.79883 -2.8,-1.5 -7.00118,-2.40078 -12.20117,-2.80078 L 213.69922,341.5 212.5,338 211.30078,334.5 210.5,339.90039 c -0.7,5.3 -0.70039,5.4 2.59961,8 1.9,1.4 4.80039,3.9 6.40039,5.5 l 3,3 -6,0.90039 c -3.3,0.5 -6.10078,0.99922 -6.30078,1.19922 -0.2,0.1 0.40117,2 1.20117,4 0.9,2.1 1.59961,4.3 1.59961,5 0,0.7 -4.19922,2.7 -9.19922,4.5 -10.89999,3.8 -14.00077,5.39961 -8.80078,4.59961 13.89999,-2.1 16.2004,-1.99961 22.90039,1.40039 5.5,2.8 6.99922,3.19922 11.69922,2.69922 L 235,380.09961 v 3.5 c 0,3.2 0.20078,3.50078 2.30078,2.80078 4.9,-1.5 11.59961,-5.49961 13.59961,-8.09961 1.9,-2.4 2.1,-3.40157 1.5,-9.60156 l -0.70117,-7 2.40039,0.60156 c 1.3,0.3 6.50039,0.89922 11.40039,1.19922 13.29999,0.9 16.90079,2.10001 22.80078,7.5 2.9,2.7 6,6.20039 7,7.90039 L 297.09961,382 h -2.5 c -3.8,0 -11.29962,-3.10079 -18.59961,-7.80078 l -6.5,-4.09961 -1.30078,3.70117 c -2.3,6.29999 -2.59884,6.59961 -13.29883,10.59961 -5.69999,2.2 -10.50117,3.99961 -10.70117,4.09961 -0.1,0.2 0.20078,2.4 0.80078,5 0.7,3 2.50039,6.60039 4.90039,9.40039 4.2,5.1 16.6,24 16,24.5 -0.2,0.1 -7.6004,2.30039 -16.40039,4.90039 -8.79999,2.5 -18,5.49961 -20.5,6.59961 L 224.5,441 l 10,-0.5 c 5.49999,-0.3 14.30001,-1.2 19.5,-2 5.19999,-0.8 14.00001,-1.5 19.5,-1.5 11.69999,0 17.2004,1.90001 27.40039,9 l 6.19922,4.40039 -2.5,1.29883 c -5.69999,3 -14.99962,4.6 -23.59961,4 -8.19999,-0.5 -9.80002,-0.99845 -28,-7.89844 -6.39999,-2.4 -22.69923,-3.00039 -29.69922,-0.90039 L 219,448.09961 l 1.69922,4.20117 c 1.9,4.9 5.7,11.29883 8.5,14.29883 1.8,2.1 1.80117,2.40079 0.20117,8.80078 -1.9,7.59999 -1.39962,7.49961 -12.09961,2.59961 -4.89999,-2.3 -12.70117,-6.89961 -16.20117,-9.59961 -0.2,-0.2 0.99961,-1.99961 2.59961,-4.09961 1.6,-2.1 3.30117,-4.3 3.70117,-5 0.8,-1.5 -3.5,-12.30157 -7,-17.60156 -1.4,-2.2 -2.3,-4.09844 -2,-4.39844 0.3,-0.3 2.5,1.09961 5,3.09961 3.8,3.2 4.89883,3.59961 7.79883,3.09961 6.19999,-1.2 6.4,-1.49923 5,-9.69922 -0.7,-3.99999 -1.59961,-7.60156 -2.09961,-8.10156 -0.4,-0.4 -4.89883,-1.09844 -9.79883,-1.39844 l -9,-0.60156 -0.70117,4.5 -0.69922,4.5 L 184,433.90039 c -5.39999,0.7 -10.39961,0.99961 -11.09961,0.59961 -0.9,-0.6 -1,0.49961 -0.5,4.59961 1.7,12.09999 2.3,22.50039 1.5,23.90039 -2.5,4.5 -2.49961,4.50039 -7.59961,0.90039 -7.29999,-5.29999 -7.10039,-5.20117 -6.90039,-1.70117 0.1,3 -0.19961,3.2 -4.09961,4 -3.99999,0.9 -13.30078,7.80039 -13.30078,9.90039 0,0.6 1.30078,1.4 2.80078,2 2.2,0.8 3.09961,1.90118 4.09961,5.70117 0.7,2.6 1.89883,5.39844 2.79883,6.39844 1.4,1.6 1.30039,2 -1.59961,5.5 -3.7,4.49999 -3.29999,5.10078 6.5,8.30078 7.89999,2.7 8.99999,4.10078 2.5,3.30078 -5,-0.6 -20.09961,2.09961 -20.09961,3.59961 0,0.4 3.1,3.69883 7,7.29883 3.8,3.5 7,6.9 7,7.5 0,0.5 -2.29961,1.60039 -5.09961,2.40039 l -5.09961,1.40039 6.09961,7.5 c 3.4,4.1 6.90039,9.59961 7.90039,12.09961 l 1.69922,4.59961 7.30078,-1.79883 c 4,-1 9.9,-3.40039 13,-5.40039 3.2,-1.9 6.3,-3.5 7,-3.5 0.6,0 1.69883,-2.09961 2.29883,-4.59961 0.7,-2.6 1.60078,-4.3 2.30078,-4 12.79999,5.4 14.5004,5.9 20.90039,6 5.5,0.1 6.59922,0.39961 6.19922,1.59961 -0.3,0.8 0.60039,0.19961 1.90039,-1.40039 3.1,-3.7 4.59961,-8.99923 4.59961,-16.19922 0,-7.29999 -1.3,-12.10079 -4.5,-16.80078 l -2.5,-3.69922 2.30078,-0.80078 c 1.2,-0.5 7.79923,-3.40039 14.69922,-6.40039 6.89999,-3 17.90001,-6.89961 24.5,-8.59961 8.79999,-2.3 15.20001,-4.69883 24,-9.29883 6.59999,-3.4 16.39922,-8.40156 21.69922,-11.10156 13.49999,-6.99999 14.50158,-6.89882 33.60156,2.70117 24.19998,12.09999 40.19963,18.00001 63.59961,23.5 l 12,2.69922 -8.5,5.59961 C 391.0004,511.39921 378.9996,516.5 365.59961,519 344.29963,523 316.59998,513.80037 298.5,496.90039 c -5.89999,-5.49999 -11.59962,-8.90078 -17.59961,-10.30078 -2.4,-0.5 -2.80039,0.001 -6.90039,8.70117 -2.5,5.1 -4.59922,9.39961 -4.69922,9.59961 -0.2,0.3 -1.30156,-0.50117 -2.60156,-1.70117 -2.5,-2.4 -10.39844,-6.09922 -12.89844,-6.19922 -1.5,0 -1.80078,0.90039 -1.80078,4.40039 0,3.9 -0.39922,4.69922 -3.19922,6.69922 -1.8,1.2 -5.8,4.70117 -9,7.70117 l -5.60156,5.5 9.20117,7.59961 c 5,4.2 10.49961,8.59883 12.09961,9.79883 3.7,2.7 4.09998,2.50039 -15,5.40039 l -15,2.30078 6,1.29883 c 4.2,0.9 9.99923,1.10156 19.19922,0.60156 13.09999,-0.6 31.70039,0.99883 32.90039,2.79883 0.3,0.5 -1.39883,1.8 -3.79883,3 -3.89999,1.9 -4.60078,2.80118 -7.30078,9.70117 -1.6,4.1 -3.10078,7.89922 -3.30078,8.19922 -0.1,0.4 2.40117,1.30039 5.70117,1.90039 3.7,0.7 7.19961,2.19922 9.09961,3.69922 1.6,1.4 3.4,2.40117 4,2.20117 0.5,-0.1 3.70039,-2.60078 6.90039,-5.30078 4.8,-4 6,-5.6 6,-8 0.1,-3.5 -2.90117,-10.40039 -5.20117,-12.40039 -4,-3.2 -1.99959,-3.59883 16.40039,-3.29883 20.49998,0.3 19.40077,-0.40116 13.30078,8.79883 -7.49999,11.29999 -22.60001,23.99961 -36,30.09961 -3.1,1.4 -14.69962,4.8 -25.59961,7.5 -10.99999,2.7 -22.70156,5.90078 -26.10156,7.30078 -3.4,1.3 -11.99923,3.6 -19.19922,5 -7.09999,1.5 -13.2,2.89961 -13.5,3.09961 -0.2,0.3 0.60078,0.69961 1.80078,1.09961 3.2,0.8 2.79921,1.80078 -2.30078,5.30078 -2.5,1.7 -4.5,3.20039 -4.5,3.40039 0,0.2 1.8,1.59961 4,3.09961 2.2,1.5 4,3.00039 4,3.40039 0,0.3 -1.90078,1.9 -4.30078,3.5 -9.69999,6.79999 -16.79844,12.89883 -15.89844,13.79883 0.5,0.5 3.29883,1.2 6.29883,1.5 4.1,0.4 6.70039,3.9e-4 10.90039,-1.59961 3,-1.2 5.60078,-2.09961 5.80078,-2.09961 0.1,0 0.89883,1.8 1.79883,4 0.9,2.2 2.00039,4 2.40039,4 1.7,0 7.09961,-7.7004 9.09961,-12.90039 1.1,-3 2.29961,-8.19961 2.59961,-11.59961 l 0.60156,-6.19922 -3.90039,-1.5 c -2.2,-0.9 -5.80117,-2.10117 -8.20117,-2.70117 C 212.89922,616.39961 211,615.5 211,615 c 0,-1.5 7.60079,-3.10078 12.80078,-2.80078 3.8,0.2 4.59844,0.60156 4.39844,2.10156 -0.4,2.1 1.10039,2.19961 2.90039,0.0996 1.7,-2 10.00079,-6.3 15.80078,-8 8.29999,-2.4 38.09961,-6.4 38.09961,-5 0,0.3 -1.49922,3.49961 -3.19922,7.09961 l -3.30078,6.59961 -10.19922,-0.69922 -10.20117,-0.70117 -4,3.20117 L 250,620.09961 252.19922,624 c 2.1,3.6 2.20156,3.89922 0.60156,5.69922 -2.1,2.3 -5.40117,9.7 -6.20117,14 L 246,647 h 10.40039 c 10.19999,0 10.3004,0 14.90039,-3.5 7.99999,-6.09999 8.89961,-8.20079 9.09961,-19.80078 0.1,-9.39999 0.4,-10.69844 3.5,-16.89844 2.7,-5.39999 4.19961,-7.10039 7.59961,-8.90039 3.7,-2 4.70039,-2.09961 6.90039,-1.09961 3.2,1.4 3.19922,2.70001 0.19922,9.5 l -2.29883,5.39844 4.29883,-1.59961 c 2.3,-0.8 4.50117,-1.29961 4.70117,-1.09961 0.2,0.3 -0.10078,3.6 -0.80078,7.5 -0.6,3.8 -1.1,7.09922 -1,7.19922 0.1,0.2 3.9,-0.79922 8.5,-2.19922 4.6,-1.4 8.49922,-2.5 8.69922,-2.5 0.1,0 -4.50001,4.7004 -10.5,10.40039 l -10.69922,10.5 5,5 c 5.19999,5.3 7.7004,6.09961 12.90039,4.09961 2.1,-0.8 2.69961,-0.6 3.59961,1 0.5,1 1,2.80039 1,3.90039 0,2.5 1.5,3.39922 4.5,2.69922 2.5,-0.6 11.5,-13.6 11.5,-16.5 0,-0.9 -1.60039,-4.40039 -3.40039,-7.90039 L 331.19922,626 l 3.90039,-2.80078 c 5.79999,-4.1 4.70116,-4.3 -2.79883,-0.5 l -6.70117,3.40039 -0.90039,-2.29883 c -1.4,-3.6 7.8e-4,-11.5 2.80078,-16.5 2.9,-5.19999 3.09961,-7.30157 1.09961,-11.10156 -1.7,-3.1 -7.10039,-8.39961 -10.40039,-10.09961 -2.2,-1.2 -1.9996,-1.6004 9.90039,-13.90039 6.79999,-6.99999 13.79961,-13.7 15.59961,-15 1.8,-1.3 7.50157,-3.89883 12.60156,-5.79883 11.59999,-4.49999 14.09885,-5.89962 31.79883,-18.59961 19.39998,-13.99998 25.99962,-18.00118 37.09961,-22.70117 12.39999,-5.3 23.9008,-7.39883 40.80078,-7.29883 17.19998,0 40.10039,3.19844 35.40039,4.89844 -3.8,1.3 -17.60118,9.00039 -24.20117,13.40039 -9.49999,6.39999 -24.39884,18.2004 -33.79883,26.90039 -10.29999,9.59999 -18.50118,15.40079 -28.20117,20.30078 -8.29999,4.1 -8.79962,4.19922 -18.59961,4.19922 -12.69999,0 -16.99883,-1.80001 -20.29883,-8.5 -1.3,-2.4 -2.30078,-3.89922 -2.30078,-3.19922 0,1 -2.1004,1.09961 -8.90039,0.59961 -4.8,-0.4 -11.30039,-0.4 -14.40039,0 l -5.5,0.69922 1.90039,5.59961 1.90039,5.60156 -3.59961,5.29883 c -2,2.9 -4.9,6.5 -6.5,8 C 337.30039,588.09961 336,589.7 336,590 c 0,1.4 15.80001,3.5 26,3.5 10.29999,0 10.59962,-0.10078 17.09961,-3.80078 3.6,-2 8.8,-4.39961 11.5,-5.09961 7.19999,-2.1 26.49961,-5.79883 27.09961,-5.29883 0.2,0.3 -1.09844,2.89844 -2.89844,5.89844 -9.49999,14.69998 -29.9008,25 -52.80078,26.5 -10.89999,0.8 -9.09999,2.00117 3,2.20117 8.99999,0.1 20.10001,-1.90078 29.5,-5.30078 3.8,-1.4 7.10078,-2.39922 7.30078,-2.19922 1.2,1.3 -18.90118,16.60001 -28.20117,21.5 -2.6,1.4 -2.80001,1.29921 -8,-3.80078 -3.7,-3.7 -5.79883,-5.1 -6.79883,-4.5 -1.1,0.7 -12.90156,9.60039 -15.10156,11.40039 -0.4,0.4 0.80078,2.60039 2.80078,4.90039 2.2,2.8 3.5,5.29922 3.5,7.19922 V 646 h 8 c 9.29999,0 9,-0.4996 6.5,11.90039 -0.9,4 -1.39961,7.59883 -1.09961,7.79883 0.3,0.3 1.6,0.60117 3,0.70117 1.3,0 2.29961,0.3 2.09961,0.5 C 368,667.30039 383,680 384,680 c 1.4,0 5.09961,-6.80079 7.59961,-13.80078 l 2.30078,-6.69922 5.5,-0.30078 5.5,-0.29883 -1.09961,-2.70117 c -0.7,-1.5 -3.40156,-6.19844 -6.10156,-10.39844 -4,-6.29999 -5.4,-7.80078 -8,-8.30078 C 381.29923,635.9 377,634.69961 377,634.09961 c 0,-0.4 3.8,-2.99922 8.5,-5.69922 15.49998,-8.99999 20.60039,-12.49961 19.90039,-13.59961 -0.5,-0.7 -0.2,-0.80078 0.5,-0.30078 2.3,1.3 9.99923,-7.50001 15.69922,-18 7.69999,-13.99999 27.60041,-35.59962 45.90039,-49.59961 6.99999,-5.39999 21.99922,-14.1 28.19922,-16.5 2.3,-0.9 4.10039,-1.90078 3.90039,-2.30078 -0.2,-0.3 0.80117,-0.69922 2.20117,-0.69922 1.5,-0.1 6.99923,-0.90078 12.19922,-1.80078 7.09999,-1.3 12.90001,-1.59883 23,-1.29883 22.99998,0.8 42.80002,6.09962 65.5,17.59961 40.19996,20.29998 73.00041,51.30043 95.40039,89.90039 9.69999,16.79998 19.29922,43.89924 22.69922,64.19922 2.4,14.09999 2.19961,38.89962 -0.40039,50.59961 -7.19999,32.39997 -24.99925,59.39963 -50.69922,77.09961 -5.69999,3.99999 -19.10001,10.90118 -32,16.70117 -7.09999,3.2 -15.79922,7.09922 -19.19922,8.69922 -3.4,1.6 -6.60156,2.90039 -7.10156,2.90039 -1.2,0 -2.19922,-2.39922 -2.19922,-5.19922 0,-1.9 -0.20078,-1.9 -2.30078,-0.5 C 600.99922,850.00078 582,852.9 582,850 c 0,-0.5 0.90039,-3.2 1.90039,-6 2.3,-5.99999 1.59961,-5.7996 -2.90039,0.90039 -1.9,2.8 -3.79922,5.09961 -4.19922,5.09961 -0.9,0 -1.60117,-2.50001 -2.70117,-9.5 l -0.79883,-5 L 572,838.69922 c -1.4,3.89999 -1.69922,4.00156 -4.69922,1.10156 L 565,837.69922 v 5.5 c 0,7.09999 -0.7,8.80078 -3.5,8.80078 -3.3,0 -4.1,-1.50001 -6,-10.5 -1.9,-9.29999 -3.69922,-15.69961 -4.19922,-15.09961 -0.2,0.2 -0.80078,5.89962 -1.30078,12.59961 -0.6,6.79999 -1.2,12.50078 -1.5,12.80078 C 547.2,853.10078 530.80039,850 527.90039,848 L 525,845.80078 l -4.40039,4.89844 c -2.4,2.6 -4.69961,4.50156 -5.09961,4.10156 -0.4,-0.4 0.50039,-5.20079 1.90039,-10.80078 1.4,-5.49999 2.59961,-10.49961 2.59961,-11.09961 0,-2.2 -6.09922,3.49962 -9.19922,8.59961 -1.8,3 -3.70078,5.29961 -4.30078,5.09961 -0.5,-0.1 -1.49961,-1.99922 -2.09961,-4.19922 L 503.30078,838.5 502.5,846 c -0.4,4.1 -0.89922,7.79922 -1.19922,8.19922 -0.8,1.4 -6.3004,-2.19845 -13.40039,-8.89844 l -7,-6.60156 0.69922,6.40039 0.59961,6.40039 -3.69922,1.80078 c -2.1,1 -3.9,1.7 -4,1.5 -0.1,-0.2 -0.70078,-9.30079 -1.30078,-20.30078 -0.6,-10.99999 -1.69844,-23.6 -2.39844,-28 l -1.20117,-8 -1.29883,9 c -0.7,4.9 -1.50117,16.00001 -1.70117,24.5 -0.3,8.49999 -0.59961,15.7 -0.59961,16 0,0.3 -2.09922,-4.4004 -4.69922,-10.40039 -5.59999,-13.09999 -13.10079,-24.60001 -19.30078,-29.5 -3.1,-2.5 -3.80039,-2.79961 -2.40039,-1.09961 4.49999,5.19999 9.30117,13.80001 12.70117,23 3.7,9.89999 8.09844,25.89922 7.39844,26.69922 -0.3,0.2 -3.2,-3.29844 -6.5,-7.89844 -6.1,-8.49999 -9.79922,-12.50156 -10.69922,-11.60156 -0.3,0.3 0.2,4.00039 1,8.40039 0.8,4.3 1.5,8.80039 1.5,9.90039 -0.1,3.6 -2.50039,-0.80001 -3.90039,-7 -1.2,-5.29999 -1.4,-5.49961 -3,-4.09961 -1,0.9 -1.99883,1.59961 -2.29883,1.59961 -0.3,0 -2.70078,-4.29961 -5.30078,-9.59961 -4.6,-9.19999 -13.09922,-20.1 -17.69922,-22.5 -1.8,-1 -1.80039,-0.8 0.0996,3 1.2,2.2 2.09961,4.4 2.09961,5 0,0.5 -2.70039,4.9004 -5.90039,9.90039 -3.3,5 -6.69922,11.49961 -7.69922,14.59961 -1,3.1 -2.00078,5.59961 -2.30078,5.59961 -0.3,0 -2.29883,-3.10078 -4.29883,-6.80078 L 391,838.5 l 0.0996,9.80078 C 391.19961,858.50077 390.9,859.7 389,858.5 c -2.6,-1.6 -5,-9.00079 -5,-15.80078 0,-5.9 -0.20039,-6.69922 -1.90039,-6.69922 -1.1,0 -2.99883,-1.3 -4.29883,-3 l -2.30078,-2.90039 0.40039,9 c 0.4,10.89999 -0.5004,11.20116 -6.40039,2.20117 -6.89999,-10.39999 -6.70039,-10.10078 -5.90039,-7.80078 0.9,2.9 5.40039,20.89922 5.40039,21.69922 0,2.3 -2.19961,7.7e-4 -7.09961,-7.19922 L 356.5,839.90039 356.19922,845 c -0.2,2.7 -0.6,5 -1,5 -0.4,0 -3.19922,-3.69922 -6.19922,-8.19922 -3,-4.49999 -7.7,-10.70117 -10.5,-13.70117 C 333.8,822.99961 325.79922,817 323.69922,817 c -0.5,0 2.4,3.50079 6.5,7.80078 C 342.49921,837.70077 348,847.30001 348,856 v 4.09961 l -5.80078,-3.90039 c -6.1,-4.2 -12.29961,-7.49922 -16.09961,-8.69922 -2,-0.6 -1.1996,0.7004 3.90039,6.90039 3.5,4.2 6.10078,7.79961 5.80078,8.09961 -0.2,0.3 -3.00117,-1.79922 -6.20117,-4.69922 l -5.79883,-5.10156 0.69922,6.5 c 0.8,7.59999 0.3996,8.10117 -7.90039,9.20117 -2.7,0.3 -8.39922,1.69961 -12.69922,3.09961 -20.89998,6.59999 -70.20091,9.1 -195.80078,10 -39.299961,0.3 -71.700391,0.7 -71.900391,1 -2.899997,2.9 7.001876,2.89922 711.701171,3.19922 644.29931,0.3 706.59881,0.20078 708.79881,-1.19922 1.3,-0.9 2.3016,-2.09961 2.1016,-2.59961 -0.2,-0.7 -16.1008,-1.00039 -46.3008,-0.90039 -72.4999,0.1 -132.0008,-2.09961 -164.3008,-6.09961 -18,-2.2 -20.8992,-2.90118 -49.6992,-11.70117 -11,-3.3 -23.2992,-6.59922 -27.1992,-7.19922 -4,-0.7 -7.3008,-1.6 -7.3008,-2 0,-0.5 0.4,-2 1,-3.5 l 1,-2.59961 -3.8008,2.90039 -3.6992,3 -21,0.59961 c -11.6,0.4 -21.4008,0.4 -21.8008,0 -0.4,-0.5 3,-4.49961 7.5,-9.09961 4.5,-4.59999 7.9016,-8.30078 7.6016,-8.30078 -1,0 -10.6008,5.69961 -14.3008,8.59961 -4.6,3.5 -5.7,3.20039 -5,-1.59961 0.3,-2.2 0.3,-4 0,-4 -0.3,0 -2.3996,2.79922 -4.5996,6.19922 -4.2,6.49999 -5.8008,7.50117 -12.3008,7.70117 -3.7,0.1 -3.7992,-10e-4 -3.1992,-2.70117 1.1,-4.3 8.4,-19.09923 12,-24.19922 5.2,-7.29999 5.2988,-7.69961 0.7988,-4.09961 -6.3,4.9 -11.9988,11.89884 -18.2988,22.29883 l -5.5996,9.30078 -5.2012,0.30078 c -2.8,0.2 -5.0996,-0.10156 -5.0996,-0.60156 0,-0.6 1.8,-3.09883 4,-5.79883 2.2,-2.6 3.7996,-5.00117 3.5996,-5.20117 -1,-0.9 -6.7,3.20039 -10.5,7.40039 L 1063,855.30078 v -2.90039 C 1063,843.8004 1069.2992,830 1073.1992,830 c 1.1,0 1.7004,-0.40039 1.4004,-0.90039 -0.7,-1.1 6.7,-7.99922 12.5,-11.69922 2.4,-1.5 6.9004,-3.8 9.9004,-5 6.9,-2.9 3.7,-3.1 -6,-0.5 -18.2,5.1 -25.5008,12.29963 -36.3008,36.09961 -2.1,4.6 -3.9992,8.60078 -4.1992,8.80078 -0.2,0.3 -3.8,-1.80156 -8,-4.60156 l -7.6992,-5.09961 -5.3008,2.59961 c -3,1.5 -5.7,2.40156 -6,2.10156 -0.3,-0.2 -1.0992,-5.30157 -1.6992,-11.10156 -0.7,-5.9 -1.5004,-10.69922 -1.9004,-10.69922 -0.4,0 -1.2008,4.59922 -1.8008,10.19922 C 1016.2996,855.4992 1016.5,855 1012.5,855 c -1.9,0 -3.5,-0.40078 -3.5,-0.80078 0,-0.5 1.1,-3.69922 2.5,-7.19922 l 2.5,-6.30078 -2.8008,2.40039 c -1.5,1.3 -4.4,3.90039 -6.5,5.90039 l -3.7988,3.5 0.6992,-4.5 c 0.3,-2.5 1.7004,-8.80001 2.9004,-14 1.3,-5.19999 2.6,-12.7 3,-16.5 0.4,-3.9 0.8992,-7.9 1.1992,-9 0.3,-1.3 0.2,-1.7 -0.5,-1 -0.5,0.5 -2.8992,6.60001 -5.1992,13.5 -5.69999,16.89998 -13.40039,31.90039 -15.40039,29.90039 -0.3,-0.3 0.10078,-3 0.80078,-6 0.8,-3 1.39883,-5.49961 1.29883,-5.59961 -0.1,-0.1 -2.59961,2.99883 -5.59961,6.79883 -3,3.8 -5.59922,6.79961 -5.69922,6.59961 -0.2,-0.2 -1.50039,-4.29962 -2.90039,-9.09961 -2.1,-7.39999 -8.49961,-21.9996 -7.09961,-16.09961 1,3.9 3.59961,16.59961 3.59961,17.59961 0,0.6 -1.19961,2.10039 -2.59961,3.40039 l -2.70117,2.5 -4.79883,-3.5 c -9.49999,-6.99999 -26.9004,-13.39961 -41.40039,-15.09961 -11.49999,-1.4 -33.50001,-10.40001 -47.5,-19.5 -9.19999,-5.89999 -25.7004,-22.79962 -30.90039,-31.59961 C 834.09962,767.80079 830,749.30076 830,726.30078 830,667.90084 847.10081,609.40035 875.30078,571.40039 885.60077,557.6004 901.80001,541.40077 914,532.80078 c 19.09998,-13.49999 44.80002,-24.5 66,-28.5 13.39999,-2.4 39.5004,-2.40039 50.9004,0.0996 9.6,2.1 19.8004,5.99883 27.9004,10.79883 8.9,5.19999 22.6992,19.6008 34.1992,35.80078 12.9,18.09998 22.8008,28.0004 33.8008,33.90039 4.5,2.4 10.1988,6.4 12.7988,9 4.9,4.8 11.4004,13.79922 11.4004,15.69922 0,0.8 -3,1.50039 -8,1.90039 l -8,0.69922 v 4.20117 c 0,2.2 -0.9,7.8004 -2,12.40039 -3.3,13.89999 -2.7992,14.79922 6.3008,10.19922 3.4,-1.8 8.8992,-4 12.1992,-5 3.3,-0.9 6.6992,-2.00039 7.6992,-2.40039 1.4,-0.6 1.7,0.0996 2,5.09961 0.3,5.79999 0.3004,5.90118 -5.0996,12.20117 l -5.4004,6.40039 3.9004,2 c 2.2,1.1 4.1008,2.09883 4.3008,2.29883 0.2,0.2 -0.7004,3.90039 -1.9004,8.40039 -1.3,4.4 -2.1008,8.2 -1.8008,8.5 0.3,0.3 3.8008,0.10039 7.8008,-0.59961 6.9,-1 7.4004,-0.99961 11.9004,1.40039 5.3,2.9 14.7988,5.09922 18.2988,4.19922 2,-0.5 2.9012,0.0996 5.7012,4.09961 1.9,2.5 5.4004,5.9 7.9004,7.5 4.7,3 6.7992,3.70039 5.6992,1.90039 -0.3,-0.5 -0.3004,-3.2 0.1,-6 l 0.7012,-5 h 8.1992 c 4.5,0 9.6008,-0.49961 11.3008,-1.09961 l 3.1992,-1.09961 -1.8008,-3 C 1230.4992,672.20079 1219.3008,667 1209.3008,667 h -5.4004 l 0.2988,-4.69922 0.3008,-4.70117 L 1210,656 c 7.7,-2.3 11.2,-4.29961 15,-8.59961 l 3.1992,-3.70117 -2.2988,-1.79883 c -1.3,-1 -5.5996,-3.69961 -9.5996,-6.09961 -6.8,-3.99999 -7.2008,-4.4 -7.8008,-8.5 -0.3,-2.4 -2.1996,-9.90001 -4.0996,-16.5 -1.9,-6.69999 -3.4004,-13.30117 -3.4004,-14.70117 0.1,-2.5 0.2996,-2.39961 4.0996,1.90039 2.2,2.5 6.4008,9.40001 9.3008,15.5 2.9,5.99999 9,16.50079 13.5,23.30078 l 8.2988,12.29883 1,10.20117 c 0.5,5.6 1.5016,12.39922 2.1016,15.19922 1.1,4.6 1.5,5.00078 4.5,5.30078 3.1,0.3 3.1992,0.19961 3.1992,-3.40039 0,-2.4 -1.4004,-6.10079 -3.9004,-10.80078 -3.7,-6.69999 -6.4,-14.00039 -5.5,-14.90039 0.2,-0.2 2.2004,1.30039 4.4004,3.40039 8,7.59999 12.0008,9.20117 24.3008,9.70117 l 11,0.5 -2.4004,-3.40039 c -2.3,-3.2 -2.4012,-4.4008 -2.7012,-19.30078 l -0.2988,-15.90039 -12.2012,0.60156 c -6.7,0.2 -13.2996,0.89844 -14.5996,1.39844 -2.3,0.9 -2.3992,1.30157 -2.1992,9.10156 0.1,5.3 -0.2008,8.19922 -0.8008,8.19922 -5.1,0 -19.3992,-22.99923 -21.1992,-34.19922 -0.6,-3.5 -1.3,-7.3 -1.5,-8.5 -0.2,-1.1 -3.8,-7.20001 -8,-13.5 -7.8,-11.39999 -17.1996,-29.30039 -16.0996,-30.40039 0.6,-0.6 18.1988,10.89923 25.7988,16.69922 2,1.6 7.6004,8.0004 12.4004,14.40039 10.4,13.79999 17.5,20.39922 27,25.19922 8,4.09999 18.3004,7.50156 19.9004,6.60156 0.6,-0.3 0.8004,-0.30078 0.4004,0.19922 -0.5,0.5 0.9,1.90078 3,3.30078 13.2,8.79999 16.7996,12.10001 19.5996,18 l 2.9004,6 2.0996,-2.5 2.1992,-2.60156 8.5,4.90039 c 9.6,5.59999 8.8996,6.29961 -3.9004,4.09961 -8.1,-1.4 -8.0984,-1.4 -10.3984,1 -1.2,1.3 -4.2016,3.60156 -6.6016,5.10156 -2.4,1.5 -5.5988,4.79961 -7.2988,7.59961 l -3,4.90039 2,4.79883 c 1.2,2.6 2.0996,5.40078 2.0996,6.30078 0,0.8 0.4,2.59961 1,4.09961 l 1,2.5 4.0996,-2.5 c 2.2,-1.5 4.2004,-2.50078 4.4004,-2.30078 0.1,0.2 1,2.4 2,5 l 1.6992,4.70117 3.6016,-2.90039 3.5996,-2.90039 4.0996,3.80078 c 4.3,4 8.3996,6.60039 9.0996,5.90039 0.2,-0.3 0.7,-3.30117 1,-6.70117 0.4,-3.4 1.0008,-6.3 1.3008,-6.5 0.4,-0.3 2.8004,0.10117 5.4004,0.70117 2.6,0.5 4.9984,0.69883 5.3984,0.29883 0.4,-0.4 1.0004,-3.50039 1.4004,-6.90039 l 0.5996,-6.19922 -5.5996,-5.19922 -5.5996,-5.20117 10.5996,0.59961 10.5996,0.60156 2.4004,-4.40039 c 1.9,-3.5 3.1004,-4.50117 5.4004,-4.70117 4.7,-0.5 4.8004,-1.49962 0.9004,-6.09961 -2,-2.3 -3.3996,-4.39922 -3.0996,-4.69922 1.1,-1.2 17.5992,8.79922 17.6992,10.69922 0,0.4 -2.0992,1.40039 -4.6992,2.40039 -2.7,1 -7.9012,3.29922 -11.7012,5.19922 l -6.9004,3.5 0.6016,4.60156 c 0.9,7.99999 2.0992,9.69923 11.1992,16.19922 4.8,3.4 8.8,6.00078 9,5.80078 0.2,-0.2 1.1992,-6.40079 2.1992,-13.80078 l 2,-13.5 2.7012,5.5 c 4.4,8.59999 5,9 12.5,9 3.6,0 9.6996,-0.7 13.5996,-1.5 3.8,-0.8 7.0996,-1.5 7.0996,-1.5 0.1,0 0.5008,1.10039 0.8008,2.40039 1,3.9 9.2,14.2 14,17.5 5.1,3.6 12.0992,4.19961 18.1992,1.59961 4,-1.7 3.9008,-2.19962 -2.1992,-10.59961 l -3.4004,-4.59961 12.8008,-5.10156 c 7,-2.8 13.9996,-5.59922 15.5996,-6.19922 l 2.9004,-1.09961 -2.4004,-2.5 c -3.4,-3.6 -9.0008,-6.3 -16.8008,-8 -3.6,-0.7 -6.8988,-1.40039 -7.2988,-1.40039 -0.5,0 -0.8008,-1.80039 -0.8008,-3.90039 0,-3.9 -0.1008,-3.89922 -2.8008,-3.19922 -4.4,1.3 -8.6988,3.69883 -16.2988,9.29883 l -7,5.10156 -3.0996,-2.90039 c -2.6,-2.5 -11.8008,-15.40117 -11.8008,-16.70117 0,-0.8 11.8004,-9.69922 12.9004,-9.69922 0.5,0 2.1992,1.50039 3.6992,3.40039 2.4,2.9 3.0996,3.3 5.0996,2.5 1.4,-0.5 5.7016,-0.90039 9.6016,-0.90039 9.4,0 10.6988,-0.80039 7.7988,-4.90039 L 1439.8008,633 1443,629.30078 l 3.3008,-3.70117 -2.2012,-1.69922 c -1.1,-0.9 -5.3996,-2.40039 -9.5996,-3.40039 l -7.5,-1.59961 -2,-6.70117 c -1.2,-3.7 -2.3992,-8.8 -2.6992,-11.5 L 1421.6992,596 h 4.1016 c 6.8,0 7.3996,-0.60079 5.5996,-4.80078 -1.8,-4.3 -6.1004,-8.99922 -12.4004,-13.69922 l -4.5,-3.19922 -8.5,1.59961 c -8.3,1.6 -8.5008,1.8 -10.3008,5.5 -3.1,6.39999 -2.0984,8.69961 5.1016,12.09961 L 1407,596.40039 V 607 h -16.4004 c -18.2,0 -29.7996,-1.6 -40.0996,-5.5 -15.4,-5.79999 -36.7004,-24.10002 -48.4004,-41.5 -3.4,-5 -6.0996,-9.59961 -6.0996,-10.09961 0,-2.3 16.0996,0.59961 23.0996,4.09961 2.7,1.4 10.6,8.5004 20,17.90039 15.5,15.69998 15.6012,15.69962 28.2012,22.09961 10.7,5.39999 13.5984,6.50078 18.8984,6.80078 6.3,0.4 12.8008,-1.00156 9.8008,-2.10156 -17.4,-6.39999 -24.5004,-9.39961 -30.4004,-13.09961 -7.5,-4.7 -11.3988,-8.19922 -10.2988,-9.19922 1,-1 15.8992,-0.99961 28.6992,-0.0996 l 11.5,0.89844 4.6992,-3.19922 4.8008,-3.19922 -4.8008,-5.60156 c -5.9,-7.2 -5.3992,-6.79923 -12.6992,-8.19923 -6.2,-1.2 -9.8,-2.79961 -8,-3.59961 3,-1.3 11.7004,-7.40039 9.9004,-6.90039 -6.8,1.8 -18.6,3.5 -25,3.5 -7.2,0 -7.4004,-0.1 -6.4004,-2 0.9,-1.7 2.1008,-2 7.3008,-2 8,0 14.5988,-3.09962 20.7988,-9.59961 7.7,-8.39999 7.8996,-7.90039 -1.9004,-7.90039 -8.2,0 -9,-0.19922 -13.5,-3.19922 L 1365.9004,522 l -3.7012,1.90039 c -2,1.1 -4.9992,3.29961 -6.6992,5.09961 -2.6,2.7 -3,3.90079 -3,8.80078 0,3.1 0.4008,7.4 0.8008,9.5 l 0.7988,3.69922 -4.2988,-0.5 c -24.2,-3.3 -38.9008,-7.40079 -53.3008,-14.80078 -10.1,-5.2 -14.4,-9.69962 -22.5,-23.09961 -6,-9.89999 -10.6004,-15.2004 -19.9004,-22.90039 -1.8,-1.5 -3.2,-2.9 -3,-3 0.2,-0.2 5.8008,-1.29961 12.3008,-2.59961 9.5,-1.8 12.7,-2.8 15.5,-5 3.9,-3 11.3996,-7.09961 13.0996,-7.09961 0.5,0 1,2 1,4.5 v 4.5 h -11.9004 L 1280,484.69922 c -0.6,2.1 -1.3992,6.6 -1.6992,10 -1,10.49999 1.8984,14.40039 13.3984,18.40039 5,1.7 6.4012,1.80078 10.7012,0.80078 3.2,-0.8 8.4988,-0.9 15.2988,-0.5 5.6,0.4 10.3008,0.29883 10.3008,-0.20117 0,-0.4 -2,-4.09961 -4.5,-8.09961 -2.5,-4.1 -4.4008,-7.79883 -4.3008,-8.29883 0.3,-0.7 16.3004,-7.50078 18.4004,-7.80078 0.1,0 -0.4988,1.89922 -1.2988,4.19922 -3.1,8.89999 0.9992,18.90117 8.1992,20.20117 2,0.4 4.2996,0.99883 5.0996,1.29883 1.1,0.4 1.4004,-0.89962 1.4004,-7.09961 V 500 l 5.1992,2.59961 c 2.9,1.5 8.4008,4.80039 12.3008,7.40039 11.1,7.49999 17.5996,10.80078 28.0996,14.30078 l 9.7012,3.29883 1.2988,4.40039 c 1.5,5 1.3012,5.29922 -4.7988,7.19922 -4.3,1.3 -4.3008,1.40157 -2.8008,12.10156 1.6,10.89999 1.8992,11.29922 10.1992,13.69922 4,1.2 8.7,3.29961 10.5,4.59961 4.2,3.2 7.0004,3.00117 14.9004,-0.79883 l 6.5,-3.30078 -0.2988,-5 -0.3008,-4.90039 7,-1.59961 c 5.8,-1.3 7.2,-2.00078 8.5,-4.30078 1.4,-2.7 1.4008,-2.89961 -0.6992,-4.59961 -3.2,-2.7 -1.8,-3.09961 3.5,-1.09961 4.2,1.6 4.7996,2.19961 5.5996,5.59961 1.4,6.39999 4.5,16.80078 5.5,18.30078 1.2,1.9 4.5992,0.60039 9.1992,-3.59961 l 3.9004,-3.5 4.9004,2.09961 c 7.8,3.3 7.9988,3.20038 8.2988,-5.59961 l 0.3008,-7.70117 5,0.20117 c 2.8,0.2 5.9004,0.59844 6.9004,0.89844 1.7,0.5 1.8004,0.10156 1.4004,-3.89844 C 1510.1008,540.50079 1499.6,525 1496,525 c -4.4,0 -5.3992,-1.3004 -4.6992,-6.40039 0.4,-2.5 0.4984,-4.59961 0.3984,-4.59961 -0.2,0 -3.1,1.1 -6.5,2.5 -6.2,2.5 -12.5996,3.29961 -13.5996,1.59961 -0.3,-0.5 0.8004,-4.79922 2.4004,-9.69922 1.7,-4.79999 3,-8.89961 3,-9.09961 0,-0.3 -2.7004,0.99922 -5.9004,2.69922 -7.3,3.9 -9,7.19961 -6,11.59961 1.8,2.8 2.6008,10.30117 1.3008,11.70117 -0.4,0.3 -2.1008,-0.40117 -3.8008,-1.70117 -3.1,-2.4 -3.1996,-2.40039 -10.0996,-0.90039 -7.4,1.6 -20.2008,1.30078 -31.3008,-0.69922 -3.2,-0.5 -8.3992,-1.4 -11.6992,-2 -9,-1.5 -20.4,-7.70001 -36,-19.5 -7.7,-5.89999 -15.4992,-11.49922 -17.1992,-12.69922 -1.8,-1.1 -3.3008,-2.30117 -3.3008,-2.70117 0,-1.4 19.1992,-7.29883 25.1992,-7.79883 4.9,-0.3 8.2012,0.0996 14.7012,2.09961 4.7,1.4 9.8988,2.59961 11.7988,2.59961 3.1,0 3.3008,-0.20039 3.3008,-3.40039 0,-4.2 1.6996,-5.59922 2.5996,-2.19922 0.8,3.3 -0.2988,6.49883 -3.2988,9.79883 -2.8,3 -2.6012,4.70078 1.2988,9.30078 1.5,1.6 2.3008,2.3 1.8008,1.5 -0.4,-0.8 -0.3004,-1.19922 0.1,-0.69922 2,1.7 1.1004,5.09844 -2.0996,7.89844 l -3.3008,2.90039 5.7012,4.80078 c 3.1,2.6 6.2996,5.50039 7.0996,6.40039 1.9,2.4 2.7988,2.1 3.7988,-1 0.9,-2.6 1.0004,-2.70078 10.9004,-2.80078 8.9,-0.1 10.4996,-0.39922 14.5996,-2.69922 4.7,-2.8 18.5,-16.80039 21.5,-21.90039 1,-1.6 2.0008,-2.90039 2.3008,-2.90039 0.3,0 1.9004,2.59922 3.4004,5.69922 l 2.9004,5.80078 1.3984,-3.40039 c 1.2,-2.9 1.7016,-3.19922 4.1016,-2.69922 9.1,2.1 16.2,2.8 20.5,2 l 4.5996,-0.90039 -2.7012,-4.09961 C 1500.8992,483.1004 1496.8,478.6 1493,476 c -1.9,-1.3 -3.7004,-2.49922 -3.9004,-2.69922 -0.2,-0.1 0.4,-1.30039 1.5,-2.40039 1.9,-2.2 1.9004,-2.20039 -0.5996,-1.40039 -7.4,2.1 -13.5992,5.89962 -19.6992,12.09961 -6.5,6.59999 -6.6,6.6 -11,6 -6.2,-0.8 -8.3008,-2.20039 -8.3008,-5.40039 0,-4.3 -2.1996,-6.59844 -5.5996,-5.89844 -1.6,0.3 -3.6996,0.89922 -4.5996,1.19922 -1.5,0.6 -1.8008,0.10039 -1.8008,-2.59961 0,-1.9 0.2996,-4.40117 0.5996,-5.70117 0.7,-2.3 0.7008,-2.29883 2.8008,-0.29883 2.6,2.4 11.3988,7.09961 13.2988,7.09961 1.3,0 1.3008,-0.3004 -0.1992,-7.40039 -0.5,-2.6 -0.4008,-2.79922 1.6992,-2.19922 1.3,0.3 5.2008,0.89922 8.8008,1.19922 l 6.5,0.59961 9.5,-9.09961 9.5,-9.09961 -3.9004,-3.40039 -4,-3.40039 1.5996,-5.59961 c 0.8,-3 1.3,-5.79961 1,-6.09961 -0.9,-0.8 -7.4996,-0.6 -11.0996,0.5 -1.9,0.5 -3.5988,0.79961 -3.7988,0.59961 -0.2,-0.1 0.3992,-2.19961 1.1992,-4.59961 0.8,-2.3 1.5,-4.39961 1.5,-4.59961 0,-0.1 -2.9992,0.59922 -6.6992,1.69922 -3.8,1.1 -8,2.19961 -9.5,2.59961 l -2.8008,0.5 1.5,-4.29883 c 3.1,-8.79999 6.8004,-15.00078 8.4004,-14.30078 3.4,1.5 12,2.7 14.5,2 2.6,-0.6 2.5996,-0.70039 1.0996,-4.40039 -1.8,-4.4 -6.7,-9.19922 -9.5,-9.19922 -1.1,0 -3.7008,1.6 -5.8008,3.5 -2.3,2 -4.2988,3.09961 -4.7988,2.59961 -0.5,-0.5 -5.6004,-3.49961 -11.4004,-6.59961 -5.8,-3.2 -11.2992,-6.39922 -12.1992,-7.19922 -1,-0.9 -1.9008,-3.70117 -2.3008,-7.20117 -0.3,-3.1 -1,-6.59922 -1.5,-7.69922 -1.1,-2.7 -9.3992,-6.10117 -11.1992,-4.70117 -2.6,2 -30.7016,16.30078 -32.1016,16.30078 -0.8,0 -5.8992,-2 -11.1992,-4.5 -5.4,-2.5 -11.7992,-4.79922 -14.1992,-5.19922 L 1360,376.69922 v -4.29883 c 0,-2.4 -0.2992,-4.40039 -0.6992,-4.40039 -0.5,0 -4.2,3.2 -8.5,7 -5.7,5.19999 -8.4,7 -10.5,7 -2.5,0 -3.2,0.80079 -6.5,7.80078 -2,4.2 -4.5004,8.7 -5.4004,10 -1.7,2.3 -1.7012,2.3 0.7988,2 2,-0.2 3.6016,0.89844 8.1016,5.89844 4.8,5.29999 5.5992,6.9 6.1992,11.5 1.1,8.99999 1.0992,9.00117 4.6992,7.20117 4.1,-2.1 4.6004,-1.10039 1.4004,3.09961 -5.3,6.89999 -27.1997,16.09924 -81.5996,34.19922 -22.3,7.39999 -46.6,15.00078 -54,16.80078 -13.1,3.3 -33.1004,6.70039 -33.9004,5.90039 -1.2,-1.2 20.6004,-21.00118 29.9004,-27.20117 3.2,-2.2 10.8004,-5.59922 16.9004,-7.69922 6.1,-2.1 10.8,-4.19961 10.5,-4.59961 -0.3,-0.5 -4.5,-0.90039 -9.5,-0.90039 -4.9,0 -8.9004,-0.40039 -8.9004,-0.90039 0,-0.5 1.1004,-1.79883 2.4004,-2.79883 1.3,-1 5.8,-4.70156 10,-8.10156 l 7.6992,-6.39844 4.5,4.19922 c 6.1,5.59999 9.9008,7 19.3008,7 7.3,0 7.9,-0.20078 8.5,-2.30078 1,-3.9 0.6988,-6.8 -1.2012,-9.5 -1,-1.5 -3.2992,-5.39922 -5.1992,-8.69922 -1.8,-3.3 -3.8992,-6.99961 -4.6992,-8.09961 -1.2,-1.9 -1.1004,-2.30078 1.0996,-3.80078 2.3,-1.4 3.3988,-1.39961 12.7988,-0.0996 5.7,0.8 11.9008,1.9 13.8008,2.5 6.6,1.9 7.8,1.49921 8.5,-3.30078 1.1,-6.99999 1.4,-7.49883 6.5,-9.29883 4.6,-1.6 5,-2.10078 6,-6.30078 1.1,-4.5 1.3,-4.7 8,-7.5 10.4,-4.3 10.5,-4.59962 0.5,-10.09961 -8.1,-4.4 -15.5,-10.4 -15.5,-12.5 0,-0.6 -2.2,-1 -5,-1 -12.5,0 1.2008,-6.4 19.3008,-9 4.3,-0.7 9.3984,-1.70039 11.3984,-2.40039 2.9,-1 3.3,-1.40039 2,-2.40039 -1.1,-0.9 -3.7996,-0.89961 -12.0996,-0.0996 -5.9,0.6 -10.9988,0.9 -11.2988,0.5 -0.4,-0.3 0.4988,-1.90039 1.7988,-3.40039 5.5,-6 5.6008,-5.60001 -0.6992,-11.5 l -5.7012,-5.29883 4.7012,-6.40039 4.5996,-6.30078 -2.3008,-3.09961 c -2,-2.5 -3.7988,-3.39883 -9.7988,-4.79883 -10.4,-2.4 -13.5012,-2.20078 -20.7012,1.19922 -7.5,3.5 -7.4988,3.49922 -6.7988,1.19922 0.3,-0.9 0.9996,-3.6 1.5996,-6 L 1274.0996,297 h 5 c 2.8,0 11.1,0.89961 18.5,2.09961 7.4,1.1 15.3996,1.90117 17.5996,1.70117 3.7,-0.3 2.8,-0.70117 -8,-3.70117 -6.8,-1.9 -12.4992,-3.69922 -12.6992,-4.19922 -0.7,-1.1 7.9,-2.40078 15.5,-2.30078 5.6,0 7.8996,0.69961 16.0996,4.59961 l 9.5996,4.5 -1.5,8.60156 -1.5,8.59961 -6.6992,-0.59961 c -6.2,-0.5 -6.8004,-0.40117 -8.4004,1.79883 -1,1.5 -1.5996,4.4 -1.5996,8 0,5.69999 -4e-4,5.80039 3.5996,6.90039 4.9,1.7 14.8008,0.79961 17.3008,-1.40039 1,-0.9 3.2996,-4.39922 5.0996,-7.69922 l 3.0996,-5.90039 3,1.69922 c 1.7,1 4.2,3.3 5.5,5 2,2.7 3.2012,3.30078 6.2012,3.30078 2,0.1 5.2992,1 7.1992,2 6.6,3.6 10.4008,5.19922 13.8008,5.69922 l 3.2988,0.60156 -0.5996,-3.90039 c -0.4,-2.1 -1.2996,-6.19961 -2.0996,-9.09961 -2,-7.19999 -1.9,-7.30078 3,-7.30078 8.2,0 8.4,-0.69962 2,-7.09961 -5.5,-5.59999 -5.7,-5.90039 -5,-9.90039 l 0.5996,-4.09961 -5.5,-2.5 c -4.8,-2.2 -5.5,-2.90039 -5.5,-5.40039 0,-1.6 -0.6996,-3.00039 -1.5996,-3.40039 -2.9,-1.1 -13.0012,-0.59961 -18.2012,0.90039 L 1346,286.09961 v 11.20117 l -3.9004,-0.60156 c -4.5,-0.6 -8.5004,-2.39962 -20.4004,-9.09961 L 1313,282.69922 1322.6992,279 c 7.7,-2.9 10.3004,-4.30078 11.9004,-6.80078 1.9,-2.8 3.9996,-10.89844 3.0996,-11.89844 -0.3,-0.2 -3.1996,1.89922 -6.5996,4.69922 l -6.0996,5.19922 -3.6992,-1.89844 c -4.5,-2.2 -7.4012,-6.90118 -7.7012,-12.20117 l -0.1,-3.90039 L 1311,254 c -7.6,5.39999 -10.7992,11.59923 -11.6992,22.19922 -0.7,7.79999 0.5988,6.90078 -13.7012,9.30078 -13,2.2 -13.0996,2.19921 -16.0996,-2.80078 -1.4,-2.4 -2.5,-4.79844 -2.5,-5.39844 0,-0.6 2.0996,-1.60117 4.5996,-2.20117 10,-2.5 17.8012,-8.99884 22.7012,-18.79883 4.1,-8.09999 4.2992,-9.30078 1.1992,-9.30078 -3.4,0 -8.3996,-2.19961 -14.0996,-6.09961 -4.2,-2.9 -4.5,-3.40039 -4,-6.40039 l 0.5996,-3.30078 -6.6992,2.5 c -3.7,1.3 -9.7008,3.10039 -13.3008,3.90039 -3.6,0.8 -7.4992,1.70039 -8.6992,1.90039 -3,0.7 -3.0008,5.00001 -0.3008,16 1.1,4.4 2,9.1 2,10.5 0,1.4 0.4996,4.20078 1.0996,6.30078 0.9,3.2 1.9012,4.19922 7.7012,7.19922 6.7,3.4 10.1992,5.79922 10.1992,6.69922 0,1 -8.5,2.90117 -23,5.20117 -8,1.3 -14.6008,2.29961 -14.8008,2.09961 -0.1,-0.1 1.9016,-4.49922 4.6016,-9.69922 4.1,-7.89999 9.8984,-24.50039 8.8984,-25.40039 -0.2,-0.2 -1.9996,3.2004 -4.0996,7.40039 -10.3,20.49998 -14.3004,25.50001 -28.9004,36.5 -6.1,4.7 -14.7992,11.79883 -19.1992,15.79883 -15,13.79999 -25.9996,21.60118 -41.0996,29.20117 -9.8,4.9 -27.1,11.29844 -28,10.39844 -0.9,-0.8 6.1992,-13.59962 11.6992,-21.09961 8.1,-10.89999 22.1004,-18.69883 44.9004,-24.79883 L 1191.5,309 l -7.6992,0.59961 -7.8008,0.59961 -0.1992,-4.39844 c -0.1,-3.7 -0.2,-3.8 -0.5,-1 -0.5,3.9 -3.6004,5.79883 -11.4004,6.79883 -6.2,0.8 -6.2,0.80116 1.5,-7.29883 l 4,-4.30078 -2.2012,-2.30078 c -1.2,-1.3 -2.1992,-2.59844 -2.1992,-2.89844 0,-0.3 2.8008,-1.10078 6.3008,-1.80078 3.4,-0.7 6.4988,-1.69922 6.7988,-2.19922 0.3,-0.5 2.2012,0.2 4.2012,1.5 5.3,3.6 12.6988,5.09883 21.2988,4.29883 L 1211,296 l 6.9004,-7.19922 7,-7.30078 -4.0996,-1.80078 -4.1016,-1.69922 0.7012,-9.80078 c 0.7,-10.89999 0.9988,-11.29883 9.7988,-15.79883 9.6,-4.89999 15.7008,-12.60001 15.8008,-20 0,-2.8 -0.4,-3.09961 -4.5,-4.09961 -4.8,-1.1 -19.4996,-10.20078 -19.0996,-11.80078 0.1,-0.5 1.9992,-1.29961 4.1992,-1.59961 3.5,-0.7 8.9996,-2.40078 13.5996,-4.30078 1.2,-0.5 1.8008,-0.10039 2.3008,1.59961 2.2,6.79999 4.7996,12.80078 5.5996,12.80078 0.6,0 0.7008,-0.60078 0.3008,-1.30078 -0.4,-0.7 -0.3008,-0.89844 0.1992,-0.39844 0.5,0.5 0.9004,1.4 0.9004,2 0,0.8 5.3008,1.89961 14.8008,3.09961 l 14.8984,1.90039 2.1016,-2.20117 L 1280.4004,226 1276,221.40039 c -8.4,-8.79999 -11.4,-11.60078 -12.5,-11.80078 -1.8,-0.3 -3.5,-1.69961 -3,-2.59961 0.3,-0.4 -1.2004,-2.20039 -3.4004,-3.90039 -5.9,-4.7 -9.6992,-4.29999 -19.1992,2 l -7.8008,5.09961 -10.7988,-1.59961 c -5.9,-0.9 -13.6,-2.09922 -17,-2.69922 l -6.2012,-1.09961 1.5,-3.70117 c 0.9,-2 1.7996,-4.20039 2.0996,-4.90039 0.4,-0.8 3.4012,-1.19922 10.2012,-1.19922 11.6,0 16.4004,-1.89961 16.9004,-6.59961 0.2,-1.6 1.5988,-8.39962 3.2988,-15.09961 1.6,-6.69999 2.9004,-13.30117 2.9004,-14.70117 v -2.5 l -8.5,3 c -4.7,1.6 -9.1004,2.90039 -9.9004,2.90039 -0.8,0 -2.6992,-3.60001 -4.6992,-9 -3,-7.79999 -4.9004,-10.99961 -4.9004,-8.09961 0,0.5 -4,4.3 -9,8.5 -4.9,4.3 -10.1,9.7 -11.5,12 l -2.5,4.29883 2.8008,10.90039 c 1.7,6.39999 3.9,12.50078 5.5,14.80078 3.4,4.9 3.3996,7.09961 0.1,10.09961 -4.1,3.8 -26.4004,13.69922 -26.4004,11.69922 0,-0.4 2.4996,-2.19961 5.5996,-4.09961 5.4,-3.2 5.5,-3.29922 4,-5.69922 -3.4,-4.99999 -1.6004,-11.90118 5.0996,-20.20117 3.3,-4 3.4004,-4.39961 1.9004,-6.09961 -2.1,-2.3 -10.5004,-6.49922 -11.4004,-5.69922 -0.3,0.4 -0.8992,2.79883 -1.1992,5.29883 -0.5,4.59999 -0.5004,4.70039 -5.4004,5.90039 -4.4,1.1 -5.0992,1.09961 -6.1992,-0.40039 -0.8,-0.9 -1.4004,-1.89922 -1.4004,-2.19922 0,-0.3 -0.4004,-1.10039 -0.9004,-1.90039 -0.8,-1.2 -2.1,-0.49922 -6.5,3.30078 -5.7,4.8 -6.2992,4.89961 -10.1992,2.09961 -2,-1.5 -2.3008,-1.2996 -6.8008,3.90039 -5,5.89999 -5.1004,6.29884 -3.4004,14.79883 1.2,5.79999 1.2008,5.90078 -1.6992,8.80078 l -3,2.90039 -5.1992,-3.20117 c -3.9,-2.4 -5.0008,-3.59922 -4.3008,-4.69922 0.7,-1.3 5,-13.79961 5,-14.59961 0,-0.1 -1.8,-0.50039 -4,-0.90039 -2.2,-0.4 -4.2004,-1.00039 -4.4004,-1.40039 -0.9,-1.3 23.1004,-36.09961 24.9004,-36.09961 0.2,0 1.5996,1.1 3.0996,2.5 5.8,5.39999 17.3012,9 24.2012,7.5 2.6,-0.6 5.0984,-2.4 8.3984,-6 l 4.6016,-5.19922 -1.6016,-3.60156 c -0.9,-2.2 -1.5996,-6.69844 -1.5996,-10.89844 l -0.1,-7.10156 -3.4004,1.40039 c -4.7,2 -4.9996,1.79961 -4.5996,-1.90039 0.3,-1.7 0.6008,-5.29844 0.8008,-7.89844 0.4,-4.59999 0.2984,-4.80039 -3.1016,-5.90039 -1.9,-0.6 -6.9984,-4.50078 -11.3984,-8.80078 l -8,-7.59961 -2.3008,2.69922 c -1.3,1.5 -2.8992,3.5 -3.6992,4.5 -0.7,1 -3.4,3.60117 -6,5.70117 -2.6,2.1 -4.5012,4.19922 -4.2012,4.69922 0.3,0.5 -0.2992,0.60117 -1.1992,0.20117 -1.3,-0.5 -1.5004,-0.30078 -0.9004,0.69922 0.5,0.8 0.4008,1.09922 -0.1992,0.69922 -0.6,-0.4 -2.1008,0.10156 -3.3008,1.10156 -2.2,1.8 -2.2,1.89883 -0.5,5.29883 1,1.9 3.1996,5.49961 5.0996,8.09961 l 3.3008,4.70117 -5.0996,1.69922 -5.2012,1.70117 -2.9004,-3.40039 c -1.5,-1.9 -3.2988,-3.29961 -3.7988,-3.09961 -0.6,0.1 -0.8,-0.10156 -0.5,-0.60156 0.3,-0.5 -0.1996,-2.19883 -1.0996,-3.79883 -0.8,-1.6 -2.0012,-4.3 -2.7012,-6 l -1.0996,-3.20117 -7,5.70117 c -5.6,4.5 -7.8992,7.30001 -11.1992,13.5 -2.3,4.3 -5.4004,9.29883 -6.9004,11.29883 -2.6,3.4 -2.6008,3.5 -0.3008,2 1.3,-0.9 2.9996,-1.29844 3.5996,-0.89844 1.4,0.9 5.7,10.09883 5,10.79883 -0.6,0.7 -13.6984,0.7 -14.3984,0 -0.3,-0.2 0.2,-1.9 1,-3.5 l 1.5996,-3.09961 -4.2012,3.30078 c -2.8,2.2 -6.1992,3.79922 -10.1992,4.69922 -3.3,0.7 -11.2992,3.19961 -17.6992,5.59961 -11.3,4.2 -13.8008,4.90039 -13.8008,3.90039 0,-0.2 1.8,-2.79922 4,-5.69922 4.7,-6.19999 11,-17.80078 11,-20.30078 0,-1.2 0.5008,-1.5 1.8008,-1 7.1,3.1 10.1984,3.69922 16.3984,3.19922 9.7,-0.6 13.2004,-3.19923 15.9004,-11.69922 1.1,-3.6 3.2008,-8.3 4.8008,-10.5 1.5,-2.2 3.3004,-5.6 3.9004,-7.5 1.1,-3.6 2.6984,-20 1.8984,-20 -0.2,0 -2.9,0.90039 -6,1.90039 -3.1,1.1 -7.6,2.29961 -10,2.59961 l -4.3984,0.69922 -1.7012,-4.29883 c -0.9,-2.3 -2.0992,-6.501174 -2.6992,-9.201171 -0.5,-2.799997 -1.1008,-5.200001 -1.3008,-5.5 -1,-0.899999 -7.3996,2.000394 -11.5996,5.40039 -9.3,7.399991 -15.0004,13.600791 -18.9004,20.800781 -2.3,4 -4.0996,8.19883 -4.0996,9.29883 0,1 2.2996,4.10078 5.0996,6.80078 9.7,9.49999 10.0996,13.49962 2.5996,23.59961 -1.5,2 -1.8992,2.20039 -2.6992,0.90039 -0.5,-0.8 -1.1996,-3.79961 -1.5996,-6.59961 -0.6,-4.89999 -0.8004,-5.20078 -4.4004,-6.30078 -2.7,-0.8 -4.9004,-2.59961 -7.9004,-6.59961 -3.8,-5 -4.2992,-5.4 -6.6992,-4.5 -1.4,0.6 -2.9004,1 -3.4004,1 -0.4,0 -1.1,-3.30039 -1.5,-7.40039 l -0.6992,-7.40039 -5.7012,-1.59961 c -10.19998,-2.8 -12,-2.3996 -18.49999,4.40039 -5.6,5.89999 -5.8,6 -8.5,4.5 C 973.19961,126.1 972,122.89999 972,117 c 0,-3 0.30001,-3.2 9.5,-4.5 7.49999,-1.1 12.5,-2.79961 12.5,-4.09961 0,-0.8 -1.30039,-2.79961 -2.90039,-4.59961 -2.7,-3 -3.2004,-3.20078 -12.40039,-3.80078 -5.3,-0.3 -9.89922,-0.899219 -10.19922,-1.199219 -0.3,-0.299999 1.00039,-3.101175 2.90039,-6.201172 1.9,-3.199996 5.1,-10.498834 7,-16.298828 l 3.5,-10.5 L 976.5,66.199219 C 971.40001,66.599218 971,66.499998 971,64.5 c 0,-3.499997 -2.89922,-10.900783 -4.69922,-12.300781 -2.3,-1.699999 -6.40156,-1.499608 -9.10156,0.40039 -3.3,2.299998 -4.7,1.300776 -6,-4.699218 C 950.09922,43.100395 946.80039,38 944.90039,38 Z m 19.66211,14.150391 c 0.3625,0.224999 0.88711,0.748829 1.53711,1.548828 1.3,1.599998 1.20117,1.700389 -0.29883,0.40039 C 964.80078,53.39961 964,52.60039 964,52.400391 c 0,-0.4 0.2,-0.475 0.5625,-0.25 z M 948.40039,91 h 2.90039 l -0.60156,6.699219 c -0.7,7.699991 -5.60001,18.500011 -13,29.000001 -4.4,6.29999 -5.69922,6.90156 -5.69922,2.60156 0,-2 -0.5,-2.30078 -4,-2.30078 -5.39999,0 -5.09921,-1.5 0.80078,-3.5 2.6,-1 5.89844,-2.19922 7.39844,-2.69922 2,-0.7 3.60156,-3.10118 6.60156,-9.20117 3.9,-7.99999 6.79844,-16.39922 5.89844,-17.199219 -0.2,-0.3 -2.79922,0.49961 -5.69922,1.599609 -5.69999,2.199998 -7.10039,2.000388 -4.90039,-0.599609 C 939.99961,93.100393 944.80039,91.1 948.40039,91 Z m -88.80078,4.300781 c -0.3,-0.399999 -0.6,0.198829 -0.5,1.298828 0,1.099999 0.29961,1.399609 0.59961,0.59961 0.3,-0.7 0.20039,-1.598438 -0.0996,-1.898438 z m 96.53711,2.875 c 0.9375,-0.125 2.36289,-0.075 4.46289,0.125 3.1,0.3 5.8,0.8 6,1 0.9,0.899999 -1.49883,10.099609 -3.29883,12.599609 -1.7,2.4 -13.00039,8.80039 -13.90039,7.90039 -0.3,-0.4 3.59922,-16.20118 5.19922,-20.701171 0.15,-0.499999 0.59961,-0.798828 1.53711,-0.923828 z M 678.30078,115.09961 c 0.4,0 0.69922,4.2004 0.69922,9.40039 0,5.19999 -0.29922,9.5 -0.69922,9.5 -1.8,-0.1 -10.40117,-15.40078 -9.20117,-16.30078 0.4,-0.3 7.50117,-2.29961 9.20117,-2.59961 z m 292,1 c 0.4,-0.1 0.69922,0.70117 0.69922,1.70117 0,1 -1.70078,4.29883 -3.80078,7.29883 -3.2,4.59999 -5.10001,6.10118 -13,10.20117 L 945,140.09961 949,145.5 c 3.5,4.8 4.00039,5.89922 3.40039,9.19922 L 951.80078,158.5 956.5,159.80078 c 3.2,0.8 6.3,2.79883 10,6.29883 l 5.30078,5 2.29883,-4 c 3.2,-5.79999 4.3,-5.20038 3.5,2.09961 -1,8.69999 -1.69922,11.80078 -2.69922,11.80078 -0.4,0 -5.39961,-4.40079 -11.09961,-9.80078 -5.69999,-5.4 -11.80156,-10.69922 -13.60156,-11.69922 -2.7,-1.6 -3.29961,-2.6 -3.09961,-5 0.4,-5.89999 -0.1,-7.49922 -2.5,-8.69922 -2.2,-0.9 -3,-0.60039 -6.5,2.09961 -3.9,3.1 -12.89922,7.60039 -13.69922,6.90039 -0.6,-0.7 1.69883,-8 3.29883,-10.5 2.1,-3.1 16.60079,-15.20039 23.80078,-19.90039 5.09999,-3.2 16.20078,-8.20078 18.80078,-8.30078 z m -218.65039,3.07617 c 1.075,-0.775 1.49883,0.82383 2.54883,3.92383 1,2.9 1.7,5.30078 1.5,5.30078 -6,2.8 -7.69922,3.79961 -7.69922,5.09961 0,0.8 1.6,3.20039 3.5,5.40039 3.7,4.1 4.39961,6.09961 2.09961,6.09961 -2,0 -12.29883,5.5 -15.79883,8.5 -1.7,1.4 -3.2,2.39922 -3.5,2.19922 -0.9,-1 6.09844,-17.79923 10.89844,-26.19922 3.65,-6.39999 5.37617,-9.54922 6.45117,-10.32422 z m 94.94922,1.625 c 0.2,-0.3 1.70039,-0.3 3.40039,0 1.9,0.3 3,0.99961 3,2.09961 0,1.2 -2.8,24.99922 -3,25.69922 0,0.1 -2.39922,0.40117 -5.19922,0.70117 -4.29999,0.5 -5.8,1.19883 -8,3.79883 -2.9,3.5 -5.80078,10.70118 -5.80078,14.70117 0,2.4 0.10079,2.5 7.80078,2 9.59999,-0.6 12.5,-2.30001 15,-9 1.6,-4.29999 1.99961,-4.70039 3.09961,-3.40039 0.7,0.9 1.6,1.39922 2,1.19922 0.4,-0.3 0.69883,0.30117 0.79883,1.20117 0,0.9 2.30156,6.59923 5.10156,12.69922 2.7,5.99999 5.99922,14.50079 7.19922,18.80078 2.1,7.59999 2.10078,7.89844 0.30078,9.89844 -1.7,2 -5.1,3.70117 -10,5.20117 -2.2,0.7 -2.30078,0.49921 -2.30078,-5.30078 0,-3.3 -0.4,-7.1 -1,-8.5 -0.8,-2.2 -1.29922,-2.39922 -3.69922,-1.69922 -1.6,0.5 -4.80156,1.8 -7.10156,3 -2.4,1.2 -4.59961,1.89922 -5.09961,1.69922 -0.5,-0.3 -1.89961,-2.99961 -3.09961,-6.09961 -1.2,-3 -2.69922,-5.5 -3.19922,-5.5 -0.5,0 -0.60078,0.4 -0.30078,1 0.3,0.5 -3.9e-4,0.70039 -0.90039,0.40039 -1,-0.4 -3.39961,1.3 -6.59961,4.5 -2.9,2.9 -4.89961,4.79883 -4.59961,4.29883 0.3,-0.5 -0.6,-0.69844 -2,-0.39844 -2.3,0.4 -2.50078,0.19961 -1.80078,-2.40039 0.6,-2.8 2.7,-7.0004 7.5,-14.90039 l 2.40039,-4 -4.69922,2.80078 C 823.20078,176.80078 820.9,178 820.5,178 c -0.4,0 -0.50078,-3.6 -0.30078,-8 0.3,-7.69999 0.20156,-8 -1.89844,-8 -2.1,0 -2.30078,0.49922 -2.30078,5.69922 0,7.69999 -1.9004,15.30158 -7.40039,30.10156 L 804,210.09961 V 227 c 0,9.39999 -0.40039,17 -0.90039,17 -1,0 -9.59922,-9.2 -12.69922,-13.5 l -2.09961,-3 0.5,3.5 c 0.3,2 2.69883,6.69961 5.79883,11.09961 l 5.30078,7.59961 L 794,252.30078 c -9.39999,4 -11.70078,9.99923 -8.80078,22.69922 1.2,5.09999 1.30078,5.19922 4.30078,4.69922 1.8,-0.3 6.40079,-1.49961 10.30078,-2.59961 l 7.09961,-2.19922 0.29883,-5.70117 L 807.5,263.5 l 2.40039,5 c 3.9,8.19999 5.19883,14.29962 5.79883,28.09961 0.9,19.69998 -1.89923,32.30003 -14.69922,66.5 -9.39999,25.09997 -15.8004,44.49963 -21.40039,64.59961 -5.1,18.59998 -9.09922,40.80118 -8.19922,46.20117 0.4,2.5 0.19961,3.09961 -0.40039,2.09961 -0.7,-1.1 -1,-7.8e-4 -1,3.69922 0,3.1 -0.49961,5.30078 -1.09961,5.30078 -0.5,0 -1.89961,-2.2 -3.09961,-5 -1.1,-2.8 -2.50156,-4.70078 -3.10156,-4.30078 -0.6,0.3 -0.69883,0.1 -0.29883,-0.5 0.4,-0.6 -2.30039,-6.59923 -5.90039,-13.19922 -17.79998,-32.69997 -36.10079,-71.80003 -46.80078,-100 -10.29999,-27.29997 -11.79962,-31.90002 -18.09961,-55 -5,-18.09998 -3.2996,-22.90079 11.40039,-32.30078 9.39999,-6 26.5004,-24.80002 36.40039,-40 2.9,-4.5 6.50039,-9.29961 7.90039,-10.59961 l 2.59961,-2.29883 3.40039,3.39844 c 3.2,3.3 11.29883,7.80078 13.79883,7.80078 0.7,0 3.1,-1.20078 5.5,-2.80078 4,-2.5 4.30078,-2.99844 3.80078,-5.89844 -0.8,-4.09999 1.00001,-9.50079 5.5,-16.30078 2,-3 3.90039,-5.99922 4.40039,-6.69922 1.5,-2.3 -4.80079,-4.50078 -10.80078,-3.80078 l -5.40039,0.69922 L 771,194.90039 c 1.9,-6.49999 1.90077,-6.50078 -7.19922,-5.80078 -4.59999,0.3 -10.2,1.00078 -12.5,1.30078 l -4.20117,0.79883 -1,-3.89844 c -1.6,-5.79999 -2.99961,-12.7 -3.09961,-14 0,-0.7 1.6,-1.50039 3.5,-1.90039 4.2,-0.7 9.69961,-4.90118 13.59961,-10.20117 2.6,-3.6 3.20079,-3.9 8.30078,-4 3,-0.1 7.19922,-0.79922 9.19922,-1.69922 2.1,-0.8 4.00039,-1.20078 4.40039,-0.80078 1.4,1.7 10,21.20156 10,22.60156 0,1.5 7.99961,4.69922 11.59961,4.69922 1.9,0 8.40039,-8.1 8.40039,-10.5 0,-2.7 -5.09961,-10.70039 -8.09961,-12.90039 C 802.60039,157.69961 799.6,157 797,157 c -2.5,0 -5.49922,-0.29961 -6.69922,-0.59961 l -2.20117,-0.59961 2.30078,-3.20117 c 1.3,-1.7 2.69961,-4.10039 3.09961,-5.40039 0.4,-1.4 2.8,-3.29922 6.5,-5.19922 6.89999,-3.5 8,-3.69961 8,-1.09961 0,1 0.7,2.39961 1.5,3.09961 1.4,1.1 2.19961,0.70039 5.59961,-2.59961 3.5,-3.4 3.90039,-4.30117 3.40039,-7.20117 -0.6,-3.1 -0.49922,-3.19922 2.80078,-3.19922 2.4,0 3.89844,0.59961 4.89844,2.09961 1.9,3 12.50157,3.2 18.60156,0.5 4.7,-2.1 4.79883,-2.6004 2.79883,-8.40039 -0.8,-2.2 -1.3,-4.19844 -1,-4.39844 z m -163.23633,7.22461 c 0.8875,-0.125 2.38711,1.02461 4.53711,3.47461 2.4,2.8 3.09961,4.49922 3.09961,7.69922 0,3.5 0.6,4.7 4.5,8.5 l 4.40039,4.5 -5.40039,7.10156 c -8.19999,10.89999 -9.30039,12.19961 -9.90039,11.59961 -0.3,-0.3 -0.79961,-8.20001 -1.09961,-17.5 -0.3,-9.29999 -0.8,-18.80117 -1,-21.20117 -0.3,-2.65 -0.0242,-4.04883 0.86328,-4.17383 z m 299.59961,5.9375 c 0.4625,-0.1625 2.93711,1.23672 7.53711,4.13672 5.99999,3.7 6.69922,3.9 10.1992,3 5.1,-1.4 6.6,0.19961 2.5,2.59961 -3.69998,2.1 -3.69959,2.40079 -1.5996,7.80078 l 1.5,4.09961 L 998.59961,163 c -2.8,4.7 -5.2,10.5 -6.5,15.5 -3.3,12.99999 -8.90039,25.5 -11.40039,25.5 -1.6,0 -0.19844,-39.89922 1.60156,-44.19922 0.8,-1.8 3.29961,-6.20078 5.59961,-9.80078 l 4.19922,-6.5 -4.59961,-4.30078 c -3.45,-3.35 -4.99961,-5.07383 -4.53711,-5.23633 z M 676,144 c 0.5,0 0.7,2.79922 0.5,6.19922 -0.3,3.5 -0.7,9.70078 -1,13.80078 -0.2,4.1 -2.3,15.10001 -4.5,24.5 -5.69999,24.19998 -6.49961,32.20041 -5.09961,52.40039 0.7,9.09999 1.6,19.39883 2,22.79883 0.8,7.09999 -3.9e-4,7.19999 -2.90039,0.5 -1.2,-2.6 -6.39961,-11.69845 -11.59961,-20.39844 -15.69998,-25.99997 -16.40038,-30.40002 -7.90039,-48 4.7,-9.79999 5.49922,-12.40079 6.19922,-19.80078 l 0.80078,-8.5 5.30078,-1.90039 c 2.9,-1 5.19922,-2.49883 5.19922,-3.29883 C 663,158.70078 673.5,144 676,144 Z m 52,3.19922 c 0.9,0.2 1.69922,2.00156 2.19922,4.60156 0.4,2.3 2.70117,8.40001 5.20117,13.5 6.09999,12.49999 7.8,20.59962 6,29.59961 -0.8,4.1 -1.70039,3.89921 -3.90039,-0.80078 -2.2,-4.9 -8.1,-10.9 -9.5,-9.5 -0.5,0.5 -1.19961,3.40039 -1.59961,6.40039 -0.6,5.39999 -0.50117,5.6 3.29883,9.5 2.1,2.3 5.00156,6.50039 6.60156,9.40039 2.6,5.1 2.69961,5.49883 1.09961,7.29883 -1.6,1.7 -1.6,2.00117 0.5,4.20117 l 2.19922,2.40039 -5.29883,8.59961 c -2.9,4.7 -5.60156,8.59961 -6.10156,8.59961 -0.5,0 -0.59922,-0.80078 -0.19922,-1.80078 0.3,-0.9 1.29922,-4.19883 2.19922,-7.29883 l 1.60156,-5.70117 -3.90039,-5.5 c -3.4,-4.9 -4.30078,-5.59922 -7.30078,-5.69922 -2.8,0 -4.70001,-1.30079 -11.5,-7.80078 -4.4,-4.2 -8.19922,-7.69922 -8.19922,-7.69922 -0.1,0 -0.40078,-2.09961 -0.80078,-4.59961 l -0.69922,-4.5 h 4.90039 c 8.59999,0 15.2,-6.6004 16,-15.90039 l 0.39844,-5.40039 -3.79883,0.59961 c -7.69999,1.3 -7.70039,1.30038 -8.90039,-4.59961 l -1,-5.40039 -3.19922,3.70117 c -3.7,4.2 -5.40078,7.29884 -8.30078,14.79883 -1.7,4.69999 -10.59961,16.5 -11.59961,15.5 -0.8,-0.8 6.8004,-17.29923 12.90039,-28.19922 l 6.19922,-11 5,-0.59961 c 2.8,-0.4 7.9,-2.10078 11.5,-3.80078 3.6,-1.7 7.2,-3.00039 8,-2.90039 z m 405.291,3.9043 c 0.3031,0.26562 0.1844,1.1207 -0.3906,2.5957 -2.6,6.99999 -18.5996,24.7 -25.5996,28.5 -2.7,1.5 -6.9004,3.00039 -9.4004,3.40039 -3.7,0.6 -5.4012,1.8 -10.7012,7 l -6.1992,6.30078 -6.4004,-4.70117 c -3.6,-2.6 -6.5,-4.99883 -6.5,-5.29883 -0.1,-0.4 3.5008,-0.8 7.8008,-1 7.1,-0.4 8.2988,-0.7 11.7988,-3.5 3.2,-2.6 9.4016,-11.20117 10.6016,-14.70117 0.1,-0.5 2.4984,-0.19961 5.3984,0.90039 4.9,1.7 5.4016,1.70117 8.6016,0.20117 1.9,-0.9 5.3988,-4.00039 7.7988,-6.90039 7.575,-9.07499 12.282,-13.59375 13.1914,-12.79687 z m -552.11522,3.30859 c 0.95,-0.3125 1.975,-0.21211 3.625,0.0879 2,0.4 7.09922,1.00039 11.19922,1.40039 l 7.5,0.59961 0.80078,5.30078 c 0.5,3 0.69844,5.59883 0.39844,5.79883 -0.2,0.2 -2.59883,-0.49922 -5.29883,-1.69922 -5.19999,-2.4 -16.49961,-3.2 -21.59961,-1.5 C 575.20078,165.10039 575,165 575,162.5 c 0,-1.6 1.2,-3.79922 3,-5.69922 1.35,-1.35 2.22578,-2.07617 3.17578,-2.38867 z m 318.75,0.98828 c 1.025,0.525 1.87461,2.14961 2.97461,5.09961 3.7,9.79999 1.9,25.69962 -3.5,31.59961 -1.7,1.9 -1.80117,1.90039 -5.20117,-0.0996 l -3.39844,-2.09961 0.39844,5.79883 c 0.5,7.49999 3.60001,13.50118 11,21.70117 6.69999,7.39999 7.5004,7.49961 17.40039,3.09961 l 5.09961,-2.19922 0.60156,-7.90039 c 0.4,-4.4 1.59922,-11.1004 2.69922,-14.9004 1.2,-3.9 2.19922,-7.60078 2.19922,-8.30078 0.2,-1.6 1.60079,-1.49922 6.80078,0.30078 2.3,0.8 5.50039,1.5 6.90039,1.5 3.4,0 10.39962,-4.4004 15.59961,-9.90039 2.1,-2.2 4.19961,-4.09961 4.59961,-4.09961 0.8,0 8.90039,10.40039 8.90039,11.40039 0,0.5 -0.69961,0.49922 -1.59961,0.19922 -2.3,-0.9 -10.3,1.2 -13.5,3.5 -1.6,1.1 -3.80039,3.3 -4.90039,5 L 950.80078,198 956.5,200.59961 c 4.7,2.2 6.69922,2.60117 11.69922,2.20117 l 6,-0.40039 -0.69922,11 c -1.4,22.69998 -5.60079,34.00001 -16.80078,45 -5.5,5.4 -9.49962,8.30039 -16.59961,11.90039 -5.2,2.5 -9.59883,4.49883 -9.79883,4.29883 -0.3,-0.2 0.0988,-3.29922 0.79883,-6.69922 1.5,-7.69999 1.70039,-30.7004 0.40039,-40.40039 -0.9,-6.49999 -0.99961,-5.59998 -1.59961,12 -0.9,29.59997 -2.20119,32.90002 -20.70117,54.5 -1.7,1.9 -3.29883,3.20078 -3.79883,2.80078 -0.9,-0.9 -4.30039,-18.10039 -4.40039,-21.90039 0,-2.5 0.1,-2.59961 5,-2.09961 5.89999,0.5 13.09922,-1.80157 17.19922,-5.60156 l 2.70117,-2.5 -1.80078,-3.5 c -1.6,-3.1 -1.6,-3.99961 -0.5,-8.09961 0.7,-2.5 1.39961,-10.09923 1.59961,-16.69922 L 925.5,224.19922 921.5,227 c -4.3,2.9 -10.89922,6.39961 -15.19922,8.09961 -2.6,0.9 -2.60039,0.90078 -1.90039,-3.19922 l 0.69922,-4.20117 L 901.30078,231 c -2,1.8 -5.60039,4.09922 -7.90039,5.19922 -4.49999,1.9 -14.10078,4.10039 -15.30078,3.40039 -0.4,-0.3 0.29961,-5.59883 1.59961,-11.79883 3.1,-14.49998 2.70078,-24.90079 -1.19922,-36.30078 -1.4,-4.4 -2.70078,-8.4 -2.80078,-9 -0.1,-0.6 3.40039,-0.79961 8.90039,-0.59961 8.79999,0.3 9.10078,0.19922 10.80078,-2.30078 1.6,-2.5 1.6,-2.7 -0.5,-5.5 -3,-4.1 -2.4,-14.49922 1,-17.19922 1.8,-1.45 3.00039,-2.025 4.02539,-1.5 z m 15.125,2.61133 c 3.525,-0.0125 6.94922,0.28867 6.94922,0.88867 0,0.5 -0.49961,2.19883 -1.09961,3.79883 -0.6,1.7 -0.80078,6 -0.30078,11.5 l 0.70117,8.70117 4.39844,-0.59961 c 4.39999,-0.6 4.50078,-0.50117 3.80078,1.79883 -0.9,3 -3.10078,6.69961 -5.80078,9.59961 l -2,2.20117 -6.69922,-6 -6.69922,-6 -0.70117,-12 C 907.19961,165.3004 907.2,159.5 907.5,159 c 0.4,-0.65 4.02578,-0.97578 7.55078,-0.98828 z M 647.59961,169.1875 c 1.2,0.0875 1.40039,0.61328 1.40039,1.61328 0,2.8 -3.59961,12.69883 -6.59961,18.29883 -2.1,3.8 -3.30078,4.90117 -4.80078,4.70117 -2,-0.3 -2.10039,-1.00001 -2.40039,-11.5 l -0.29883,-11.10156 5.29883,-1 c 3.99999,-0.75 6.20039,-1.09922 7.40039,-1.01172 z M 1003,170 c 0.8,0 4.8008,0.90039 8.8008,1.90039 8.4,2.2 8.7,2.40039 7,3.90039 -1.8,1.4 -21.1008,12.3 -21.3008,12 -0.1,-0.2 0.70039,-4.2 1.90039,-9 C 1000.9004,172.60079 1001.9,170 1003,170 Z m 61.5996,10 c 1.8,0 2.4004,0.49922 2.4004,2.19922 0,2.1 -3.4992,6.80078 -5.1992,6.80078 -0.9,0 -1.1012,-5.10039 -0.2012,-7.40039 0.3,-0.9 1.7,-1.59961 3,-1.59961 z m -41.2988,2 c 0.4,0 0.6992,0.40039 0.6992,0.90039 0,0.5 -0.2996,2.3 -0.5996,4 -0.6,2.9 -0.8996,3.09961 -5.0996,3.09961 -4.3,0 -4.4004,3.9e-4 -2.9004,-2.09961 1.5,-2.2 6.4004,-5.90039 7.9004,-5.90039 z m -540.00002,7.09961 c 0.5,-0.1 3.9,1.90078 7.5,4.30078 11.59999,7.69999 11.09844,6.99922 8.39844,11.69922 -1.3,2.2 -1.99922,4.40039 -1.69922,4.90039 0.3,0.6 3.29961,1 6.59961,1 H 510 v 5.59961 c 0,5.29999 -0.19922,5.80078 -4.19922,9.30078 l -4.10156,3.69922 3.40039,2.5 c 1.9,1.4 5.90039,3.60039 8.90039,4.90039 3,1.3 5.90039,2.69961 6.40039,3.09961 0.4,0.4 1.19961,4.89961 1.59961,10.09961 0.5,5.09999 1.2,10.3 1.5,11.5 0.7,2.1 0.40078,2.30078 -3.19922,2.30078 -3.1,0 -4.80157,0.79961 -9.10156,4.59961 -2.9,2.6 -5.49883,4.40117 -5.79883,4.20117 -0.2,-0.3 1.19922,-4.00039 3.19922,-8.40039 2.8,-5.89999 3.70078,-9.30039 3.80078,-13.40039 0.1,-5.09999 -0.30079,-6.0004 -4.80078,-11.90039 -2.7,-3.4 -6.3,-8.60039 -8,-11.40039 L 496.5,222.5 h -13.80078 l -1.09961,-4 c -1,-3.4 -1.6,-3.99922 -5,-4.69922 -2.1,-0.5 -5.09922,-1.6 -6.69922,-2.5 L 467,209.59961 l -4.19922,4.09961 c -3.7,3.4 -5.30079,4.20156 -10.80078,5.10156 -3.6,0.5 -7.29922,1.29922 -8.19922,1.69922 -1.3,0.5 -2.70117,-0.99962 -6.20117,-7.09961 C 435.09961,209.10039 433,205.1 433,204.5 c 0,-0.6 4.5004,-2.50039 9.90039,-4.40039 l 10,-3.29883 2.90039,2.09961 c 3.6,2.5 4,2.59883 6.5,0.29883 1.7,-1.5 19,-9.79961 21,-10.09961 z m 356.13672,1.95117 c 0.2625,0.125 0.61289,0.64883 0.96289,1.54883 0.3,0.8 0.19961,1.20039 -0.40039,0.90039 -0.6,-0.3 -1,-0.99961 -1,-1.59961 0,-0.7 0.175,-0.97461 0.4375,-0.84961 z M 1008.3008,192 c 0.4,0 0.6992,0.79922 0.6992,1.69922 0,2.7 4.5004,11.60117 6.9004,13.70117 1.7,1.6 2.0996,2.99883 2.0996,8.29883 V 222 h 4 c 2.3,0 6.9996,-0.69961 10.5996,-1.59961 3.6,-0.9 6.6004,-1.50078 6.9004,-1.30078 0.2,0.2 -0.2996,4.09961 -1.0996,8.59961 -1.5,8.89999 -4.0996,16.40118 -8.0996,23.20117 l -2.6016,4.5 -3,-3.40039 -3,-3.40039 -2.1992,2.80078 c -3.1,3.8 -5.5,11.79883 -5.5,17.79883 0,4.69999 -0.099,5.00039 -3.6992,6.40039 C 986.20081,285.2996 966.49998,294.20001 945,305 c -14.49999,7.19999 -18.19922,9.5 -17.19922,10.5 0.4,0.5 0.1,0.49961 -0.5,0.0996 -1.6,-0.9 -5.6,1.70039 -4.5,2.90039 0.4,0.5 0.1,0.49961 -0.5,0.0996 -0.7,-0.4 -2.20039,-0.19961 -3.40039,0.40039 -1.1,0.7 -2.29961,0.89922 -2.59961,0.69922 -0.3,-0.3 1.19844,-3.79844 3.39844,-7.89844 5.19999,-9.89999 13.2008,-17.70001 28.80078,-28 14.79999,-9.69999 24.3,-18.90001 28,-27 1.3,-2.9 4,-10.90079 6,-17.80078 1.9,-6.89999 5.40078,-16.3 7.80078,-21 4,-8.09999 16.40002,-26 18.00002,-26 z m -445.10158,3.09961 c 0.3,-0.1 2.00117,3.39961 3.70117,7.59961 l 3,7.80078 -5,0.80078 L 560,212.19922 V 218 c 0,7.79999 2.20001,12.29961 7.5,15.09961 3.3,1.8 4.69961,2.00117 7.09961,1.20117 3.5,-1.1 6.40039,-6.30039 6.40039,-11.40039 0,-3.3 -0.19922,-3.6 -4.19922,-4.5 l -4.20117,-0.90039 -2.19922,-8.19922 c -1.2,-4.49999 -1.99961,-8.40117 -1.59961,-8.70117 0.3,-0.3 3.29883,-0.59961 6.79883,-0.59961 6.59999,0 7.40039,0.70079 7.40039,6.80078 0,2 0.40078,2.19922 2.30078,1.69922 1.2,-0.4 4.09844,-1.79961 6.39844,-3.09961 2.3,-1.3 4.80117,-2.40039 5.70117,-2.40039 2.9,0 8.29961,7.90079 11.09961,16.30078 3.3,9.99999 3.69961,21.09923 1.09961,33.69922 -2.4,11.29999 -5.49961,21 -6.59961,21 -0.6,0 -1,-0.40078 -1,-0.80078 0,-1.3 -11.89962,-14.89845 -19.09961,-21.89844 -3.5,-3.3 -10.9004,-9.5 -16.40039,-13.5 -11.09999,-8.19999 -26.3,-23.10079 -30,-29.30078 -1.3,-2.2 -2.69961,-5.30039 -3.09961,-6.90039 -0.6,-2.6 -0.50078,-2.79922 1.19922,-2.19922 1,0.4 4.49961,1.39922 7.59961,2.19922 5.39999,1.4 6,1.4 10,-0.5 2.4,-1 5.70078,-2.80039 7.30078,-3.90039 1.7,-1.1 3.29922,-2.09961 3.69922,-2.09961 z m 46.40039,1.09961 c 0.1,0.2 1.90039,4.80079 3.90039,10.30078 4,10.59999 10.80001,23.70002 21,40 10.89999,17.49998 15.3004,25.60002 22.40039,41.5 5.6,12.69999 20.7,51.60079 23.5,60.80078 0.4,1.2 1.19922,1.99922 1.69922,1.69922 0.5,-0.3 0.60117,0.19961 0.20117,1.09961 -0.4,1.2 -0.3,1.5 0.5,1 0.7,-0.4 0.99961,-0.19883 0.59961,0.70117 -0.2,0.7 0.39961,3.49922 1.59961,6.19922 3.2,7.39999 2.70077,7.19921 -8.69922,-3.30078 -1.3,-1.2 -3.00156,-1.89961 -3.60156,-1.59961 -0.7,0.4 -0.9,0.40039 -0.5,-0.0996 1.1,-1.1 -1.19961,-3.60078 -2.59961,-2.80078 -0.6,0.3 -0.79883,0.20078 -0.29883,-0.19922 0.4,-0.5 -5.00157,-6.7004 -12.10156,-13.90039 -19.09998,-19.49998 -30.10001,-33.8004 -36.5,-47.40039 L 617.5,283.5 l -0.0996,-26 c 0,-23.99998 -0.20117,-26.49962 -2.20117,-32.59961 -2.5,-7.49999 -12.29922,-23.40117 -13.69922,-22.20117 -0.6,0.5 -0.0996,-0.19844 0.90039,-1.39844 1,-1.3 2.19961,-2.00078 2.59961,-1.80078 0.4,0.2 1.5,-0.4 2.5,-1.5 1,-1.1 1.89961,-1.90078 2.09961,-1.80078 z m 3,2.10156 c 0.3,0.3 0.39961,1.19844 0.0996,1.89844 -0.3,0.8 -0.59961,0.50039 -0.59961,-0.59961 -0.1,-1.1 0.2,-1.69883 0.5,-1.29883 z M 816.5,201 c 0.7,0 1.90078,1.89922 2.80078,4.19922 0.9,2.4 1.69844,5.40078 1.89844,6.80078 0.3,2.3 -10e-6,2.49961 -4.5,2.59961 -2.6,0.1 -4.99883,-0.29883 -5.29883,-0.79883 C 810.50039,212.40078 815.1,201 816.5,201 Z m 497.6992,4.19922 -3.7988,3.90039 c -3.5,3.6 -4.2996,3.90039 -9.0996,3.90039 -4.8,0 -5.3008,0.19922 -5.3008,2.19922 0,1.2 1.4,4.6 3,7.5 l 3.0996,5.40039 -2,4.40039 c -1.1,2.5 -2.5988,4.5 -3.2988,4.5 -0.7,0 -3.9016,-2.09961 -7.1016,-4.59961 -4.6,-3.6 -5.4996,-4.00039 -4.0996,-1.90039 2.6,3.9 9.5008,8.3 16.8008,10.5 5,1.6 7.2,2.90078 9.5,5.80078 1.7,2 3.0996,3.09844 3.0996,2.39844 0,-1.8 1.4996,-1.49922 5.5996,1.30078 7,4.7 16.6004,2.30077 22.4004,-5.69922 3.8,-5.19999 3.8,-6.50078 0,-7.30078 -1.6,-0.4 -3,-1.00078 -3,-1.30078 0,-0.4 1.4,-2.69922 3,-5.19922 1.6,-2.5 3,-4.79922 3,-5.19922 0,-0.4 -1.1992,-1.40117 -2.6992,-2.20117 -5.7,-3.1 -5.6012,-2.89923 -1.7012,-8.69922 l 3.5996,-5.40039 -8.2988,-0.80078 c -6.8,-0.7 -9.0012,-0.49961 -12.7012,0.90039 l -4.5,1.70117 -2.7988,-3.10156 z m -787.8867,4.5 c 0.2375,-0.325 0.68828,-0.19883 1.48828,0.20117 0.9,0.6 3.29883,2.50039 5.29883,4.40039 4.49999,4.2 15.3,18.59844 14.5,19.39844 -0.3,0.3 -4.29883,0.001 -8.79883,-0.79883 L 530.5,231.5 l -2.19922,-8 c -1.2,-4.4 -2.20117,-9.50078 -2.20117,-11.30078 -0.05,-1.4 -0.0246,-2.175 0.21289,-2.5 z M 1075.5,210 c 1.3,0 6.0996,1.8 10.5996,4 4.6,2.2 9.3,4 10.5,4 3.3,0 3.8012,1.60039 1.2012,3.40039 -2.8,2 -5.7004,1.99961 -8.4004,0.0996 -1.9,-1.3 -2.3004,-1.2 -4.4004,1 -2.4,2.5 -5.4008,3.19961 -7.8008,1.59961 -1.1,-0.6 -1.0988,-1.69961 -0.2988,-4.59961 1.1,-4.1 0.7996,-5.5 -1.4004,-5.5 -0.7,0 -5.9992,5.89962 -11.6992,13.09961 -14.9,18.99998 -15.4008,19.60117 -20.3008,21.20117 -2.4,0.9 -4.6008,1.29961 -4.8008,1.09961 -0.9,-0.9 3.4012,-15.80079 6.2012,-21.30078 l 2.6992,-5.59961 7.5,-0.59961 c 6.9,-0.5 11.9004,-2.00117 11.9004,-3.70117 0,-2.1 6.4,-8.19922 8.5,-8.19922 z m 61.3008,3.08789 c 0.625,-0.0625 1.2,0.31133 2,1.11133 3,3 -0.1016,7.30156 -3.6016,5.10156 -2.3,-1.4 -2.4992,-2.90117 -0.6992,-4.70117 1,-0.95 1.6758,-1.44922 2.3008,-1.51172 z m 70.8867,1.03906 c 1.4562,0.0172 2.3125,0.19727 2.3125,0.57227 0,0.4 -0.9992,1.7 -2.1992,3 -1.6,1.7 -3.3008,2.30078 -6.3008,2.30078 -4.1,0 -4.1008,4e-4 -5.8008,5.40039 -2.4,7.49999 -3.6,8.29961 -10,6.59961 -7.3,-1.9 -18.1,-6.30039 -18,-7.40039 0.2,-1.3 9.7016,-5 17.6016,-7 8.25,-2.1 18.018,-3.52422 22.3867,-3.47266 z M 679.30078,217 c 0.3,0 2.8,1.1 5.5,2.5 6,3 9.49923,3.1 14.69922,0.5 2.2,-1.1 4.19961,-2 4.59961,-2 0.4,0 0.50117,0.9 0.20117,2 -0.3,1.1 -1.10039,2 -1.90039,2 -3,0 -12.40039,4.19961 -12.40039,5.59961 0,2.4 3.59922,8.69961 6.69922,11.59961 3.2,3.1 10.8,6.70078 14,6.80078 1.6,0 2.90039,1.19922 4.40039,4.19922 l 2.20117,4.20117 -3.90039,3.40039 c -2.2,1.8 -9.5004,6.89883 -16.40039,11.29883 l -12.5,8 -0.69922,-3.29883 c -2.3,-10.99999 -5.9,-56.80078 -4.5,-56.80078 z m 432.34372,0.89453 c 0.2235,0.0547 0.3555,0.28008 0.3555,0.70508 0,1 -5.1992,9.4 -6.1992,10 -1.4,0.9 0,-3.3004 2.3984,-7.40039 1.275,-2.1 2.775,-3.46875 3.4453,-3.30469 z m -301.74411,8.80469 4.19922,0.60156 c 2.3,0.3 4.59961,1 5.09961,1.5 0.5,0.5 0.10039,2.7 -1.09961,5.5 -1.1,2.5 -2.29922,6.29883 -2.69922,8.29883 -0.7,3.7 -0.69961,3.69961 2.40039,3.09961 1.7,-0.3 5.3,-1.99961 8,-3.59961 L 830.69922,239 l 5,4.80078 c 6.29999,5.9 6.90078,6.99884 7.80078,15.79883 1.2,12.89999 -3.99962,33.90041 -12.59961,50.40039 l -3.40039,6.5 -0.59961,-14 c -0.7,-15.89998 -1.1004,-17.90002 -7.40039,-36.5 -5.09999,-15.29998 -8.00039,-26.0004 -8.90039,-33.90039 z M 1157.9004,227 c 0.3,0 5.0996,1.30078 10.5996,2.80078 5.5,1.6 12.9992,3.49922 16.6992,4.19922 5.8,1 7.0008,1.70039 8.8008,4.40039 l 2,3.09961 -3.0996,2.80078 c -2.6,2.4 -2.9008,3.1 -1.8008,4.5 3.1,4.1 8.7008,9.49844 11.8008,11.39844 l 3.4004,2.20117 -6,1.19922 c -9.8,2 -16.3012,6.19962 -20.2012,13.09961 -0.8,1.5 -1.4004,1.50039 -4.9004,0.40039 -6.7,-2.2 -6.8984,-2.49961 -4.8984,-6.59961 1,-2 1.4992,-4 1.1992,-4.5 -1.3,-2.1 -10.1008,-2.9 -16.8008,-1.5 -3.7,0.8 -7.1992,1.7 -7.6992,2 -1.6,1 -1.2004,-0.9004 2.5996,-14.90039 3.9,-14.49999 7.3008,-24.59961 8.3008,-24.59961 z m -286.23829,0.22461 c 0.8875,-0.325 1.33789,-0.27422 1.33789,0.17578 0,0.1 -0.49961,1.79922 -1.09961,3.69922 -0.6,1.9 -1.3,4.00078 -1.5,4.80078 -0.3,1 -1.2,0.7 -3.5,-1.5 L 864,231.59961 l 3.69922,-2.29883 c 1.75,-1.05 3.07539,-1.75117 3.96289,-2.07617 z m 279.15039,1.90039 c 0.9625,0.05 1.7883,1.27578 1.9883,3.67578 0.3,3.3 -1.9008,8.19922 -3.8008,8.19922 -1.7,0 -2.2,-4.8 -1,-8.5 0.75,-2.3 1.85,-3.425 2.8125,-3.375 z m -658.36328,0.0371 c 1.4,-0.4375 2.65078,0.1875 3.55078,1.9375 3,5.59999 -1.9004,10.20117 -10.40039,9.70117 -6.29999,-0.3 -7.70039,-1.8 -2.90039,-3 2,-0.5 3.90078,-2.10117 5.30078,-4.20117 1.5,-2.5 3.04922,-4 4.44922,-4.4375 z M 1067.3008,231 c 0.4,0 0.6992,3.1 0.6992,7 v 7 h -5.5996 c -3.1,0 -5.4012,-0.40078 -5.2012,-0.80078 0.7,-1.9 9.4016,-13.09922 10.1016,-13.19922 z m -542.05471,4.19922 c 2.59688,-0.30469 6.6297,0.95039 15.55469,4.40039 13.19999,5.29999 20.79962,10.00079 30.09961,18.80078 8.69999,8.39999 20.29883,23.79922 18.79883,25.19922 -0.3,0.3 -4.29883,-1.39883 -8.79883,-3.79883 -4.59999,-2.3 -8.69961,-4.40156 -9.09961,-4.60156 -0.5,-0.2 -0.80078,1.00117 -0.80078,2.70117 0,1.9 -0.79922,3.5 -2.19922,4.5 -3.4,2.4 -4.30039,0.59921 -2.90039,-5.80078 1.4,-6.09999 1.30039,-6.29923 -4.59961,-15.19922 -2.9,-4.39999 -3.70117,-4.99961 -6.20117,-4.59961 -2.3,0.3 -3.3,-0.30117 -6,-3.70117 C 546.89961,250.39961 545,249 543.5,249 540.6,249 525.80078,241.50078 523.30078,238.80078 521.80078,237.10078 521.8,236.7 523,236 c 0.675,-0.425 1.38047,-0.69922 2.24609,-0.80078 z m -76.93359,2.73828 c 0.7125,-0.0375 2.43672,3.71173 4.88672,11.26172 C 458.19921,264.7992 458.29961,265 457.09961,265 455.19961,265 451,256.30038 449.5,249.40039 447.9,241.8004 447.6,237.975 448.3125,237.9375 Z M 1104,242 c 1.9,0 3,0.50039 3,1.40039 0,1.9 -3.9996,6.69961 -5.0996,6.09961 -0.5,-0.4 -0.9004,-2.19961 -0.9004,-4.09961 0,-3.1 0.2,-3.40039 3,-3.40039 z m -257.63672,0.23828 c 0.3625,-0.1375 1.18633,0.0113 2.73633,0.36133 1.9,0.5 6.90118,0.90039 11.20117,0.90039 4.2,0 7.69922,0.10078 7.69922,0.30078 0,2.2 -14.79961,28.19922 -16.09961,28.19922 -0.3,0 -1.50117,-6.00079 -2.70117,-13.30078 -1.1,-7.29999 -2.39883,-14.09961 -2.79883,-15.09961 -0.3,-0.8 -0.39961,-1.22383 -0.0371,-1.36133 z m 48.88672,0.0859 c 1,0.1 0.74961,1.025 0.34961,2.875 -0.4,1.8 -2.19922,4.90039 -4.19922,6.90039 l -3.5,3.59961 3.40039,2.90039 c 1.9,1.6 4.19961,3.30039 5.09961,3.90039 1.9,1.1 4.59961,15.69962 4.59961,25.09961 0,6.99999 -1.79961,16.09961 -4.09961,20.59961 -2.6,5.09999 -13.1004,16.10157 -25.90039,27.10156 -20.49998,17.69998 -31.50079,28.39845 -42.80078,41.39844 -6.2,7.29999 -11.59883,13.30078 -11.79883,13.30078 -1.5,0 8.49962,-29.4008 15.59961,-45.80078 3.2,-7.59999 11.10001,-24.69923 17.5,-38.19922 10.59999,-22.19998 20.1,-43.10079 23.5,-51.80078 1.7,-4.3 5.7004,-6.99883 14.40039,-9.79883 4.6,-1.45 6.84961,-2.17617 7.84961,-2.07617 z m 159.25,4.9375 c 0.225,-0.2625 0.8992,-0.11172 2.1992,0.23828 1.6,0.3 4.9016,1.59922 7.6016,2.69922 2.6,1.1 5.7992,3.40039 7.1992,4.90039 1.6,2 3.3004,2.90039 5.4004,2.90039 1.7,0 4.6004,0.69961 6.4004,1.59961 3.4,1.6 3.5,1.6 8,-1.5 2.5,-1.7 5.9988,-3.39883 7.7988,-3.79883 2.8,-0.5 1.8008,0.69962 -8.6992,11.09961 l -11.9004,11.69922 -9,-4.40039 c -4.9,-2.4 -10.6992,-4.89961 -12.6992,-5.59961 -2.1,-0.7 -3.8008,-1.39961 -3.8008,-1.59961 0,-0.1 0.7,-1.59961 1.5,-3.09961 1.7,-3.4 1.9,-9.10039 0.5,-12.90039 -0.5,-1.3 -0.725,-1.97578 -0.5,-2.23828 z m -597.40039,0.53906 7.70117,0.69922 c 4.2,0.4 7.79844,0.80039 7.89844,0.90039 0.5,0.4 -9.09883,15.59961 -9.79883,15.59961 -0.7,0 -2.80078,-5.90079 -4.80078,-13.30078 z m 684.80079,0.29883 c 0.2,0 0.099,2.20039 -0.3008,4.90039 -1,7.19999 -4.2992,17.30039 -6.6992,20.40039 -1.7,2.1 -2.8996,2.59961 -6.5996,2.59961 h -4.6016 l 0.6016,-5.5 0.5996,-5.5 -6.2012,0.59961 c -15,1.2 -16.5996,1.60039 -21.0996,4.40039 -6.2,3.9 -6.0988,1.70077 0.2012,-4.69922 5.9,-6.09999 12.3,-9.30078 18.5,-9.30078 3.1,0 4.6992,-0.40078 4.6992,-1.30078 0.1,-3 2.0996,-3.89961 11.0996,-5.09961 5.2,-0.8 9.6008,-1.4 9.8008,-1.5 z M 465.80078,273 h 6.89844 c 3.7,0 8.80078,0.69961 11.30078,1.59961 3.3,1.2 6.29961,1.40039 11.59961,0.90039 6.69999,-0.6 7.4004,-0.39922 15.90039,3.30078 l 9,3.89844 V 287.5 c 0,4.8 0.10079,4.9 5.30078,8 2.9,1.7 5.69883,3.79961 6.29883,4.59961 1.4,2.3 1.10078,6.40079 -1.19922,12.30078 -1.9,5.2 -1.90039,5.69922 -0.40039,8.69922 2,3.7 1.59999,4.10078 -5,6.30078 -9.69999,3.2 -15.30039,7.59922 -16.90039,13.19922 -0.6,1.8 0.1,2.50078 4.5,4.30078 7.59999,3.3 19.50001,6.09961 26,6.09961 6.79999,0 17.20079,-3 22.30078,-6.5 2,-1.3 3.59961,-2.29961 3.59961,-2.09961 0,0.2 -0.69961,3.89922 -1.59961,8.19922 -2.2,10.49999 -1.39961,26.5004 1.90039,39.40039 1.4,5.19999 2.89883,10.4 3.29883,11.5 0.4,1.1 -3.69883,-2.30001 -9.29883,-7.5 -9.39999,-8.99999 -19.30117,-15.79961 -21.20117,-14.59961 -0.9,0.6 -0.0992,6.69962 2.30078,16.59961 0.9,3.6 1.59961,8.3 1.59961,10.5 0,2.3 1.2,6.60078 3,10.30078 1.7,3.5 3,7.49844 3,8.89844 0,2.5 0.1004,2.50156 6.90039,2.10156 8.39999,-0.4 9.89961,-1.60118 14.59961,-11.20117 l 3.30078,-6.90039 3.69922,7.40039 c 10.99999,21.99998 28.09963,48.80042 52.09961,81.90039 17.59998,24.19998 17.70077,24.49921 9.30078,19.19922 -2.4,-1.6 -12.0004,-7.49923 -21.40039,-13.19922 -38.59996,-23.39998 -55.50002,-34.40079 -72,-46.80078 -35.49996,-26.69997 -51.9,-46.19963 -55.5,-66.09961 -0.9,-4.8 -0.8,-8.99923 0.5,-19.69922 1.9,-16.59998 1.99961,-37.7004 0.0996,-46.40039 -0.8,-3.6 -1.29961,-6.60078 -1.09961,-6.80078 0.2,-0.2 1.99961,-0.49883 4.09961,-0.79883 2.2,-0.2 5.60117,-1.09961 7.70117,-2.09961 3.6,-1.6 3.69922,-1.70118 3.69922,-7.20117 0.1,-3.1 0.7,-8.49961 1.5,-12.09961 1.5,-6.89999 1.79922,-13.39922 0.69922,-16.19922 -0.6,-1.4 -1.49845,-1.30078 -8.89844,1.19922 -5.09999,1.7 -10.00117,4.10039 -12.70117,6.40039 -5.4,4.4 -9.49961,4.79922 -13.09961,1.19922 -2.2,-2.2 -3.4,-5.3004 -5,-13.40039 z m -148.20117,3.30078 c 0.3,0.3 0.39961,1.19844 0.0996,1.89844 -0.3,0.8 -0.59961,0.50039 -0.59961,-0.59961 -0.1,-1.1 0.2,-1.69883 0.5,-1.29883 z M 1089,279 c 0.3,0 1.4004,1.59961 2.4004,3.59961 1.6,3.5 1.5988,3.70078 -0.2012,5.80078 -3.2,3.5 -4.2996,6.3004 -4.5996,12.40039 l -0.2988,5.69922 -8.4004,-0.40039 c -4.6,-0.2 -8.5012,-0.39961 -8.7012,-0.59961 -0.2,-0.1 2.9004,-4.89962 6.9004,-10.59961 C 1082.9996,284.8004 1087.7,279 1089,279 Z m -526.09961,3 c 0.3,0 1.4,1.19961 2.5,2.59961 1.8,2.5 1.79922,2.8 0.19922,6 -2.2,4.69999 -2.1,5.40118 1.5,9.70117 l 3.20117,3.79883 -1.30078,7.5 c -0.7,4.1 -1.50078,7.40039 -1.80078,7.40039 -0.2,0 -2.19844,-1.59961 -4.39844,-3.59961 -3.79999,-3.4 -18.8,-11.40039 -21.5,-11.40039 -2.5,0 -1.10116,-1.50039 4.79883,-5.40039 3.4,-2.2 8.29961,-6.59922 11.09961,-10.19922 2.9,-3.5 5.40117,-6.40039 5.70117,-6.40039 z m -126,0.0996 c 0.2,-0.1 3.00039,0.6 6.40039,1.5 6.89999,1.8 14.19844,5.20078 17.39844,8.30078 2.9,2.7 3.70039,9.00001 2.90039,22 -0.7,12.39999 -1.4,16.09961 -3,16.09961 -0.6,0 -4.8004,-2.79961 -9.40039,-6.09961 l -8.39844,-6.20117 3.79883,-3.79883 3.80078,-3.90039 -2.09961,-2.19922 c -2.4,-2.6 -6.80078,-11.00039 -7.80078,-14.90039 l -0.59961,-2.59961 -6.30078,1.19922 c -3.5,0.7 -7.70039,2 -9.40039,3 l -3,1.80078 -4.79883,-4.5 -4.70117,-4.5 L 421.09961,286 c 5.19999,-0.6 10.80039,-1.8 12.40039,-2.5 1.7,-0.7 3.20039,-1.40039 3.40039,-1.40039 z m 637.27541,1.08789 C 1074.9258,283.1 1075,283.4 1075,284 c 0,2.5 -19.0992,20.00079 -28.6992,26.30078 -5.6,3.7 -10.4012,6.69922 -10.7012,6.69922 -0.8,0 -0.7,-8.5 0,-13 l 0.5996,-3.5 h 6.6016 c 9.1,-0.1 11.4992,-1.49923 19.1992,-10.69922 1.8,-2.1 4.8008,-4.20156 7.3008,-5.10156 2.7,-0.95 4.125,-1.42422 4.875,-1.51172 z M 1013.6992,285 c 0.7,0 2.3,1.8 3.5,4 2.4,4.2 3.2,4.49961 7,3.09961 1.3,-0.4 3.5016,-1.1 5.1016,-1.5 2.7,-0.6 2.6992,-0.59882 2.6992,4.20117 0,12.09999 -7.9,31.69924 -21,52.19922 -7.6,11.99999 -13.1,22.99923 -17.5,35.19922 -4,11.29999 -11.0004,25.40118 -15.90039,32.20117 -9.09999,12.69999 -25.00043,25.10002 -62.40039,48.5 -34.79997,21.79998 -48.79924,32.50002 -69.69922,53.5 -16.79998,16.89998 -28.39962,31.59885 -41.09961,52.29883 -0.7,1 -1.50039,1.50156 -1.90039,1.10156 -1.6,-1.5 -0.89961,-52.60079 0.90039,-65.80078 3.9,-29.09997 10.59923,-52.90041 21.19922,-74.90039 5.89999,-12.29999 6.00117,-12.6 4.20117,-11.5 -0.7,0.4 -0.80039,0.3 -0.40039,-0.5 0.4,-0.6 1.09961,-0.89961 1.59961,-0.59961 0.4,0.3 1.80039,-1.09961 2.90039,-3.09961 7.49999,-12.69999 26.79962,-34.10001 41.09961,-45.5 15.79998,-12.69999 39.40002,-25.70118 57.5,-31.70117 14.89999,-5 14.4,-4.99843 18.5,1.10156 3.6,5.1 4.5004,5.59961 9.40039,4.59961 1.6,-0.3 2.20039,0.39883 2.90039,3.29883 0.5,2.1 1.29844,3.80078 1.89844,3.80078 0.5,0 3.40078,2.5 6.30078,5.5 3.7,3.9 6.00078,5.5 7.80078,5.5 2.9,0 3.49844,-1.70001 3.89844,-11.5 0.3,-5 0.30078,-4.99961 3.80078,-5.09961 1.9,0 5.20078,-0.10117 7.30078,-0.20117 L 997,343 v -9.19922 l 3.5996,0.59961 c 1.9,0.4 4.1008,0.89922 4.8008,1.19922 0.8,0.3 2.4,-1.3 4,-4 1.4,-2.4 4.1992,-5.80039 6.1992,-7.40039 3.5,-2.9 3.6,-3 2,-6 -1.7,-3.3 -6.6,-7.69922 -9.5,-8.69922 -1.5,-0.4 -2.0988,-1.70078 -2.2988,-4.80078 L 1005.5,300.5 999.80078,299.59961 C 996.60078,299.09961 994,298.4 994,298 c 0,-1.5 17.3992,-13 19.6992,-13 z M 597,290 c 0.4,0 2.59922,2.79922 4.69922,6.19922 2.2,3.5 5.40117,8.90156 7.20117,12.10156 l 3.19922,5.79883 -4.40039,-0.59961 -4.5,-0.69922 L 602.5,317 c -0.7,3.9 -0.90039,4.1 -3.40039,3.5 -4.1,-1 -13.3,-2.5 -16,-2.5 -1.5,0 -2.09883,0.40039 -1.79883,1.40039 0.3,0.7 0.79922,3.29883 1.19922,5.79883 l 0.69922,4.40039 L 579,331.5 c -3.8,1.7 -6,4.59961 -6,7.59961 0,0.6 3.4,2.6 7.5,4.5 7.19999,3.3 7.5,3.60039 7.5,6.90039 0,1.9 0.49961,4.50039 1.09961,5.90039 1,2.2 1.90079,2.50039 7.80078,2.90039 3.6,0.2 7.49961,0.69961 8.59961,1.09961 1.5,0.5 4.90001,-1.20079 13.5,-6.80078 6.29999,-4.2 11.99922,-7.59961 12.69922,-7.59961 3,0 24.40118,19.50001 32.20117,29.5 5.89999,7.49999 11.49883,16.39961 10.79883,17.09961 -0.3,0.3 -1.29844,0.1 -2.39844,-0.5 -2.6,-1.6 -16.80118,-7.89922 -23.70117,-10.69922 -3.4,-1.4 -6.40039,-2 -6.90039,-1.5 -0.8,0.8 3.40157,3.70001 13.10156,8.5 2.8,1.4 5.19922,2.79961 5.19922,3.09961 0,0.2 -4.39922,0.40039 -9.69922,0.40039 -5.39999,0 -15.90079,0.7 -23.30078,1.5 -7.39999,0.9 -15.5,1.59922 -18,1.69922 l -4.5,0.0996 3.5,1 c 1.9,0.5 13.60001,1.10117 26,1.20117 12.39999,0.2 24.59922,0.8 27.19922,1.5 23.89997,5.69999 38.00159,23.49928 63.10156,79.69922 18.39998,41.09996 28.59961,83.29965 30.09961,124.59961 0.2,6 -3.9e-4,10.80079 -0.40039,10.80079 -0.5,0 -2.00039,-2 -3.40039,-4.5 C 744.19962,598.70001 737.19997,590.59997 710.5,563 701.70001,553.90001 689.70038,541.49999 683.90039,535.5 640.60043,490.80004 609.29998,452.39997 588,418 c -9.99999,-15.99998 -18,-36.60079 -18,-45.80078 0,-2.6 0.90039,-6.49922 1.90039,-8.69922 1.5,-3.5 1.70039,-5.10001 0.90039,-11.5 -1,-7.69999 -0.6,-42.49961 0.5,-43.59961 0.3,-0.3 2.89922,0.19961 5.69922,1.09961 l 5.09961,1.69922 3.40039,-2.59961 c 3.1,-2.4 6.19961,-8.39962 8.09961,-16.09961 0.3,-1.4 1.00039,-2.5 1.40039,-2.5 z m 579.4961,4.87109 c -0.1781,0.005 -0.2703,0.97891 -0.1953,2.62891 0,2.2 0.1984,2.99922 0.3984,1.69922 0.2,-1.2 0.2,-3 0,-4 -0.075,-0.225 -0.1437,-0.32969 -0.2031,-0.32813 z m 93.6035,2.32813 c 0.3,-0.1 0.8012,0.10039 1.2012,0.40039 0.3,0.3 -0.8,4.30079 -2.5,8.80078 -2.3,5.89999 -3.1012,9.99962 -3.2012,15.09961 -0.1,4.9 -0.7004,7.8 -1.9004,9.5 -0.9,1.3 -1.6992,2.99922 -1.6992,3.69922 0,1.9 -11.8996,1.80039 -12.5996,-0.0996 -0.8,-2 1.1988,-8.8004 5.7988,-19.40039 2.2,-5.2 3.7004,-9.59844 3.4004,-9.89844 -1.1,-1.1 -7.5992,0.79961 -13.6992,4.09961 -3.7,1.9 -10.0008,4.1 -14.8008,5 l -8.5,1.69922 -1.7988,-2.29883 -1.7578,-2.24414 c 2.888,-2.17831 5.5087,-3.86992 7.8574,-5.05664 8.1,-4 21.0996,-6.90078 40.0996,-8.80078 1.9,-0.2 3.7996,-0.5 4.0996,-0.5 z m -797.54687,3.38086 c 0.83086,0.3123 2.07227,2.37031 3.94727,5.82031 4.3,7.69999 6.90039,18.99962 6.90039,30.09961 0,8.59999 -2.90039,29.10078 -4.40039,30.80078 -1.7,2 -7.69961,-20.60079 -8.09961,-30.80078 -0.2,-3.3 -0.19961,-12.90079 -0.0996,-21.30078 0.125,-10.81249 0.36719,-15.13965 1.75195,-14.61914 z M 1104,305.09961 c 0.3,0 2.2004,1.30039 4.4004,2.90039 3.6,2.8 3.9992,2.8 6.6992,1.5 1.5,-0.8 4.8004,-1.5 7.4004,-1.5 5.5,0 5.5004,0.0992 1.9004,13.69922 -3.3,12.49999 -10.0008,26.70118 -16.3008,34.70117 -2.8,3.5 -4.6988,6.69961 -4.2988,7.09961 0.4,0.5 0.1996,0.49922 -0.4004,0.19922 -1.5,-0.8 -5.6008,2.2 -4.8008,3.5 0.4,0.7 0.2,0.80039 -0.5,0.40039 -0.7,-0.5 -2.8004,0.19961 -4.9004,1.59961 -2,1.3 -9.0992,4.80117 -15.6992,7.70117 -27.9,12.49999 -42.6008,19.80001 -66.3008,33 -6.2,3.5 -11.49881,6.09883 -11.79881,5.79883 -0.3,-0.3 -0.1,-1.4 0.5,-2.5 0.60001,-1.1 4.39961,-11.89962 8.59961,-24.09961 9,-26.49997 13.3004,-35.39923 21.4004,-44.19922 8.3,-9.09999 13.8996,-12.50039 25.5996,-15.40039 5.4,-1.3 10.1992,-2.80078 10.6992,-3.30078 0.6,-0.6 -1.8992,-1.49883 -6.1992,-2.29883 -3.9,-0.8 -7.2008,-1.90039 -7.3008,-2.40039 -0.1,-0.6 3.6008,-2.90078 8.3008,-5.30078 7,-3.5 11.4,-4.79961 25.5,-7.59961 9.4,-1.9 17.2,-3.4 17.5,-3.5 z m -139.80078,2.09961 c 1.5,0.2 2.30078,1.00078 2.30078,2.30078 0,2.2 -3.09922,3.29922 -4.69922,1.69922 -1.7,-1.7 -0.10156,-4.4 2.39844,-4 z M 1218,311.58398 v 2.2168 c 0,3 -2.2004,7.09961 -5.4004,10.09961 -2.2,2 -2.5996,3.29961 -2.5996,7.59961 0,5 -0.2,5.40039 -4,7.90039 -5.3,3.5 -5.0992,5.70001 0.8008,12 2.6,2.8 5.5992,6.29922 6.6992,7.69922 2,2.6 2.2008,2.59961 11.8008,2.09961 L 1235,360.80078 V 357.5 c 0,-3.2 4e-4,-3.20039 6.9004,-4.40039 8.7,-1.4 14.6004,-1.4 20.9004,0 3.8,0.8 7.1992,0.60039 14.6992,-0.59961 15,-2.4 14.4,-0.5996 -1.5,4.40039 -15.7,4.9 -14.3992,4.79922 -13.6992,1.19922 l 0.5996,-3.09961 -9.8008,5 c -5.4,2.7 -12.6992,5.69961 -16.1992,6.59961 -16,4.1 -40.5004,-3.19962 -52.4004,-15.59961 l -2.9004,-3 2,-2.19922 c 14.277,-16.30578 25.6138,-27.57959 34.4004,-34.2168 z M 954.52539,314.125 c 0.9,-0.175 1.275,0.275 0.875,1.375 -0.8,1.9 -3.59961,3.5 -6.09961,3.5 -1.4,0 -1.20078,-0.5 1.19922,-2.5 1.7,-1.4 3.12539,-2.2 4.02539,-2.375 z M 1132.8008,315 c 0.4,0 1.7,1.6 3,3.5 l 2.2988,3.59961 -2.7988,2.40039 c -3.9,3.4 -5.0016,3.09921 -4.6016,-1.30078 0.6,-5.4 1.2016,-8.19922 2.1016,-8.19922 z M 1266.5,337 c 0.6,0 2.1008,1.59961 3.3008,3.59961 2.3,3.5 2.4988,3.60117 7.2988,3.20117 4.8,-0.5 6.7012,0.49961 3.2012,1.59961 -6.3,2 -29.3008,2.49922 -29.3008,0.69922 0,-1.9 12.3,-9.09961 15.5,-9.09961 z M 292,350 c 1.3,0 2,0.7 2,2 0,2.4 -2.59922,3.79922 -4.19922,2.19922 C 288.20078,352.59922 289.6,350 292,350 Z m 90.36328,6.21289 c 3.2625,-0.0875 6.63751,0.18711 10.9375,0.78711 9.09999,1.3 22.89884,6.59962 32.79883,12.59961 C 433.2996,373.9996 449,387.10078 449,388.80078 c 0,0.4 -4.20079,0.79883 -9.30078,0.79883 C 430.99923,389.79961 416.99999,392.4 411,395 c -2.4,1 -2.4,1.00039 0.5,0.40039 1.6,-0.3 2.59922,-0.20078 2.19922,0.19922 -0.5,0.5 -2.7,0.9 -5,1 -2.3,0.1 -6.69922,0.70078 -9.69922,1.30078 -3,0.7 -10.70001,1.19922 -17,1.19922 -25.09997,-0.1 -41.60001,-6.09884 -49,-17.79883 l -2.90039,-4.5 2.70117,-3.70117 c 1.5,-2.1 3.79961,-5.99883 5.09961,-8.79883 l 2.29883,-5 7.60156,-0.70117 c 4.1,-0.3 7.69844,-0.49883 7.89844,-0.29883 0.1,0.2 -1.6,2.39883 -4,4.79883 -4.9,5.09999 -12.29961,20.70039 -10.59961,22.40039 0.6,0.5 6.1004,1 12.40039,1 h 11.30078 l 3.09961,-4.5 c 1.7,-2.5 3.69961,-6.19922 4.59961,-8.19922 1.4,-3.4 1.40039,-4.2 -0.0996,-8 -2.8,-7.69999 -2.80078,-7.70117 -0.30078,-8.20117 3.85,-0.85 7.00117,-1.29922 10.26367,-1.38672 z m -101.11719,1.08984 c 0.29688,-0.14687 0.97969,-0.0531 2.05469,0.29688 1.2,0.3 6.69923,2.89961 12.19922,5.59961 8.59999,4.19999 11.29923,5 19.69922,6 8.89999,1.1 9.70078,1.40117 10.80078,3.70117 2.2,4.7 6,16.40039 5.5,16.90039 -0.7,0.8 -16.1004,-2.70078 -21.40039,-4.80078 -3.6,-1.4 -5.59922,-3.09961 -9.19922,-8.09961 -2.8,-3.9 -7.50118,-8.60117 -12.20117,-12.20117 -5.775,-4.35 -8.34375,-6.95586 -7.45313,-7.39649 z M 1257,362 c 0.3,0 1.8992,2.40078 3.6992,5.30078 2.3,3.8 2.8008,5.3 1.8008,5.5 -1.6,0.4 -12.6008,-1.50078 -13.3008,-2.30078 -0.6,-0.6 6.8008,-8.5 7.8008,-8.5 z m -87.6992,2.46289 c 4.4,-0.4875 8.25,0.0867 12.5,1.63672 23.4,8.69999 34.3,10.7 58,11 L 1252,377.19922 V 383.5 c 0,6.69999 1.8004,13.0004 4.9004,17.40039 l 2,2.69922 -9,6.30078 c -10.7,7.59999 -20.4,15.79885 -41.5,35.29883 -8.7,8.09999 -18,16.30078 -20.5,18.30078 -12.6,9.69999 -31.7996,17.7 -50.5996,21 -10.6,1.9 -37.8008,2.2 -50.8008,0.5 -4.9,-0.6 -18,-2.9 -29,-5 -32.5,-6.29999 -46.4,-8.09922 -66.5,-8.69922 -17.49998,-0.5 -31.4004,0.69883 -40.90039,3.29883 -3.8,1.1 -1.69882,-1.10039 4.20117,-4.40039 2.8,-1.6 8.59922,-4.79922 12.69922,-7.19922 4.1,-2.3 17.60001,-9.59923 30,-16.19922 C 1009.4,440.10079 1022.4,433 1026,431 c 11.7,-6.79999 40.6992,-20.89961 48.1992,-23.59961 C 1098.0992,399.0004 1107.1,397.3 1132,396.5 c 9.6,-0.3 19.1,-0.89922 21,-1.19922 l 3.5,-0.60156 -3.5,-1 c -1.9,-0.5 -13.6,-1.19961 -26,-1.59961 L 1104.5996,391.5 1113,387.30078 c 7.5,-3.79999 21.4,-10.00157 40.5,-18.10156 6.45,-2.7 11.4008,-4.24883 15.8008,-4.73633 z m -586.81252,14.0625 c -0.4375,0.225 -0.73828,0.72422 -0.98828,1.57422 -1.6,4.89999 -0.69922,13.70079 1.80078,19.30078 2.4,5.3 2.79961,5.59961 6.09961,5.59961 3.7,0 5.99922,-1.6 9.19922,-6.5 1,-1.5 2.60117,-3.00039 3.70117,-3.40039 1.8,-0.7 1.39961,-1.39883 -3.90039,-6.79883 -5.59999,-5.69999 -9.29961,-8.20117 -14.09961,-9.70117 -0.8,-0.25 -1.375,-0.29922 -1.8125,-0.0742 z m -288.71289,8.53711 c 1.125,-0.0125 2.47461,0.43789 3.72461,1.33789 2.4,1.7 1.29922,4.59961 -1.80078,4.59961 -2.9,0 -5.19883,-2.4 -4.29883,-4.5 0.35,-0.95 1.25,-1.425 2.375,-1.4375 z m 18.86133,7.07422 c 1.2125,-0.1875 3.81329,0.86289 8.86328,2.96289 4.1,1.8 7.5,3.50078 7.5,3.80078 0,0.4 -2,1.69961 -4.5,3.09961 -3.5,1.9 -4.59961,3.19961 -5.09961,5.59961 -0.5,2.7 -1.30118,3.50117 -5.70117,5.20117 -4.5,1.6 -5.4,2.39922 -6.5,5.69922 -1.6,4.8 -0.99961,6.5 2.40039,6.5 1.7,0 3.99961,-1.3 6.59961,-4 2.2,-2.2 4.3,-4 4.5,-4 0.3,0 1.40039,1.10039 2.40039,2.40039 4,5.1 15.19961,7.00039 19.59961,3.40039 2.9,-2.4 6.30078,-11.10118 6.30078,-16.20117 0,-3.8 0.30039,-4.69883 1.40039,-4.29883 0.8,0.3 6.90001,1.19844 13.5,1.89844 6.69999,0.7 12.09961,1.50039 12.09961,1.90039 0,1.4 -21.69922,9.90039 -25.69922,9.90039 -1.6,0.1 -1.60078,0.19922 0.19922,1.19922 3.1,1.9 16.0004,1.30039 25.90039,-1.09961 4.9,-1.2 9.49922,-1.99922 10.19922,-1.69922 1.7,0.6 2.00078,9.59961 0.30078,9.59961 -0.5,0 -3.2,-0.69961 -6,-1.59961 -2.7,-0.9 -6.00078,-1.4 -7.30078,-1 -4.8,1.2 -17.49961,21.49923 -17.59961,28.19922 0,1.8 0.89962,2.10117 9.59961,3.20117 8.29999,0.9 10.20078,0.9 13.80078,-0.5 11.69999,-4.59999 14.99961,-12.1004 11.09961,-24.90039 -2.5,-8.29999 -2.00039,-11.80117 2.09961,-15.20117 6.99999,-5.8 28.9004,-11.09922 39.90039,-9.69922 5.99999,0.7 16.1,3.69922 21,6.19922 8.89999,4.59999 17.89964,12.60003 46.59961,41.5 C 517.49959,469.6992 532.40078,484 533.30078,484 c 0.9,0 1.69922,0.4 1.69922,1 0,1.3 0.59998,1.40039 -17,-2.09961 C 500.00002,479.40039 478.80075,476.9 442.80078,474 410.80081,471.5 393.49959,468.9996 376.59961,464.59961 352.79963,458.29962 338.09998,452.39999 322,442.5 c -16.49998,-10.09999 -26.30001,-15.39961 -31.5,-17.09961 -11.59999,-3.8 -21.09962,-9.10118 -27.59961,-15.20117 l -6.40039,-6 8.09961,0.5 8.20117,0.60156 L 282,400 c 5,-2.9 9.3,-5.00078 9.5,-4.80078 0.3,0.2 -2.9,4.10157 -7,8.60156 -10.79999,11.89999 -10.79921,11.79961 -2.69922,12.59961 3.7,0.3 9.69883,0.59961 13.29883,0.59961 h 6.70117 l 5.59961,-6.90039 c 5.6,-6.79999 5.59922,-6.80001 4.69922,-11.5 -0.5,-2.85 -0.67539,-4.27539 0.53711,-4.46289 z m -126.54102,5.14062 c -2.30156,-0.40781 -5.54571,0.54688 -15.5957,3.92188 -6.59999,2.3 -12.60078,4.20039 -13.30078,4.40039 -0.9,0.3 0.001,2.30078 2.70117,6.30078 C 165.90038,422.60038 167.99923,424 175.69922,424 186.89921,424 192,419.29999 192,409 c 0,-6.09999 -0.49922,-7.20039 -3.69922,-8.90039 -0.775,-0.4 -1.43789,-0.68633 -2.20508,-0.82227 z m 1202.9629,1.39844 c 1.6172,-0.18281 2.7414,2.12423 5.4414,9.32422 1.5,4.1 3.1996,7.90039 3.5996,8.40039 2.9,3.2 -1.3004,7.7 -8.9004,9.5 -4.8,1.2 -25.4,3.19961 -26,2.59961 -0.2,-0.1 0.5,-3.19961 1.5,-6.59961 l 1.9004,-6.40039 6.9004,-0.5 6.9004,-0.5 0.7988,-3.80078 c 1,-5 3.3,-9.19961 6,-11.09961 0.725,-0.525 1.3203,-0.86289 1.8594,-0.92383 z M 1452.8008,417.5 c 0.2,0.2 -0.9,2.60039 -2.5,5.40039 -1.5,2.8 -3.5008,5.09961 -4.3008,5.09961 -1.7,0 -6.7,-3.60078 -6,-4.30078 1.1,-1 12.5008,-6.49922 12.8008,-6.19922 z m -51.2012,4.59961 c 0.6,-0.1 2.1008,2.10078 3.3008,4.80078 1.1,2.7 2.0996,5.2 2.0996,5.5 0,0.3 -3.1,0.59961 -7,0.59961 -4.6,0 -7,-0.39961 -7,-1.09961 0,-1.6 7.1996,-9.70078 8.5996,-9.80078 z M 1481,432.19922 c 1.9,0 2.7008,0.2 1.8008,0.5 -1,0.2 -2.6,0.2 -3.5,0 -1,-0.3 -0.2008,-0.5 1.6992,-0.5 z m -41.5,6.30078 0.3008,3.80078 c 0.2,2.2 0.099,4.49883 -0.2012,5.29883 -0.4,1.1 -1.3988,1.00078 -5.2988,-0.69922 -2.6,-1.2 -5.9008,-2.49961 -7.3008,-3.09961 l -2.5,-1 2.9004,-1.5 c 1.6,-0.8 4.9,-1.80117 7.5,-2.20117 z m -37,3.5 c 5.2,0 9.5,0.29922 9.5,0.69922 -0.1,1.3 -7.2,6.30078 -9,6.30078 -1.8,0 -10,-5.09922 -10,-6.19922 0,-0.5 4.3,-0.80078 9.5,-0.80078 z m -28.6992,1 c 0.5,0 2.8992,0.7 5.1992,1.5 5.1,1.8 15.0992,7.80079 19.6992,11.80078 3.9,3.4 4.2016,6.59962 1.6016,14.59961 -1.9,5.89999 -2.9016,6.20039 -10.6016,2.90039 -5.4,-2.4 -9.6988,-2.80156 -16.2988,-1.60156 l -4.2012,0.80078 -0.3984,-4.59961 c -0.3,-4 0.1984,-5.40118 3.3984,-10.20117 3.8,-5.8 4.5004,-7.99883 2.9004,-9.79883 -1.4,-1.4 -2.2988,-5.40039 -1.2988,-5.40039 z m -28.625,1.09961 c 1.075,-0.05 1.8242,0.14961 1.8242,0.59961 0,0.3 -1.0996,2.1 -2.5996,4 -2.4,3.2 -6.5008,4.5 -7.8008,2.5 -0.8,-1.3 2.3008,-5.19922 4.8008,-6.19922 1.3,-0.55 2.7004,-0.85039 3.7754,-0.90039 z m 100.3242,4.20117 c 0.5,0.2 0.9008,2.09883 0.8008,4.29883 -0.4,5.49999 -0.7012,7.40039 -1.2012,6.90039 -0.8,-0.9 -6.1,3.69922 -5.5,4.69922 0.4,0.6 0.2012,0.8 -0.2988,0.5 -0.6,-0.4 -2.9012,7.8e-4 -5.2012,0.80078 -2.6,0.9 -4.0996,1.00039 -4.0996,0.40039 0,-2.9 13.5,-18.19961 15.5,-17.59961 z M 264.69922,459 c 0.5,0 4.80078,1.4 9.80078,3 8.99999,3 12.4,4.60078 11.5,5.30078 -0.3,0.2 -4.1,2.4 -8.5,5 -9.69999,5.6 -23.09922,11.09961 -24.69922,10.09961 -0.7,-0.5 0.39961,-1.29961 2.59961,-2.09961 2.1,-0.7 4.19961,-1.10078 4.59961,-0.80078 1.7,1.1 4,-6.5004 4,-13.40039 0,-3.9 0.29922,-7.09961 0.69922,-7.09961 z m 1070.37498,1.19922 c 1.675,0.05 2.275,0.55 2.625,1.5 1.7,4.69999 7.4016,12.10078 11.1016,14.30078 2.3,1.4 4.1992,2.90039 4.1992,3.40039 0,2.1 -12.0996,-0.40039 -22.0996,-4.40039 -8.3,-3.4 -12.2004,-4.40039 -18.9004,-4.90039 l -8.5,-0.59961 4,-2.09961 c 5.3,-2.8 10.6,-4.5 18.5,-6 4.65,-0.85 7.3992,-1.25117 9.0742,-1.20117 z m -562.47459,7.10156 c 0.3,0.3 0.39961,1.19844 0.0996,1.89844 -0.3,0.8 -0.59961,0.50039 -0.59961,-0.59961 -0.1,-1.1 0.2,-1.69883 0.5,-1.29883 z m 651.52539,0.76172 c 1.05,0.0375 2.0754,0.48789 2.7754,1.33789 1.9,2.3 -2.2004,5.99961 -4.9004,4.59961 -2.3,-1.2 -2.6008,-3.00078 -0.8008,-4.80078 0.8,-0.8 1.8758,-1.17422 2.9258,-1.13672 z M 183.30078,469 h 8.39844 l 9.10156,6.5 c 5,3.6 9.39883,6.99961 9.79883,7.59961 1.1,1.8 -9.90001,0.20039 -18,-2.59961 C 185.59962,478.1 175,471.8 175,470 c 0,-0.6 3.50079,-1 8.30078,-1 z m 1021.09962,15.09961 c 1.1,-0.1 1.6988,0.2 1.2988,0.5 -0.3,0.3 -1.1984,0.39961 -1.8984,0.0996 -0.8,-0.3 -0.5004,-0.59961 0.5996,-0.59961 z M 184.40039,485 c 0.2,0 2.29922,0.69961 4.69922,1.59961 7.79999,2.7 17.80118,4.40039 29.70117,4.90039 6.39999,0.3 11.49883,0.80078 11.29883,1.30078 -0.3,0.4 -3.40039,2.99844 -6.90039,5.89844 C 214.59923,505.49921 205.70039,510 200.90039,510 H 197 v 9.5 l -5.30078,1.69922 c -3,1 -5.79844,1.80078 -6.39844,1.80078 -0.9,0 -6.80118,-7.20079 -14.20117,-17.30078 -1.3,-1.8 -1.10039,-1.89922 5.09961,-1.19922 l 6.5,0.69922 0.60156,-3.79883 c 0.4,-2.2 0.69922,-6.70117 0.69922,-10.20117 0,-3.4 0.20039,-6.19922 0.40039,-6.19922 z m 1044.48631,12.34961 c 5.4125,0.2 9.5625,1.29961 13.3125,3.34961 8.8,4.69999 17.2,14.30041 28.5,32.40039 11,17.49998 13.9012,21.8004 24.7012,35.40039 3.7,4.7 6.6,8.80039 6.5,8.90039 -0.2,0.2 -4.7004,-0.3 -9.9004,-1 -5.2,-0.7 -15.3,-2.00039 -22.5,-2.90039 -25.1,-3.1 -35.3,-6.10079 -48,-14.30078 -7.4,-4.7 -10.8,-7.39923 -27.5,-21.69922 -12.3,-10.49999 -23.3996,-18.00039 -25.0996,-16.90039 -0.8,0.5 -0.9004,0.30039 -0.4004,-0.59961 0.6,-1 0.3992,-1.19922 -0.8008,-0.69922 -0.9,0.3 -1.8988,0.1 -2.2988,-0.5 -0.5,-0.7 -0.2004,-0.80078 0.5996,-0.30078 0.9,0.6 1.0996,0.39961 0.5996,-0.40039 -0.5,-0.7 -1.4988,-0.99922 -2.2988,-0.69922 -0.8,0.3 -1.1008,0.0996 -0.8008,-0.40039 0.3,-0.5 -3.1992,-2.79961 -7.6992,-5.09961 -8.1,-4.1 -8.2008,-4.3 -5.3008,-5 1.7,-0.4 12.2,-1.80039 23.5,-2.90039 11.3,-1.2 26.6,-3.29961 34,-4.59961 8.8,-1.55 15.4742,-2.25078 20.8867,-2.05078 z M 295.59961,499 c 0.3,0 2.20117,1.59961 4.20117,3.59961 4.2,4.1 11.39923,9.00117 18.69922,12.70117 2.8,1.5 4.8,2.69922 4.5,2.69922 -0.3,0.1 -5.20001,0.39961 -11,0.59961 -5.79999,0.3 -14.1,1.2 -18.5,2 -7.79999,1.3 -8.09922,1.30078 -9.69922,-0.69922 -1,-1.1 -2.20078,-3.70117 -2.80078,-5.70117 -1.1,-4.2 -0.60078,-4.79922 5.19922,-5.69922 3.2,-0.6 4.30039,-1.39961 6.40039,-5.09961 1.4,-2.4 2.8,-4.40039 3,-4.40039 z m 895.80079,1.09961 c 1.1,-0.1 1.6988,0.2 1.2988,0.5 -0.3,0.3 -1.1984,0.39961 -1.8984,0.0996 -0.8,-0.3 -0.5004,-0.59961 0.5996,-0.59961 z m -110.2129,12 c 1.6875,0.175 4.5625,0.65 9.3125,1.5 10.9,1.9 40.2008,9.40078 50.3008,12.80078 19.6,6.69999 31.8,16.40041 39.5,31.40039 5.5,10.59999 16.5984,46.59845 17.3984,56.39844 0.5,5.89999 0.3004,7.30118 -2.0996,12.20117 -2.6,5.4 -2.9004,5.6 -9.4004,7.5 -13.9,4.2 -20.6,5.99961 -21,5.59961 -0.3,-0.2 -1.2988,-4.70001 -2.2988,-10 -2.8,-13.89999 -7.3996,-25.80079 -13.5996,-34.80078 -2.9,-4.3 -5.3008,-7.99883 -5.3008,-8.29883 0,-0.2 2.4008,-1.50117 5.3008,-2.70117 l 5.1992,-2.39844 -8.0996,-0.60156 c -4.5,-0.3 -10.7004,-1.59883 -13.9004,-2.79883 -8,-3 -17.2996,-12.49962 -27.0996,-27.59961 -4.2,-6.49999 -10.3,-15.90078 -13.5,-20.80078 -3.2,-5 -7.5004,-10.70078 -9.4004,-12.80078 -1.9,-2.1 -3.5,-3.99883 -3.5,-4.29883 0,-0.35 0.5,-0.47578 2.1875,-0.30078 z m -790.6875,12.5 c 0.6,-0.2 2.4,-0.2 4,0 1.7,0.2 10.40001,0.60039 19.5,0.90039 9.09999,0.4 25.70001,1.20039 37,1.90039 l 20.5,1.19922 -6,2.80078 c -20.19998,9.39999 -47.60082,13.3 -84.80078,12 l -18.79883,-0.70117 2.29883,-2.39844 c 1.2,-1.3 4.8,-3.30039 8,-4.40039 3.2,-1.2 5.80078,-2.60117 5.80078,-3.20117 0,-0.7 -0.39922,-0.69844 -1.19922,0.10156 -0.7,0.7 -1.60117,1.19922 -2.20117,1.19922 -0.5,0 -0.0996,-0.60039 0.90039,-1.40039 1.1,-0.8 2.29922,-1.2 2.69922,-1 0.4,0.3 1.40078,-0.9 2.30078,-2.5 1.1,-2.3 2.40078,-3.19961 5.30078,-3.59961 2,-0.4 4.19922,-0.80039 4.69922,-0.90039 z M 1458.1992,528 1455.5,532.69922 c -1.5,2.7 -3.7004,5.80156 -4.9004,7.10156 l -2.1992,2.29883 -3.3008,-4.59961 c -2.9,-4 -3.5992,-4.5 -7.1992,-4.5 -5.2,0 -4.9996,-1.6 0.4004,-3 2.3,-0.6 4.6992,-1.3 5.1992,-1.5 0.6,-0.2 4.1008,-0.40039 7.8008,-0.40039 z M 218.125,544.13672 c -0.225,0.0375 -0.42461,0.41406 -0.72461,1.16406 -0.8,2.1 0.19961,3.49961 1.59961,2.09961 0.5,-0.5 0.6,-1.50039 0,-2.40039 -0.4,-0.6 -0.65,-0.90078 -0.875,-0.86328 z m 1156.7754,9.26367 c 0.5,-0.5 2.0996,-0.5 2.0996,0 0,0.2 -1.1992,2.49922 -2.6992,5.19922 -1.5,2.7 -3.0008,6.70039 -3.3008,8.90039 l -0.5,4 -7.5996,-0.0996 c -6.2,-0.1 -8.7,-0.70117 -14,-3.20117 -5.7,-2.7 -15.0008,-10.49961 -13.8008,-11.59961 0.2,-0.2 9.2004,-0.89961 19.9004,-1.59961 10.7,-0.6 19.7004,-1.29961 19.9004,-1.59961 z m -992.87501,21.71289 c 0.675,0.0125 1.37422,0.43711 2.57422,1.28711 2.6,2 0.70078,5.09961 -3.19922,5.09961 -3.5,0 -4.60078,-3.19922 -1.80078,-5.19922 1.1,-0.8 1.75078,-1.2 2.42578,-1.1875 z M 1240,579.30078 c 0,-0.8 13.2992,0.29883 21.6992,1.79883 3.5,0.6 6.3008,1.40078 6.3008,1.80078 0,0.4 -2.2996,1.5 -5.0996,2.5 l -5,1.79883 0.2988,8.90039 c 0.2,4.89999 0.1012,8.90039 -0.2988,8.90039 -2.3,0 -17.9004,-22.29922 -17.9004,-25.69922 z M 1297.3008,585 c 8.4,0 10.9,0.4 14,2 2.1,1.1 4.2984,1.70039 4.8984,1.40039 0.6,-0.4 0.8,-0.30117 0.5,0.29883 -0.4,0.6 1.1008,2.30078 3.3008,3.80078 2.2,1.5 4,3.1 4,3.5 0,0.4 -2.2992,1.10039 -5.1992,1.40039 -8.1,1.1 -13.4016,4.79883 -12.6016,8.79883 0.2,0.6 -0.3,1.20078 -1,1.30078 -2.3,0 -4.0984,4.59922 -2.8984,7.19922 0.5,1.2 1.4,2.00156 2,1.60156 0.6,-0.3 0.6988,-0.10039 0.2988,0.59961 -0.4,0.6 0.4012,3.10039 1.7012,5.40039 5,8.39999 4.3,7.99922 13.5,6.69922 4.6,-0.7 8.4992,-1.10039 8.6992,-0.90039 0.2,0.2 -0.5,2.90039 -1.5,5.90039 -1.1,3 -1.9,6.99922 -2,8.69922 0,1.8 -0.3008,3.30078 -0.8008,3.30078 -1.3,0 -10.3996,-6.40079 -19.5996,-13.80078 -4.9,-4 -9.4992,-7.1 -10.1992,-7 -0.7,0.2 -1.3012,-0.0996 -1.2012,-0.59961 0,-0.4 -1.8996,-2.7 -4.0996,-5 l -4.1992,-4.29883 5.5,-4.70117 5.5,-4.69922 -3.3008,-4.40039 c -3.6,-4.7 -7.1996,-13.8 -6.0996,-15.5 0.3,-0.6 4.9008,-1 10.8008,-1 z m 43.2988,22.69922 c 0.2,-0.2 4.1012,0.5 8.7012,1.5 4.5,1 12.6,2.1 18,2.5 9.1,0.6 11.1996,1.30039 8.5996,2.90039 -0.6,0.3 -2.4,2.70117 -4,5.20117 -2.3,3.6 -2.9008,5.69961 -2.8008,9.59961 0.2,9.09999 0.6004,9.70039 9.9004,13.40039 4.7,1.8 8.9008,3.69922 9.3008,4.19922 0.5,0.4 -1.2012,2 -3.7012,3.5 L 1380,653.19922 1374.3008,648 c -5.8,-5.29999 -12.2016,-8.9 -15.6016,-9 -1.6,0 -1.7996,-0.80001 -1.5996,-7.5 0.2,-4.1 0.1008,-7.60039 -0.1992,-7.90039 -0.2,-0.3 -2.3,0.30078 -4.5,1.30078 -8.3,3.7 -8.0012,3.79921 -8.7012,-2.80078 -0.4,-3.3 -1.3,-7.8 -2,-10 -0.8,-2.2 -1.2996,-4.20039 -1.0996,-4.40039 z M 1394.9004,611 h 6.5 c 6.3,0 6.6988,0.19922 11.2988,4.19922 2.7,2.4 6.0008,5.00039 7.3008,5.90039 l 2.5,1.59961 -3,4.90039 C 1416.3,632.6996 1408.9992,639 1406.1992,639 c -0.9,0 -2.9992,-2.10078 -4.6992,-4.80078 -2.6,-4.1 -3.0004,-5.59923 -2.9004,-11.19922 0.1,-5 -0.2988,-7.09922 -1.7988,-9.19922 z M 371.07422,636.02539 c 1.6,-0.1 2.92578,0.37461 2.92578,1.47461 0,2.2 -3.7,5.5 -6,5.5 -2.8,0 -3.6,-2.7 -1.5,-5 1.1,-1.2 2.97422,-1.87461 4.57422,-1.97461 z M 1411.1992,664 c 0.9,0 2.2008,0.9 2.8008,2 2.2,4 -2.6008,8.00078 -5.8008,4.80078 -2.5,-2.5 -0.6,-6.80078 3,-6.80078 z M 418.09961,827 c 0.3,0 2.30039,3.49922 4.40039,7.69922 4.1,8.29999 5.89961,19.30156 3.59961,21.60156 -1.6,1.6 -1.99922,0.69843 -2.19922,-6.10156 l -0.20117,-6.09961 -2.89844,2.20117 c -1.7,1.2 -3.9,2.69883 -5,3.29883 -1.1,0.7 -3.50039,2.7 -5.40039,4.5 L 407,857.40039 v -5.09961 c 0,-3.5 0.6,-5.80156 2,-7.60156 1.1,-1.4 2,-2.99961 2,-3.59961 0,-1.5 6.29961,-14.09961 7.09961,-14.09961 z m 838.30079,50.09961 c 1.1,-0.1 1.6988,0.2 1.2988,0.5 -0.3,0.3 -1.1984,0.39961 -1.8984,0.0996 -0.8,-0.3 -0.5004,-0.59961 0.5996,-0.59961 z"
/>`, tx, sx, sy)
}
