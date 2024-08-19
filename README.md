# gtree

`gtree` is a Go package designed to generate family tree diagrams. 

[![Test Status](https://github.com/iand/gtree/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/iand/gtree/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/iand/gtree)](https://goreportcard.com/report/github.com/iand/gtree)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/iand/gtree)

## Purpose

The `gtree` package is designed to help you generate genealogical diagrams in the form of family trees. It supports both ancestor and descendant chart layouts, making it easy to visualize family connections across generations. The package is ideal for developers looking to integrate family tree generation into their Go applications or for anyone interested in generating genealogical diagrams programmatically.


## Capabilities

- **Generate Ancestor Charts**: Visualize an individual's ancestors, with the root person on the left and each successive generation aligned vertically to the right.
- **Generate Descendant Charts**: Illustrate an individual's descendants, with the root person at the top and each successive generation arranged in horizontal rows below.
- **SVG Output**: Export charts as SVG (Scalable Vector Graphics) for easy integration into web pages or further editing in vector graphic editors.
- **Text-Based Parser**: Easily parse textual representations of descendant lists into `gtree` structures, allowing you to quickly generate charts from text-based genealogical data.

## Usage


## Getting Started

Run the following in the directory containing your project's `go.mod` file:

```sh
go get github.com/iand/gtree@latest
```

Documentation is at [https://pkg.go.dev/github.com/iand/gtree](https://pkg.go.dev/github.com/iand/gtree)

## License

This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying [`UNLICENSE`](UNLICENSE) file.

