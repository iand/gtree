package gtree

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func lines(elems ...string) string { return strings.Join(elems, "\n") }

var testCases = []struct {
	name string
	in   string
	want *DescendantChart
}{
	{
		in: "1. A. Brown",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
				},
			},
		},
	},
	{
		in: "1. A. Brown b. 24 May 1819, London, England.; d. 22 Jan 1901, Isle of Wight, England.",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"b. 24 May 1819, London, England.",
					"d. 22 Jan 1901, Isle of Wight, England.",
				},
			},
		},
	},
	{
		in: "1. A. Brown (b. 24 May 1819, London, England.)",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"b. 24 May 1819, London, England.",
				},
			},
		},
	},
	{
		in: "1. A. Brown (1819-1901)",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
				},
			},
		},
	},
	{
		in: "1. A. Brown (1819-1901; carpenter)",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
					"carpenter",
				},
			},
		},
	},
	{
		in: "1. A. Brown (1819-1901 (carpenter))",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901 (carpenter)",
				},
			},
		},
	},
	{
		in: "1. A. Brown (1819-1901 carpenter",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"(1819-1901 carpenter",
				},
			},
		},
	},
	{
		name: "missing closing parantheses with inner parantheses",
		in:   "1. A. Brown (1819-1901 (carpenter)",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901 (carpenter",
				},
			},
		},
	},
	{
		name: "no whitespace before detail",
		in:   "1. A. Brown(b. 24 May 1819)",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"b. 24 May 1819",
				},
			},
		},
	},
	{
		name: "no name",
		in:   "1. (b. 24 May 1819)",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"b. 24 May 1819",
				},
			},
		},
	},
	{
		name: "no name no whitespace before detail",
		in:   "1.b. 24 May 1819",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"b. 24 May 1819",
				},
			},
		},
	},

	{
		name: "name with ancestry style details",
		in:   "1.Henry Johnson  b: Abt. 1806 in Kilford, Ireland. d: 17 Sep 1861 in Swindon, Wiltshire, England; age: 55.",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"Henry Johnson",
					"b: Abt. 1806 in Kilford, Ireland. d: 17 Sep 1861 in Swindon, Wiltshire, England",
					"age: 55.",
				},
			},
		},
	},

	{
		name: "name with gramps style details",
		in:   "1. Bennett, Edward (b. 1843-11-01 - St. David's, Carmarthenshire, Wales, d. before 1871), m. 1867-12-07 - St. Andrew's Catholic Church, High Street, Swansea, Glamorgan, Wales",
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"Bennett, Edward",
					"(b. 1843-11-01 - St. David's, Carmarthenshire, Wales, d. before 1871), m. 1867-12-07 - St. Andrew's Catholic Church, High Street, Swansea, Glamorgan, Wales",
				},
			},
		},
	},

	{
		in: lines(
			"1. A. Brown (1819-1901)",
			"   sp. B. Green (1819-1861)",
		),
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
				},
				Families: []*DescendantFamily{
					{
						Other: &DescendantPerson{
							ID: 2,
							Details: []string{
								"B. Green",
								"1819-1861",
							},
						},
					},
				},
			},
		},
	},
	{
		in: lines(
			"1. A. Brown (1819-1901)",
			"   sp. B. Green (1819-1861)",
		),
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
				},
				Families: []*DescendantFamily{
					{
						Other: &DescendantPerson{
							ID: 2,
							Details: []string{
								"B. Green",
								"1819-1861",
							},
						},
					},
				},
			},
		},
	},
	{
		name: "one spouse two children",
		in: lines(
			"1. A. Brown (1819-1901)",
			"  sp. B. Green (1819-1861)",
			"   2. C. Brown (1840-1901)",
			"   2. D. Brown (1841-1910)",
		),
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
				},
				Families: []*DescendantFamily{
					{
						Other: &DescendantPerson{
							ID: 2,
							Details: []string{
								"B. Green",
								"1819-1861",
							},
						},
						Children: []*DescendantPerson{
							{
								ID: 3,
								Details: []string{
									"C. Brown",
									"1840-1901",
								},
							},
							{
								ID: 4,
								Details: []string{
									"D. Brown",
									"1841-1910",
								},
							},
						},
					},
				},
			},
		},
	},
	{
		name: "two spouses one child each",
		in: lines(
			"1. A. Brown (1819-1901)",
			"sp. B. Green (1819-1861)",
			"   2. C. Brown (1840-1901)",
			"sp. E. Violet (1825-1920)",
			"   2. D. Brown (1850-1940)",
		),
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
				},
				Families: []*DescendantFamily{
					{
						Other: &DescendantPerson{
							ID: 2,
							Details: []string{
								"B. Green",
								"1819-1861",
							},
						},
						Children: []*DescendantPerson{
							{
								ID: 3,
								Details: []string{
									"C. Brown",
									"1840-1901",
								},
							},
						},
					},
					{
						Other: &DescendantPerson{
							ID: 4,
							Details: []string{
								"E. Violet",
								"1825-1920",
							},
						},
						Children: []*DescendantPerson{
							{
								ID: 5,
								Details: []string{
									"D. Brown",
									"1850-1940",
								},
							},
						},
					},
				},
			},
		},
	},
	{
		name: "two spouses one child",
		in: lines(
			"1. A. Brown (1819-1901)",
			"   sp. B. Green (1819-1861)",
			"   sp. E. Violet (1825-1920)",
			"   2. D. Brown (1850-1940)",
		),
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
				},
				Families: []*DescendantFamily{
					{
						Other: &DescendantPerson{
							ID: 2,
							Details: []string{
								"B. Green",
								"1819-1861",
							},
						},
					},
					{
						Other: &DescendantPerson{
							ID: 3,
							Details: []string{
								"E. Violet",
								"1825-1920",
							},
						},
						Children: []*DescendantPerson{
							{
								ID: 4,
								Details: []string{
									"D. Brown",
									"1850-1940",
								},
							},
						},
					},
				},
			},
		},
	},
	{
		name: "two spouses two children not in family",
		in: lines(
			"1. A. Brown (1819-1901)",
			"   2. C. Brown (1840-1901)",
			"   2. D. Brown (1850-1940)",
			"sp. B. Green (1819-1861)",
			"sp. E. Violet (1825-1920)",
		),
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"A. Brown",
					"1819-1901",
				},
				Families: []*DescendantFamily{
					{
						Children: []*DescendantPerson{
							{
								ID: 2,
								Details: []string{
									"C. Brown",
									"1840-1901",
								},
							},
							{
								ID: 3,
								Details: []string{
									"D. Brown",
									"1850-1940",
								},
							},
						},
					},
					{
						Other: &DescendantPerson{
							ID: 4,
							Details: []string{
								"B. Green",
								"1819-1861",
							},
						},
					},
					{
						Other: &DescendantPerson{
							ID: 5,
							Details: []string{
								"E. Violet",
								"1825-1920",
							},
						},
					},
				},
			},
		},
	},
	{
		name: "empty leading lines",
		in: lines(
			"",
			"1. John Doe (b. 1950)",
			"  2. Jane Doe (b. 1975)",
			"  sp. Richard Roe (b. 1974)",
			"    3. Sam Roe (b. 2000)",
			"  2. Jim Doe (b. 1978)",
		),
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: 1,
				Details: []string{
					"John Doe",
					"b. 1950",
				},
				Families: []*DescendantFamily{
					{
						Children: []*DescendantPerson{
							{
								ID: 2,
								Details: []string{
									"Jane Doe",
									"b. 1975",
								},
								Families: []*DescendantFamily{
									{
										Other: &DescendantPerson{
											ID: 3,
											Details: []string{
												"Richard Roe",
												"b. 1974",
											},
										},
										Children: []*DescendantPerson{
											{
												ID: 4,
												Details: []string{
													"Sam Roe",
													"b. 2000",
												},
											},
										},
									},
								},
							},
							{
								ID: 5,
								Details: []string{
									"Jim Doe",
									"b. 1978",
								},
							},
						},
					},
				},
			},
		},
	},

	{
		name: "ancestry descendant chart",
		in: `
1.Henry Johnson  b: Abt. 1806 in Kilford, Ireland. d: 17 Sep 1861 in Swindon, Wiltshire, England; age: 55.
  + Alice O’Connor  b: Abt. 1800 in Limerick, Ireland. d: 12 Oct 1896 in Trowbridge, Wiltshire, England; age: 96.
  2.Elizabeth Johnson  b: 7 Dec 1838 in Chippenham, Wiltshire, England. d: Bef. 1928 in Swindon, Wiltshire, England; age: 89.
  + George Martin  b: abt 1835 in Ireland. m: 28 Jun 1857 in Swindon, Wiltshire, England. d: Mar 1883 in Swindon, Wiltshire, England; age: 48.
    3.Elizabeth Ann Martin  b: 24 Apr 1858. d: 1859; age: 0.
    3.Martha Martin  b: abt 1860 in Trowbridge, Wiltshire, England. d: Deceased.
  2.Susan Johnson  b: 25 Apr 1840 in Swindon, Wiltshire, England. d: Deceased.
  2.Anna Johnson  b: 22 May 1842 in Chippenham, Wiltshire, England. d: 1 Oct 1898 in Bath, Somerset, England; age: 56.
  + William Brown  b: Abt. 1839 in Limerick, Ireland. m: 13 Nov 1864 in Swindon, Wiltshire, England. d: 11 May 1867 in St. Luke’s Infirmary, Bath, Somerset, England; age: 28.
    3.Thomas Brown  b: 3 Nov 1865 in Trowbridge, Wiltshire, England. d: Deceased.
    + Charles Lewis  b: 1 Nov 1843 in Bristol, Gloucestershire, England. m: 7 Dec 1867 in Swindon, Wiltshire, England. d: Bef. 1871 in Trowbridge, Wiltshire, England; age: 27.
    3.Emily Lewis  b: 15 Oct 1868 in Swindon, Wiltshire, England. d: 8 Aug 1956 in Wiltshire, England; age: 87.
    + Alfred Green  b: 25 Feb 1864 in Norton, Somerset, England. m: 4 Sep 1888 in Swindon, Wiltshire, England. d: 28 Feb 1955 in Chippenham, Wiltshire, England; age: 91.
    + Joseph Navarro  b: 1840 in Bristol, Gloucestershire, England. m: 28 Oct 1872 in Swindon, Wiltshire, England. d: 15 July 1880 in Trowbridge, Wiltshire, England; age: 40.
  2.David Johnson  b: 15 Feb 1844 in Chippenham, Wiltshire, England. d: Oct 1916 in Swindon, Wiltshire, England; age: 72.
  + Martha Jane Harper  b: abt 1846 in Fleur-de-Lys, Monmouthshire, Wales. m: 17 Sep 1873 in St. Luke's Church, Swindon, Wiltshire, England. d: Jul 1923 in Swindon, Wiltshire, England; age: 77.
    3.John H Johnson  b: abt 1875 in Swindon, Wiltshire, England. d: Deceased.
    3.Jane Elizabeth Harper Johnson  b: 1880 in Swindon, Wiltshire, England. d: Deceased.
  2.James Johnson  b: 30 Mar 1849 in Chippenham, Wiltshire, England. d: 6 Apr 1849 in Chippenham, Wiltshire, England; age: 0.
  2.Peter Johnson  b: 2 Nov 1851 in Trowbridge, Wiltshire, England. d: Jun 1936 in Swindon, Wiltshire, England; age: 84.
  + Helen Clark  b: abt 1854 in Trowbridge, Wiltshire, England. m: 2 Dec 1872 in Christchurch, Wiltshire, England. d: 25 Jul 1935 in Swindon, Wiltshire, England; age: 81.
    3.Samuel Johnson  b: abt 1874 in Trowbridge, Wiltshire, England. d: Dec 1948 in Chippenham, Wiltshire, England; age: 74.
    + Mary Wells  b: abt 1875 in Nk, Wiltshire, England. m: Jul 1902 in Wiltshire, England. d: Deceased.
    3.Eliza Johnson  b: abt 1882 in Devizes, Wiltshire, England. d: Abt 1961 in Salisbury, Wiltshire, England; age: 79.
 `,
		want: &DescendantChart{
			Root: &DescendantPerson{
				ID: int(1),
				Details: []string{
					string("Henry Johnson"),
					string("b: Abt. 1806 in Kilford, Ireland. d: 17 Sep 1861 in Swindon, Wiltshire, England"),
					string("age: 55."),
				},
				Families: []*DescendantFamily{
					{
						Other: &DescendantPerson{
							ID: int(2),
							Details: []string{
								string("Alice O’Connor"),
								string("b: Abt. 1800 in Limerick, Ireland. d: 12 Oct 1896 in Trowbridge, Wiltshire, England"),
								string("age: 96."),
							},
							Families: []*DescendantFamily(nil),
						},
						Details: []string(nil),
						Children: []*DescendantPerson{
							{
								ID: int(3),
								Details: []string{
									string("Elizabeth Johnson"),
									string("b: 7 Dec 1838 in Chippenham, Wiltshire, England. d: Bef. 1928 in Swindon, Wiltshire, England"),
									string("age: 89."),
								},
								Families: []*DescendantFamily{
									{
										Other: &DescendantPerson{
											ID: int(4),
											Details: []string{
												string("George Martin"),
												string("b: abt 1835 in Ireland. m: 28 Jun 1857 in Swindon, Wiltshire, England. d: Mar 1883 in Swindon, Wiltshire, England"),
												string("age: 48."),
											},
											Families: []*DescendantFamily(nil),
										},
										Details: []string(nil),
										Children: []*DescendantPerson{
											{
												ID: int(5),
												Details: []string{
													string("Elizabeth Ann Martin"),
													string("b: 24 Apr 1858. d: 1859"),
													string("age: 0."),
												},
												Families: []*DescendantFamily(nil),
											},
											{
												ID: int(6),
												Details: []string{
													string("Martha Martin"),
													string("b: abt 1860 in Trowbridge, Wiltshire, England. d: Deceased."),
												},
												Families: []*DescendantFamily(nil),
											},
										},
									},
								},
							},
							{
								ID: int(7),
								Details: []string{
									string("Susan Johnson"),
									string("b: 25 Apr 1840 in Swindon, Wiltshire, England. d: Deceased."),
								},
								Families: []*DescendantFamily(nil),
							},
							{
								ID: int(8),
								Details: []string{
									string("Anna Johnson"),
									string("b: 22 May 1842 in Chippenham, Wiltshire, England. d: 1 Oct 1898 in Bath, Somerset, England"),
									string("age: 56."),
								},
								Families: []*DescendantFamily{
									{
										Other: &DescendantPerson{
											ID: int(9),
											Details: []string{
												string("William Brown"),
												string("b: Abt. 1839 in Limerick, Ireland. m: 13 Nov 1864 in Swindon, Wiltshire, England. d: 11 May 1867 in St. Luke’s Infirmary, Bath, Somerset, England"),
												string("age: 28."),
											},
											Families: []*DescendantFamily(nil),
										},
										Details: []string(nil),
										Children: []*DescendantPerson{
											{
												ID: int(10),
												Details: []string{
													string("Thomas Brown"),
													string("b: 3 Nov 1865 in Trowbridge, Wiltshire, England. d: Deceased."),
												},
												Families: []*DescendantFamily{
													{
														Other: &DescendantPerson{
															ID: int(11),
															Details: []string{
																string("Charles Lewis"),
																string("b: 1 Nov 1843 in Bristol, Gloucestershire, England. m: 7 Dec 1867 in Swindon, Wiltshire, England. d: Bef. 1871 in Trowbridge, Wiltshire, England"),
																string("age: 27."),
															},
															Families: []*DescendantFamily(nil),
														},
														Details:  []string(nil),
														Children: []*DescendantPerson(nil),
													},
												},
											},
											{
												ID: int(12),
												Details: []string{
													string("Emily Lewis"),
													string("b: 15 Oct 1868 in Swindon, Wiltshire, England. d: 8 Aug 1956 in Wiltshire, England"),
													string("age: 87."),
												},
												Families: []*DescendantFamily{
													{
														Other: &DescendantPerson{
															ID: int(13),
															Details: []string{
																string("Alfred Green"),
																string("b: 25 Feb 1864 in Norton, Somerset, England. m: 4 Sep 1888 in Swindon, Wiltshire, England. d: 28 Feb 1955 in Chippenham, Wiltshire, England"),
																string("age: 91."),
															},
															Families: []*DescendantFamily(nil),
														},
														Details:  []string(nil),
														Children: []*DescendantPerson(nil),
													},
													{
														Other: &DescendantPerson{
															ID: int(14),
															Details: []string{
																string("Joseph Navarro"),
																string("b: 1840 in Bristol, Gloucestershire, England. m: 28 Oct 1872 in Swindon, Wiltshire, England. d: 15 July 1880 in Trowbridge, Wiltshire, England"),
																string("age: 40."),
															},
															Families: []*DescendantFamily(nil),
														},
														Details:  []string(nil),
														Children: []*DescendantPerson(nil),
													},
												},
											},
										},
									},
								},
							},
							{
								ID: int(15),
								Details: []string{
									string("David Johnson"),
									string("b: 15 Feb 1844 in Chippenham, Wiltshire, England. d: Oct 1916 in Swindon, Wiltshire, England"),
									string("age: 72."),
								},
								Families: []*DescendantFamily{
									{
										Other: &DescendantPerson{
											ID: int(16),
											Details: []string{
												string("Martha Jane Harper"),
												string("b: abt 1846 in Fleur-de-Lys, Monmouthshire, Wales. m: 17 Sep 1873 in St. Luke's Church, Swindon, Wiltshire, England. d: Jul 1923 in Swindon, Wiltshire, England"),
												string("age: 77."),
											},
											Families: []*DescendantFamily(nil),
										},
										Details: []string(nil),
										Children: []*DescendantPerson{
											{
												ID: int(17),
												Details: []string{
													string("John H Johnson"),
													string("b: abt 1875 in Swindon, Wiltshire, England. d: Deceased."),
												},
												Families: []*DescendantFamily(nil),
											},
											{
												ID: int(18),
												Details: []string{
													string("Jane Elizabeth Harper Johnson"),
													string("b: 1880 in Swindon, Wiltshire, England. d: Deceased."),
												},
												Families: []*DescendantFamily(nil),
											},
										},
									},
								},
							},
							{
								ID: int(19),
								Details: []string{
									string("James Johnson"),
									string("b: 30 Mar 1849 in Chippenham, Wiltshire, England. d: 6 Apr 1849 in Chippenham, Wiltshire, England"),
									string("age: 0."),
								},
								Families: []*DescendantFamily(nil),
							},
							{
								ID: int(20),
								Details: []string{
									string("Peter Johnson"),
									string("b: 2 Nov 1851 in Trowbridge, Wiltshire, England. d: Jun 1936 in Swindon, Wiltshire, England"),
									string("age: 84."),
								},
								Families: []*DescendantFamily{
									{
										Other: &DescendantPerson{
											ID: int(21),
											Details: []string{
												string("Helen Clark"),
												string("b: abt 1854 in Trowbridge, Wiltshire, England. m: 2 Dec 1872 in Christchurch, Wiltshire, England. d: 25 Jul 1935 in Swindon, Wiltshire, England"),
												string("age: 81."),
											},
											Families: []*DescendantFamily(nil),
										},
										Details: []string(nil),
										Children: []*DescendantPerson{
											{
												ID: int(22),
												Details: []string{
													string("Samuel Johnson"),
													string("b: abt 1874 in Trowbridge, Wiltshire, England. d: Dec 1948 in Chippenham, Wiltshire, England"),
													string("age: 74."),
												},
												Families: []*DescendantFamily{
													{
														Other: &DescendantPerson{
															ID: int(23),
															Details: []string{
																string("Mary Wells"),
																string("b: abt 1875 in Nk, Wiltshire, England. m: Jul 1902 in Wiltshire, England. d: Deceased."),
															},
															Families: []*DescendantFamily(nil),
														},
														Details:  []string(nil),
														Children: []*DescendantPerson(nil),
													},
												},
											},
											{
												ID: int(24),
												Details: []string{
													string("Eliza Johnson"),
													string("b: abt 1882 in Devizes, Wiltshire, England. d: Abt 1961 in Salisbury, Wiltshire, England"),
													string("age: 79."),
												},
												Families: []*DescendantFamily(nil),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestParse(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			p := new(Parser)
			got, err := p.Parse(ctx, strings.NewReader(tc.in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
