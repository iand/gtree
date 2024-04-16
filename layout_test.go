package gtree

import "testing"

var (
	onePerson = &DescendantChart{
		Root: &DescendantPerson{
			ID:      1,
			Details: []string{"Person One"},
		},
	}

	onePersonWithSpouse = &DescendantChart{
		Root: &DescendantPerson{
			ID:      1,
			Details: []string{"Person One"},
			Families: []*DescendantFamily{
				{
					Other: &DescendantPerson{
						ID:      2,
						Details: []string{"Person Two"},
					},
				},
			},
		},
	}

	onePersonWithThreeSpouses = &DescendantChart{
		Root: &DescendantPerson{
			ID:      1,
			Details: []string{"Person One"},
			Families: []*DescendantFamily{
				{
					Other: &DescendantPerson{
						ID:      2,
						Details: []string{"Person Two"},
					},
				},
				{
					Other: &DescendantPerson{
						ID:      3,
						Details: []string{"Person Three"},
					},
				},
				{
					Other: &DescendantPerson{
						ID:      4,
						Details: []string{"Person Four"},
					},
				},
			},
		},
	}

	onePersonWithSpouseAndChildren = &DescendantChart{
		Root: &DescendantPerson{
			ID:      1,
			Details: []string{"Person One"},
			Families: []*DescendantFamily{
				{
					Other: &DescendantPerson{
						ID:      2,
						Details: []string{"Person Two"},
					},
					Children: []*DescendantPerson{
						{
							ID:      3,
							Details: []string{"Person Three"},
						},
						{
							ID:      4,
							Details: []string{"Person Four"},
						},
					},
				},
			},
		},
	}
)

func TestVerticalLayout(t *testing.T) {
	testCases := []struct {
		name       string
		in         *DescendantChart
		assertions []layoutAssertion
	}{
		{
			name: "one person",
			in:   onePerson,
			assertions: []layoutAssertion{
				blurb(1).
					hasText("Person One").
					hasNoParent().
					hasNoLeftNeighbour().
					inRow(0),
			},
		},
		{
			name: "one person with spouse",
			in:   onePersonWithSpouse,
			assertions: []layoutAssertion{
				blurb(1).
					hasText("Person One").
					hasNoParent().
					hasNoLeftNeighbour().
					hasKeepWith(-2).
					inRow(0),
				blurb(-2).
					hasText("=").
					hasNoParent().
					hasLeftNeighbour(1).
					hasKeepWith(1).
					hasKeepWith(2).
					inRow(0),
				blurb(2).
					hasText("Person Two").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-2).
					hasKeepWith(-2).
					inRow(0),
			},
		},
		{
			name: "one person with three spouses",
			in:   onePersonWithThreeSpouses,
			assertions: []layoutAssertion{
				blurb(1).
					hasText("Person One").
					hasNoParent().
					hasNoLeftNeighbour().
					hasKeepWith(-2).
					inRow(0),
				blurb(-2).
					hasText("= (1)").
					hasNoParent().
					hasLeftNeighbour(1).
					hasKeepWith(1).
					hasKeepWith(2).
					inRow(0),
				blurb(2).
					hasText("Person Two").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-2).
					hasKeepWith(-2).
					inRow(0),
				blurb(-3).
					hasText("= (2)").
					hasNoParent().
					hasLeftNeighbour(2).
					hasKeepWith(1).
					hasKeepWith(3).
					inRow(0),
				blurb(3).
					hasText("Person Three").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-3).
					hasKeepWith(-3).
					inRow(0),
				blurb(-4).
					hasText("= (3)").
					hasNoParent().
					hasLeftNeighbour(3).
					hasKeepWith(1).
					hasKeepWith(4).
					inRow(0),
				blurb(4).
					hasText("Person Four").
					hasNoShift().
					hasLeftNeighbour(-4).
					hasKeepWith(-4).
					inRow(0),
			},
		},
		{
			name: "one person with spouse and children",
			in:   onePersonWithSpouseAndChildren,
			assertions: []layoutAssertion{
				blurb(1).
					hasText("Person One").
					hasNoParent().
					hasNoLeftNeighbour().
					hasKeepWith(-2).
					hasLeftStop(3).
					hasRightStop(4).
					inRow(0),
				blurb(-2).
					hasText("=").
					hasNoParent().
					hasLeftNeighbour(1).
					hasKeepWith(1).
					hasKeepWith(2).
					inRow(0),
				blurb(2).
					hasText("Person Two").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-2).
					hasKeepWith(-2).
					hasLeftStop(3).
					inRow(0),
				blurb(3).
					hasText("Person Three").
					hasParent(-2).
					hasNoLeftNeighbour().
					hasKeepWith(-2).
					inRow(1),
				blurb(4).
					hasText("Person Four").
					inRow(1).
					hasParent(-2).
					hasKeepWith(-2).
					hasLeftNeighbour(3),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			l := tc.in.Layout(nil)

			for _, a := range tc.assertions {
				a.assert(t, l)
			}
		})
	}
}

