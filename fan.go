package gtree

import (
	"fmt"
	"math"
)

// FanChart defines the top-level structure for a radial ancestor chart.
// It represents a single subject (Root) and their lineage going backward.
// The chart includes a title, optional notes, and a recursive ancestry tree.
type FanChart struct {
	Title string     // Title displayed on the diagram.
	Notes []string   // Optional lines of annotation, shown at the edge of the layout.
	Root  *FanPerson // The subject of the chart (generation 0). All ancestors trace back from this person.
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
	NoteStyle      TextStyleOption // NoteStyle is the style of the font to use for the notes of the chart.
	HeadingStyle   TextStyleOption // HeadingStyle is the style of the font to use for the headling lines of each node.
	DetailStyle    TextStyleOption // DetailStyle is the style of the font to use for the subsequent lines of each node after the first.
	DPI            int

	titleStyle     TextStyle
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
		Margin:         15,
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
		NoteStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   28,
			LineHeight: 28,
			Color:      "#000",
		},
		HeadingStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   28,
			LineHeight: 30,
			Color:      "#000",
		},
		DetailStyle: TextStyleOption{
			FontNames:  DefaultSerifFontNames(),
			FontSize:   16,
			LineHeight: 18,
			Color:      "#000",
		},
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

	legend := CenterAlignedLegend(
		Point{X: rootBlurb.X, Y: height - margin},
		chart.Title,
		opts.titleStyle,
		chart.Notes,
		opts.noteStyle)
	legend.Y = height - margin

	// Ensure there is enough room between root person and title
	base := rootBlurb.X + rootBlurb.Height/2
	gap := legend.Y - base - opts.titleStyle.LineHeight
	if gap < 0 {
		legend.Y += -gap
		height += -gap
	}

	layout := &FanLayout{
		legend:     legend,
		blurbs:     blurbs,
		connectors: connectors,
		extent:     Extent{Width: width, Height: height},
		margin:     margin,
		debug:      false,
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

func radiansToDegrees(rad float64) float64 {
	return rad * 180.0 / math.Pi
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
	legend     *Blurb
	blurbs     []*Blurb
	connectors []*Connector
	// width      Pixel
	// height     Pixel
	margin Pixel
	debug  bool
	extent Extent
}

func (f *FanLayout) Height() Pixel            { return f.extent.Height }
func (f *FanLayout) Width() Pixel             { return f.extent.Width }
func (f *FanLayout) Margin() Pixel            { return f.margin }
func (f *FanLayout) Legend() *Blurb           { return f.legend }
func (f *FanLayout) Blurbs() []*Blurb         { return f.blurbs }
func (f *FanLayout) Connectors() []*Connector { return f.connectors }
func (f *FanLayout) Debug() bool              { return f.debug }
func (f *FanLayout) Dimensions() Extent       { return f.extent }
func (f *FanLayout) SetDimensions(e Extent) {
	f.extent = e
}

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

// adjustConnector trims or extends the start and end points of a connector by given distances.
// Positive values extend the connector, negative values trim it.
func adjustConnector(conn *Connector, deltaStart, deltaEnd Pixel) {
	if len(conn.Points) < 2 {
		return
	}

	// Start adjustment (keep end fixed, move start)
	if deltaStart != 0 {
		p0 := conn.Points[0]
		p1 := conn.Points[1]
		dx := float64(p1.X - p0.X)
		dy := float64(p1.Y - p0.Y)
		length := math.Hypot(dx, dy)
		if length != 0 {
			ux := dx / length
			uy := dy / length
			conn.Points[0] = Point{
				X: Pixel(float64(p0.X) + float64(deltaStart)*ux),
				Y: Pixel(float64(p0.Y) + float64(deltaStart)*uy),
			}
		}
	}

	// End adjustment (keep start fixed, move end)
	if deltaEnd != 0 {
		n := len(conn.Points)
		p0 := conn.Points[n-2]
		p1 := conn.Points[n-1]
		dx := float64(p1.X - p0.X)
		dy := float64(p1.Y - p0.Y)
		length := math.Hypot(dx, dy)
		if length != 0 {
			ux := dx / length
			uy := dy / length
			conn.Points[n-1] = Point{
				X: Pixel(float64(p1.X) + float64(deltaEnd)*ux),
				Y: Pixel(float64(p1.Y) + float64(deltaEnd)*uy),
			}
		}
	}
}
