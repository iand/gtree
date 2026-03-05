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
				node(1).
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
				node(1).
					hasText("Person One").
					hasNoParent().
					hasNoLeftNeighbour().
					hasKeepTightRight(-2).
					inRow(0),
				node(-2).
					hasText("=").
					hasNoParent().
					hasLeftNeighbour(1).
					inRow(0),
				node(2).
					hasText("Person Two").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-2).
					inRow(0),
			},
		},
		{
			name: "one person with three spouses",
			in:   onePersonWithThreeSpouses,
			assertions: []layoutAssertion{
				node(1).
					hasText("Person One").
					hasNoParent().
					hasNoLeftNeighbour().
					hasKeepTightRight(-2).
					inRow(0),
				node(-2).
					hasText("= (1)").
					hasNoParent().
					hasLeftNeighbour(1).
					inRow(0),
				node(2).
					hasText("Person Two").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-2).
					inRow(0),
				node(-3).
					hasText("= (2)").
					hasNoParent().
					hasLeftNeighbour(2).
					inRow(0),
				node(3).
					hasText("Person Three").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-3).
					inRow(0),
				node(-4).
					hasText("= (3)").
					hasNoParent().
					hasLeftNeighbour(3).
					inRow(0),
				node(4).
					hasText("Person Four").
					hasNoShift().
					hasLeftNeighbour(-4).
					inRow(0),
			},
		},
		{
			name: "one person with spouse and children",
			in:   onePersonWithSpouseAndChildren,
			assertions: []layoutAssertion{
				node(1).
					hasText("Person One").
					hasNoParent().
					hasNoLeftNeighbour().
					hasKeepTightRight(-2).
					inRow(0),
				node(-2).
					hasText("=").
					hasNoParent().
					hasLeftNeighbour(1).
					inRow(0),
				node(2).
					hasText("Person Two").
					hasNoParent().
					hasNoShift().
					hasLeftNeighbour(-2).
					inRow(0),
				node(3).
					hasText("Person Three").
					hasParent(-2).
					hasNoLeftNeighbour().
					inRow(1),
				node(4).
					hasText("Person Four").
					inRow(1).
					hasParent(-2).
					hasLeftNeighbour(3),
			},
		},

		{
			name: "three descendants with two marriages",
			in: &DescendantChart{
				Title: string(""),
				Notes: []string(nil),
				Root: &DescendantPerson{
					ID: int(0),
					Headings: []string{
						string("Person 0"),
					},
					Details: []string{
						string("bap. 6 Oct 1714"),
						string("bur. 11 Oct 1807"),
					},
					Families: []*DescendantFamily{
						{
							Other: &DescendantPerson{
								ID: int(1),
								Headings: []string{
									string("Person 0 spouse"),
								},
								Details: []string{
									string("bap. 4 Dec 1716"),
									string("bur. 30 Jan 1810"),
								},
								Families: []*DescendantFamily(nil),
								Tags:     []string(nil),
							},
							Details: []string{
								string("m. 3 Feb 1736/7"),
								string("bur. 11 Oct 1807"),
							},
							Children: []*DescendantPerson{
								{
									ID: int(2),
									Headings: []string{
										string("Person 2"),
									},
									Details: []string{
										string("bap. 25 Dec 1738"),
										string("bur. 21 Mar 1815"),
									},
									Families: []*DescendantFamily{
										{
											Other: &DescendantPerson{
												ID: int(3),
												Headings: []string{
													string("Person 2 first spouse"),
												},
												Details: []string{
													string("bur. 21 Feb 1782"),
												},
												Families: []*DescendantFamily(nil),
												Tags:     []string(nil),
											},
											Details: []string{
												string("m. 28 Jul 1766"),
												string("bur. 21 Feb 1782"),
											},
											Children: []*DescendantPerson{
												{
													ID: int(4),
													Headings: []string{
														string("Person 3"),
													},
													Details: []string{
														string("bap. 1 May 1767"),
														string("d. 29 Nov 1854"),
													},
													Families: []*DescendantFamily{
														{
															Other: &DescendantPerson{
																ID: int(5),
																Headings: []string{
																	string("Person 3 spouse"),
																},
																Details: []string{
																	string("bap. 7 Nov 1762"),
																	string("bur. 13 Oct 1830"),
																},
																Families: []*DescendantFamily(nil),
																Tags:     []string(nil),
															},
															Details: []string{
																string("m. 13 Oct 1788"),
																string("bur. 13 Oct 1830"),
															},
															Children: []*DescendantPerson(nil),
														},
													},
													Tags: []string(nil),
												},
											},
										},
										{
											Other: &DescendantPerson{
												ID: int(6),
												Headings: []string{
													string("Person 2 second spouse"),
												},
												Details: []string{
													string("bap. 19 Mar 1751/2"),
													string("bur. 8 Jul 1820"),
												},
												Families: []*DescendantFamily(nil),
												Tags:     []string(nil),
											},
											Details: []string{
												string("m. 25 Jul 1785"),
												string("bur. 21 Mar 1815"),
											},
											Children: []*DescendantPerson(nil),
										},
									},
									Tags: []string(nil),
								},
							},
						},
					},
					Tags: []string(nil),
				},
			},
			assertions: []layoutAssertion{
				node(0).
					hasText("Person 0", "bap. 6 Oct 1714", "bur. 11 Oct 1807").
					hasNoParent().
					hasNoLeftNeighbour().
					hasKeepTightRight(-1).
					inRow(0),
				node(-1).
					hasText("=", "m. 3 Feb 1736/7", "bur. 11 Oct 1807").
					hasNoParent().
					hasLeftNeighbour(0).
					hasKeepTightRight(1).
					inRow(0),
				node(1).
					hasText("Person 0 spouse", "bap. 4 Dec 1716", "bur. 30 Jan 1810").
					hasNoParent().
					hasLeftNeighbour(-1).
					inRow(0),

				node(2).
					hasText("Person 2", "bap. 25 Dec 1738", "bur. 21 Mar 1815").
					hasParent(-1).
					hasNoLeftNeighbour().
					hasKeepTightRight(-3).
					inRow(1),
				node(-3).
					hasText("= (1)", "m. 28 Jul 1766", "bur. 21 Feb 1782").
					hasNoParent().
					hasLeftNeighbour(2).
					hasKeepTightRight(3).
					inRow(1),
				node(3).
					hasText("Person 2 first spouse", "bur. 21 Feb 1782").
					hasNoParent().
					hasLeftNeighbour(-3).
					inRow(1),
				node(-6).
					hasText("= (2)", "m. 25 Jul 1785", "bur. 21 Mar 1815").
					hasNoParent().
					hasLeftNeighbour(3).
					hasKeepTightRight(6).
					inRow(1),
				node(6).
					hasText("Person 2 second spouse", "bap. 19 Mar 1751/2", "bur. 8 Jul 1820").
					hasNoParent().
					hasLeftNeighbour(-6).
					inRow(1),

				node(4).
					hasText("Person 3", "bap. 1 May 1767", "d. 29 Nov 1854").
					hasParent(-3).
					hasNoLeftNeighbour().
					hasKeepTightRight(-5).
					inRow(2),
				node(-5).
					hasText("=", "m. 13 Oct 1788", "bur. 13 Oct 1830").
					hasNoParent().
					hasLeftNeighbour(4).
					hasKeepTightRight(5).
					inRow(2),
				node(5).
					hasText("Person 3 spouse", "bap. 7 Nov 1762", "bur. 13 Oct 1830").
					hasNoParent().
					hasLeftNeighbour(-5).
					inRow(2),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			l, err := tc.in.Layout(nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, a := range tc.assertions {
				a.assert(t, l)
			}

			// Check that no two nodes in the same row overlap.
			for rowIdx, row := range l.rows {
				for i := 1; i < len(row); i++ {
					if row[i].Left() < row[i-1].Right() {
						t.Errorf("row %d: node %d (left=%v) overlaps node %d (right=%v)",
							rowIdx, row[i].ID, row[i].Left(), row[i-1].ID, row[i-1].Right())
					}
				}
			}
		})
	}
}

