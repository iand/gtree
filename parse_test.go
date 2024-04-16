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
		in: "1. A. Brown; b. 24 May 1819, London, England.; d. 22 Jan 1901, Isle of Wight, England.",
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
		in: "1. A. Brown (1819-1901) carpenter",
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
					"1819-1901 carpenter",
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
					"1819-1901",
					"carpenter",
				},
			},
		},
	},
	{
		in: "1. A. Brown(b. 24 May 1819)",
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
		in: "1. (b. 24 May 1819)",
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
		in: "1.;b. 24 May 1819",
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
		in: lines(
			"1. A. Brown (1819-1901)",
			"   sp. B. Green (1819-1861)",
			"   1. C. Brown (1840-1901)",
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
		name: "two spouses",
		in: lines(
			"1. A. Brown (1819-1901)",
			"   sp. B. Green (1819-1861)",
			"   1. C. Brown (1840-1901)",
			"   sp. E. Violet (1825-1920)",
			"   1. D. Brown (1850-1940)",
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
		in: lines(
			"1. A. Brown (1819-1901)",
			"   sp. B. Green (1819-1861)",
			"   sp. E. Violet (1825-1920)",
			"   1. D. Brown (1850-1940)",
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
