# gtree

A Go package for generating family tree diagrams as SVG.

[![Test Status](https://github.com/iand/gtree/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/iand/gtree/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/iand/gtree)](https://goreportcard.com/report/github.com/iand/gtree)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/iand/gtree)

## Chart Types

- **Descendant Chart** -- root person at the top, successive generations in horizontal rows below.
- **Ancestor Chart** -- root person on the left, ancestors expanding to the right.
- **Fan Chart** -- root person at the centre, ancestors radiating outward in concentric arcs.
- **Butterfly Chart** -- root person in the middle, paternal ancestors to one side and maternal to the other.

All chart types output SVG.

## Getting Started

```sh
go get github.com/iand/gtree@latest
```

## Usage

Build a chart data structure, compute its layout, then render to SVG:

```go
ch := &gtree.DescendantChart{
    Title: "The Smith Family",
    Root: &gtree.DescendantPerson{
        ID:       1,
        Headings: []string{"John Smith"},
        Details:  []string{"b. 1850", "d. 1914"},
        Families: []*gtree.DescendantFamily{
            {
                Other: &gtree.DescendantPerson{
                    ID:       2,
                    Headings: []string{"Jane Doe"},
                },
                Children: []*gtree.DescendantPerson{
                    {ID: 3, Headings: []string{"Alice Smith"}},
                    {ID: 4, Headings: []string{"Bob Smith"}},
                },
            },
        },
    },
}

layout, err := ch.Layout(nil) // nil uses default options
if err != nil {
    log.Fatal(err)
}

svg, err := gtree.SVG(layout, nil) // nil omits fixed paper size
if err != nil {
    log.Fatal(err)
}
fmt.Println(svg)
```

Ancestor charts follow the same pattern with `AncestorChart` and `AncestorPerson`:

```go
ch := &gtree.AncestorChart{
    Title: "Ancestors of John Smith",
    Root: &gtree.AncestorPerson{
        ID:      1,
        Details: []string{"John Smith", "b. 1850"},
        Father:  &gtree.AncestorPerson{ID: 2, Details: []string{"Father Smith"}},
        Mother:  &gtree.AncestorPerson{ID: 3, Details: []string{"Mother Brown"}},
    },
}

layout, err := ch.Layout(nil)
```

### Parsing a Descendant List

The `Parser` reads an indented text format into a `DescendantChart`:

```
1. John Smith (b. 1850, d. 1914)
  sp. Jane Doe (m. 1875)
    2. Alice Smith (b. 1876)
    2. Bob Smith (b. 1878)
      sp. Carol Jones (m. 1900)
        3. Eve Smith (b. 1901)
```

```go
parser := new(gtree.Parser)
chart, err := parser.Parse(ctx, strings.NewReader(input))
```

Lines prefixed with a generation number (1, 2, 3...) are people.
Lines prefixed with `sp` or `+` are spouses. Detail text goes in
parentheses. See the `Parser` documentation for the full format.

### Layout Options

Each chart type has its own options struct with defaults:

```go
opts := gtree.DefaultLayoutOptions()       // descendant charts
opts := gtree.DefaultAncestorLayoutOptions() // ancestor charts
```

Options include font styles, spacing, margins, line widths, and
background colour. Pass `nil` to `Layout()` to use defaults.

## Examples

Runnable examples are in the repository root:

```sh
go run descendant_example.go > chart.svg
go run ancestor_example.go > chart.svg
go run butterfly_example.go > chart.svg
go run fan_example.go > chart.svg
go run parse_example.go > chart.svg
```

## Documentation

Full API documentation is at [pkg.go.dev/github.com/iand/gtree](https://pkg.go.dev/github.com/iand/gtree).

## License

This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying [`UNLICENSE`](UNLICENSE) file.
