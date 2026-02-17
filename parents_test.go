package gtree

import (
	"testing"
)

func TestParentsLayout(t *testing.T) {
	ch := &DescendantChart{
		Root: &DescendantPerson{
			ID:       10,
			Headings: []string{"Joseph Chambers"},
			Details:  []string{"1738-1815"},
			Families: []*DescendantFamily{
				{
					Other: &DescendantPerson{
						ID:       11,
						Headings: []string{"Sarah Meen"},
						Details:  []string{"1751-1820"},
					},
					Children: []*DescendantPerson{
						{
							ID:       0,
							Headings: []string{"James Chambers"},
							Details:  []string{"1794-1854"},
							Families: []*DescendantFamily{
								{
									Other: &DescendantPerson{
										ID:       1,
										Headings: []string{"Mary Ann Martin"},
										Details:  []string{"1801-1873"},
									},
									Details: []string{"1821"},
									Children: []*DescendantPerson{
										{ID: 2, Headings: []string{"John Chambers"}, Details: []string{"1822-1901"}},
										{ID: 3, Headings: []string{"Robert Chambers"}, Details: []string{"1824-?"}},
										{ID: 4, Headings: []string{"James Chambers"}, Details: []string{"1827-?"}},
										{ID: 5, Headings: []string{"Mary Anne Chambers"}, Details: []string{"1830-?"}},
										{
											ID:       6,
											Headings: []string{"William Chambers"},
											Details:  []string{"1833-1879"},
											Families: []*DescendantFamily{
												{
													Other: &DescendantPerson{
														ID:       7,
														Headings: []string{"Rebecca Brooks"},
														Details:  []string{"1841-1881"},
													},
													Details: []string{"1860"},
												},
											},
										},
										{ID: 8, Headings: []string{"George Chambers"}, Details: []string{"1836-?"}},
										{ID: 9, Headings: []string{"Emma Maria Chambers"}, Details: []string{"1838-?"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	l, err := ch.Layout(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The relationship marker between the parents should be centered over their child
	rel := l.nodes[-11]
	james := l.nodes[0]

	if rel == nil {
		t.Fatal("rel node -11 not found")
	}
	if james == nil {
		t.Fatal("james node 0 not found")
	}

	relCenter := rel.Left() + rel.Width/2
	jamesCenter := james.Left() + james.Width/2

	diff := relCenter - jamesCenter
	if diff < 0 {
		diff = -diff
	}
	if diff > 50 {
		t.Errorf("parent row not centered over child: rel center=%d, james center=%d, diff=%d", relCenter, jamesCenter, diff)
	}
}
