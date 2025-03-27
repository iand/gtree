//go:build ignore

// run this using go run butterfly_example.go

package main

import (
	"flag"
	"fmt"

	"github.com/iand/gtree"
)

var debugFlag = flag.Bool("debug", false, "emit debugging information")

func main() {
	flag.Parse()
	ch := &gtree.ButterflyChart{
		TitleLine1: "Example Butterfly Chart",
		Root: &gtree.ButterflyPerson{
			ID:          1,
			Forenames:   "Person",
			Surname:     "SMITH",
			DetailLine1: "b. 25 Oct 1850",
			DetailLine2: "d. 12 Dec 1914",
			Father: &gtree.ButterflyPerson{
				ID:          2,
				Forenames:   "Father",
				Surname:     "SMITH",
				DetailLine1: "b. 25 Oct 1822",
				DetailLine2: "d. 1 Mar 1868",
				Father: &gtree.ButterflyPerson{
					ID:          3,
					Forenames:   "Grandfather",
					Surname:     "SMITH",
					DetailLine1: "b. 6 Jan 1799",
					DetailLine2: "d. 27 Sep 1860",
				},
				Mother: &gtree.ButterflyPerson{
					ID:          4,
					Forenames:   "Grandmother",
					Surname:     "PURCELL",
					DetailLine1: "b. 12 Oct 1800",
					DetailLine2: "d. 19 Jun 1840",
					Father: &gtree.ButterflyPerson{
						ID:          5,
						Forenames:   "Great Grandfather",
						Surname:     "PURCELL",
						DetailLine1: "b. 25 May 1777",
						DetailLine2: "",
					},
				},
			},
			Mother: &gtree.ButterflyPerson{
				ID:          6,
				Forenames:   "Mother",
				Surname:     "BROWN",
				DetailLine1: "b. 25 Oct 1828",
				DetailLine2: "d. 9 Feb 1890",
				Father: &gtree.ButterflyPerson{
					ID:          7,
					Forenames:   "Grandfather",
					Surname:     "BROWN",
					DetailLine1: "b. 19 Feb 1800",
					DetailLine2: "d. 11 Oct 1858",
				},
				Mother: &gtree.ButterflyPerson{
					ID:          8,
					Forenames:   "Grandmother",
					Surname:     "PAINE",
					DetailLine1: "b. 14 Jan 1806",
					DetailLine2: "d. 4 Dec 1880",
				},
			},
		},
	}

	opts := gtree.DefaultButterflyLayoutOptions()
	opts.Debug = *debugFlag
	s, err := ch.RenderSVG(opts)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(s)
}