type layoutAssertion interface {
	assert(*testing.T, *DescendantLayout)
}

func node(id int) *nodeAsserter {
	return &nodeAsserter{id: id}
}

type nodeAsserter struct {
	id  int
	fns []func(t *testing.T, b *AlignedNode, l *DescendantLayout)
}

func (a *nodeAsserter) assert(t *testing.T, l *DescendantLayout) {
	b, ok := l.nodes[a.id]
	if !ok {
		t.Errorf("node %d is missing", a.id)
		return
	}

	for _, fn := range a.fns {
		fn(t, b, l)
	}
}

func (ba *nodeAsserter) hasText(texts ...string) *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if len(b.DetailTexts.Lines) != len(texts)-1 {
			t.Fatalf("node %d: got %d detail texts, wanted %d", ba.id, len(b.DetailTexts.Lines), len(texts)-1)
		}

		for i := range texts {
			if i < len(b.HeadingTexts.Lines) {
				if b.HeadingTexts.Lines[i] != texts[i] {
					t.Errorf("node %d: got heading text %q, wanted %q", ba.id, b.HeadingTexts.Lines[i], texts[i])
				}
				continue
			}
			if b.DetailTexts.Lines[i-len(b.HeadingTexts.Lines)] != texts[i] {
				t.Errorf("node %d: got detail text %q, wanted %q", ba.id, b.DetailTexts.Lines[i-1], texts[i])
			}
		}
	})
	return ba
}

