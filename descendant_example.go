//go:build ignore

// run this using go run descendant_example.go

package main

import (
	"flag"
	"fmt"

	"github.com/iand/gtree"
)

var debugFlag = flag.Bool("debug", false, "emit debugging information")

func main() {
	flag.Parse()
	ch := &gtree.DescendantChart{
		Title: "Example Descendant Chart",
		Notes: []string{},
		Root: &gtree.DescendantPerson{
			ID:      1,
			Details: []string{"Person One", "b. 25 Oct 1850", "d. 12 Dec 1914"},
			Families: []*gtree.DescendantFamily{
				{
					Other: &gtree.DescendantPerson{
						ID:      2,
						Details: []string{"Spouse A"},
					},
					Children: []*gtree.DescendantPerson{
						{
							ID:      3,
							Details: []string{"Fam A Child One"},
						},
						{
							ID:      4,
							Details: []string{"Fam A Child Two"},
							Families: []*gtree.DescendantFamily{
								{
									Other: &gtree.DescendantPerson{
										ID:      5,
										Details: []string{"Spouse C"},
									},
									Children: []*gtree.DescendantPerson{
										{
											ID:      6,
											Details: []string{"Fam C Child One"},
										},
										{
											ID:      7,
											Details: []string{"Fam C Child Two"},
										},
									},
								},
							},
						},
					},
					Details: []string{"m. 14 Aug 1875"},
				},
				{
					Other: &gtree.DescendantPerson{
						ID:      8,
						Details: []string{"Spouse B"},
					},
					Children: []*gtree.DescendantPerson{
						{
							ID:      9,
							Details: []string{"Fam B Child One"},
						},
						{
							ID:      10,
							Details: []string{"Fam B Child Two"},
						},
						{
							ID:      11,
							Details: []string{"Fam B Child Three"},
						},
					},
					Details: []string{"m. 14 Aug 1892"},
				},
			},
		},
	}

	opts := gtree.DefaultLayoutOptions()
	opts.Debug = *debugFlag

	lay := ch.Layout(opts)
	s, err := gtree.SVG(lay)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(s)
}
