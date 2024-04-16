//go:build ignore

// run this using go run ancestor_example.go

package main

import (
	"flag"
	"fmt"

	"github.com/iand/gtree"
)

var debugFlag = flag.Bool("debug", false, "emit debugging information")

func main() {
	flag.Parse()
	ch := &gtree.AncestorChart{
		Title: "Example Ancestor Chart",
		Notes: []string{},
		Root: &gtree.AncestorPerson{
			ID:      1,
			Details: []string{"Person Smith", "b. 25 Oct 1850", "d. 12 Dec 1914"},
			Father: &gtree.AncestorPerson{
				ID:      2,
				Details: []string{"Father Smith", "b. 25 Oct 1822", "d. 1 Mar 1868"},
				Father: &gtree.AncestorPerson{
					ID:      3,
					Details: []string{"Grandfather Smith", "b. 6 Jan 1799", "d. 27 Sep 1860"},
				},
				Mother: &gtree.AncestorPerson{
					ID:      4,
					Details: []string{"Grandmother Purcell", "b. 12 Oct 1800", "d. 19 Jun 1840"},
					Father: &gtree.AncestorPerson{
						ID:      5,
						Details: []string{"Great Grandfather Purcell", "b. 25 May 1777"},
					},
				},
			},
			Mother: &gtree.AncestorPerson{
				ID:      6,
				Details: []string{"Mother Brown", "b. 25 Oct 1828", "d. 9 Feb 1890"},
				Father: &gtree.AncestorPerson{
					ID:      7,
					Details: []string{"Father Brown", "b. 19 Feb 1800", "d. 11 Oct 1858"},
				},
				Mother: &gtree.AncestorPerson{
					ID:      8,
					Details: []string{"Mother Brown", "b. 14 Jan 1806", "d. 4 Dec 1880"},
				},
			},
		},
	}

	opts := gtree.DefaultAncestorLayoutOptions()
	opts.Debug = *debugFlag
	lay := ch.Layout(opts)

	s, err := gtree.SVG(lay)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(s)
}