type layoutAssertion interface {
	assert(*testing.T, *DescendantLayout)
}

func blurb(id int) *blurbAsserter {
	return &blurbAsserter{id: id}
}

type blurbAsserter struct {
	id  int
	fns []func(t *testing.T, b *Blurb, l *DescendantLayout)
}

func (a *blurbAsserter) assert(t *testing.T, l *DescendantLayout) {
	b, ok := l.blurbs[a.id]
	if !ok {
		t.Errorf("blurb %d is missing", a.id)
		return
	}

	for _, fn := range a.fns {
		fn(t, b, l)
	}
}

func (ba *blurbAsserter) hasText(texts ...string) *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if len(b.DetailTexts) != len(texts)-1 {
			t.Fatalf("blurb %d: got %d detail texts, wanted %d", ba.id, len(b.DetailTexts), len(texts)-1)
		}

		for i := range texts {
			if i == 0 {
				if b.HeadingText != texts[i] {
					t.Errorf("blurb %d: got heading text %q, wanted %q", ba.id, b.HeadingText, texts[i])
				}
				continue
			}
			if b.DetailTexts[i-1] != texts[i] {
				t.Errorf("blurb %d: got detail text %q, wanted %q", ba.id, b.DetailTexts[i-1], texts[i])
			}
		}
	})
	return ba
}

func (ba *blurbAsserter) inRow(row int) *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if b.Row != row {
			t.Errorf("blurb %d: got row %d, wanted %d", b.ID, b.Row, row)
		}

		if len(l.rows) <= row {
			t.Errorf("blurb %d: layout didn't have row %d", ba.id, row)
			return
		}

		if len(l.rows[row]) == 0 {
			t.Errorf("blurb %d: layout didn't have any blurbs in row %d", ba.id, row)
			return

		}

		for _, b2 := range l.rows[row] {
			if b2.ID == b.ID {
				return
			}
		}
		t.Errorf("blurb %d: missing from layout row %d", ba.id, row)
	})
	return ba
}

func (ba *blurbAsserter) hasLeftNeighbour(id int) *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if b.LeftNeighbour == nil {
			t.Errorf("blurb %d: got no left neighbour, wanted %d", ba.id, id)
		} else {
			if b.LeftNeighbour.ID != id {
				t.Errorf("blurb %d: got left neighbour %d, wanted %d", ba.id, b.LeftNeighbour.ID, id)
			}
		}
	})
	return ba
}

func (ba *blurbAsserter) hasNoLeftNeighbour() *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if b.LeftNeighbour != nil {
			t.Errorf("blurb %d: got left neighbour %d, wanted none", ba.id, b.LeftNeighbour.ID)
		}
	})
	return ba
}

func (ba *blurbAsserter) hasParent(id int) *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if b.Parent == nil {
			t.Errorf("blurb %d: got no parent, wanted %d", ba.id, id)
		} else {
			if b.Parent.ID != id {
				t.Errorf("blurb %d: got parent %d, wanted %d", ba.id, b.Parent.ID, id)
			}
		}
	})
	return ba
}

func (ba *blurbAsserter) hasNoParent() *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if b.Parent != nil {
			t.Errorf("blurb %d: got parent %d, wanted none", ba.id, b.Parent.ID)
		}
	})
	return ba
}

func (ba *blurbAsserter) hasNoShift() *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if !b.NoShift {
			t.Errorf("blurb %d: got allowed to shift, wanted no shift", ba.id)
		}
	})
	return ba
}

func (ba *blurbAsserter) hasLeftStop(id int) *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if b.LeftStop == nil {
			t.Errorf("blurb %d: got no left stop, wanted %d", ba.id, id)
		} else {
			if b.LeftStop.ID != id {
				t.Errorf("blurb %d: got left stop %d, wanted %d", ba.id, b.LeftStop.ID, id)
			}
		}
	})
	return ba
}

func (ba *blurbAsserter) hasRightStop(id int) *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if b.RightStop == nil {
			t.Errorf("blurb %d: got no right stop, wanted %d", ba.id, id)
		} else {
			if b.RightStop.ID != id {
				t.Errorf("blurb %d: got right stop %d, wanted %d", ba.id, b.RightStop.ID, id)
			}
		}
	})
	return ba
}

func (ba *blurbAsserter) hasKeepWith(id int) *blurbAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *Blurb, l *DescendantLayout) {
		if len(b.KeepWith) == 0 {
			t.Errorf("blurb %d: missing keep with %d", ba.id, id)
		} else {
			for _, v := range b.KeepWith {
				if v.ID == id {
					return
				}
			}
			t.Errorf("blurb %d: missing keep with %d", ba.id, id)
		}
	})
	return ba
}