func (ba *nodeAsserter) inRow(row int) *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if b.Row != row {
			t.Errorf("node %d: got row %d, wanted %d", b.ID, b.Row, row)
		}

		if len(l.rows) <= row {
			t.Errorf("node %d: layout didn't have row %d", ba.id, row)
			return
		}

		if len(l.rows[row]) == 0 {
			t.Errorf("node %d: layout didn't have any nodes in row %d", ba.id, row)
			return

		}

		for _, b2 := range l.rows[row] {
			if b2.ID == b.ID {
				return
			}
		}
		t.Errorf("node %d: missing from layout row %d", ba.id, row)
	})
	return ba
}

func (ba *nodeAsserter) hasLeftNeighbour(id int) *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if b.LeftNeighbour == nil {
			t.Errorf("node %d: got no left neighbour, wanted %d", ba.id, id)
		} else {
			if b.LeftNeighbour.ID != id {
				t.Errorf("node %d: got left neighbour %d, wanted %d", ba.id, b.LeftNeighbour.ID, id)
			}
		}
	})
	return ba
}

func (ba *nodeAsserter) hasNoLeftNeighbour() *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if b.LeftNeighbour != nil {
			t.Errorf("node %d: got left neighbour %d, wanted none", ba.id, b.LeftNeighbour.ID)
		}
	})
	return ba
}

func (ba *nodeAsserter) hasParent(id int) *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if b.Parent == nil {
			t.Errorf("node %d: got no parent, wanted %d", ba.id, id)
		} else {
			if b.Parent.ID != id {
				t.Errorf("node %d: got parent %d, wanted %d", ba.id, b.Parent.ID, id)
			}
		}
	})
	return ba
}

func (ba *nodeAsserter) hasNoParent() *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if b.Parent != nil {
			t.Errorf("node %d: got parent %d, wanted none", ba.id, b.Parent.ID)
		}
	})
	return ba
}

func (ba *nodeAsserter) hasNoShift() *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if !b.NoShift {
			t.Errorf("node %d: got allowed to shift, wanted no shift", ba.id)
		}
	})
	return ba
}

// func (ba *nodeAsserter) hasLeftStop(id int) *nodeAsserter {
// 	ba.fns = append(ba.fns, func(t *testing.T, b *Node, l *DescendantLayout) {
// 		if b.LeftStop == nil {
// 			t.Errorf("node %d: got no left stop, wanted %d", ba.id, id)
// 		} else {
// 			if b.LeftStop.ID != id {
// 				t.Errorf("node %d: got left stop %d, wanted %d", ba.id, b.LeftStop.ID, id)
// 			}
// 		}
// 	})
// 	return ba
// }

// func (ba *nodeAsserter) hasRightStop(id int) *nodeAsserter {
// 	ba.fns = append(ba.fns, func(t *testing.T, b *Node, l *DescendantLayout) {
// 		if b.RightStop == nil {
// 			t.Errorf("node %d: got no right stop, wanted %d", ba.id, id)
// 		} else {
// 			if b.RightStop.ID != id {
// 				t.Errorf("node %d: got right stop %d, wanted %d", ba.id, b.RightStop.ID, id)
// 			}
// 		}
// 	})
// 	return ba
// }

func (ba *nodeAsserter) hasKeepTightRight(id int) *nodeAsserter {
	ba.fns = append(ba.fns, func(t *testing.T, b *AlignedNode, l *DescendantLayout) {
		if b.KeepTightRight == nil {
			t.Errorf("node %d: missing keep tight right %d", ba.id, id)
		} else {
			if b.KeepTightRight.ID != id {
				t.Errorf("node %d: incorrect keep tight right %d, wanted %d", ba.id, b.KeepTightRight.ID, id)
			}
		}
	})
	return ba
}
