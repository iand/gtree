//go:build ignore

// run this using go run fan_example.go

package main

import (
	"flag"
	"fmt"

	"github.com/iand/gtree"
)

var debugFlag = flag.Bool("debug", false, "emit debugging information")

func main() {
	flag.Parse()
	ch := &gtree.FanChart{}

	_ = ch

	var idCounter int
	nextID := func() int {
		idCounter++
		return idCounter
	}

	makePerson := func(name string, birth, death int, occ string) *gtree.FanPerson {
		return &gtree.FanPerson{
			ID:       nextID(),
			Headings: []string{name},
			Details: []string{
				fmt.Sprintf("b. %d", birth),
				fmt.Sprintf("d. %d", death),
				occ,
			},
		}
	}

	// Generation 2 – 4 couples = 8 people
	g3 := []*gtree.FanPerson{
		makePerson("John Smith", 1890, 1955, "Farmer"),
		makePerson("Mary Smith", 1895, 1972, "Homemaker"),
		makePerson("James Brown", 1892, 1959, "Coal Miner"),
		makePerson("Helen Brown", 1896, 1981, "Seamstress"),
		makePerson("George White", 1888, 1951, "Baker"),
		makePerson("Elizabeth White", 1893, 1970, "Nurse"),
		makePerson("Thomas Green", 1891, 1964, "Carpenter"),
		makePerson("Sarah Green", 1894, 1978, "Teacher"),
	}

	g2 := []*gtree.FanPerson{
		makePerson("Robert Smith", 1920, 1988, "Engineer"),
		makePerson("Alice Brown", 1923, 1992, "Librarian"),
		makePerson("Edward White", 1921, 1990, "Doctor"),
		makePerson("Florence Green", 1924, 2001, "Artist"),
	}

	for i := range g2 {
		g2[i].Father = g3[2*i]
		g2[i].Mother = g3[2*i+1]
	}

	g1 := []*gtree.FanPerson{
		makePerson("James Smith", 1945, 2000, "Mechanic"),
		makePerson("Anne White", 1946, 2010, "Teacher"),
	}
	for i := range g1 {
		g1[i].Father = g2[2*i]
		g1[i].Mother = g2[2*i+1]
	}

	root := makePerson("Charles Smith", 1975, 0, "Software Engineer")
	root.Father = g1[0]
	root.Mother = g1[1]

	chart := &gtree.FanChart{
		Title: "Example Fan Chart",
		Notes: []string{"produced with go run fan_example.go"},
		Root:  root,
	}

	opts := gtree.DefaultFanLayoutOptions()
	opts.Debug = *debugFlag
	opts.MaxGenerations = 3

	lay, err := gtree.GenerateFanLayout(chart, opts)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	s, err := gtree.SVG(lay)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(s)
}
