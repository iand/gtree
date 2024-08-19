package gtree

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var reLine = regexp.MustCompile(`^(\s*)(\d+|sp|\+)(?:\.)?\s*(.+)$`)

// A Parser parses a textual descendent list.
//
// A descendant list is a list of person entries each consisting of a prefix followed by
// detail text. Each entry begins on a new line. Leading whitespace is significant
// (for spouse disambiguation) but trailing is not. Lines consisting only of whitespace
// are ignored.
//
// The prefix of each entry denotes the relationship of the person to an earlier person.
// A prefix text may be a generation number followed by a dot (.) which indicates
// the position of the person in the family tree relative to the root ancestor. A
// generation number of 1 indicates the root ancestor, 2 indicates their children, and
// so on.
//
// Alternatively the prefix may be the two characters 'sp' or the single
// character '+' which indicates that the person is the spouse of the preceding
// numbered person with equal or lesser indentation.

// The text after the prefix is the person's name followed by optional details,
// such as birth, death, marriage, and other life events. The text may wrap
// onto subsequent lines until a line with a generation number or spouse prefix is
// encountered.
//
// The detail text is delimited from the name by one of the following:
//
//   - an open paranthesis '('
//   - one of the following event abbreviations 'b', 'm', 'd', 'bap', 'mar', 'marr'
//     or 'bur' followed by a dot '.' or colon ':'
//   - one of the following event keywords 'born', 'baptised', 'married', 'died', 'buried'
//
// All text up to the delimiter is to be the name of the person, the remaining text is
// the detail.
//
// The name and the detail are trimmed to remove leading and trailing whitespace. Outer
// matching parantheses are removed from the detail text before trimming.
//
// Any semicolons ';' within the detail text are treated as line breaks, resulting in
// multiple lines of text.
//
// Identifiers are assigned using the line number of the person's entry. People in a family
// group are placed in the order the lines are read from the input.
type Parser struct{}

func (p *Parser) Parse(ctx context.Context, r io.Reader) (*DescendantChart, error) {
	s := bufio.NewScanner(r)
	lineno := 0

	type entry struct {
		lineno     int
		indent     int
		generation int
		isSpouse   bool
		text       string
		person     *DescendantPerson
	}

	entries := []*entry{}

	var cur *entry
	for s.Scan() {
		lineno++
		line := strings.TrimRightFunc(s.Text(), unicode.IsSpace)
		if len(line) == 0 {
			continue
		}
		matches := reLine.FindStringSubmatch(line)
		if len(matches) == 4 {
			// start a new entry
			cur = &entry{
				lineno: lineno,
				indent: len(matches[1]),
				text:   strings.TrimSpace(matches[3]),
				person: &DescendantPerson{
					ID:      len(entries) + 1,
					Details: p.parseDetails(ctx, strings.TrimSpace(matches[3])),
				},
			}

			if matches[2] == "sp" || matches[2] == "+" {
				cur.isSpouse = true
			} else {
				gen, err := strconv.Atoi(matches[2])
				if err != nil {
					return nil, fmt.Errorf("line %d: malformed generation number: %w", lineno, err)
				}
				cur.generation = gen
			}

			entries = append(entries, cur)
		} else {
			if cur == nil {
				return nil, fmt.Errorf("line %d: malformed entry", lineno)
			}
			cur.text += " " + strings.TrimSpace(line)
		}
	}
	if s.Err() != nil {
		return nil, s.Err()
	}

	lin := new(DescendantChart)

	ppl := []*entry{}
	for _, e := range entries {
		if len(ppl) == 0 {
			if e.isSpouse {
				return nil, fmt.Errorf("line %d: spouse encountered before first person", e.lineno)
			}
			if e.generation != 1 {
				return nil, fmt.Errorf("line %d: first person must have generation number 1", e.lineno)
			}
			if lin.Root == nil {
				lin.Root = e.person
			}
			ppl = append(ppl, e)
		} else {
			prev := ppl[len(ppl)-1]
			if e.isSpouse {
				for e.indent < prev.indent && len(ppl) > 0 {
					ppl = ppl[:len(ppl)-1]
					if len(ppl) == 0 {
						return nil, fmt.Errorf("line %d: invalid person indent", e.lineno)
					}
					prev = ppl[len(ppl)-1]
				}
				// start a family
				fam := &DescendantFamily{
					Other: e.person,
				}
				prev.person.Families = append(prev.person.Families, fam)
			} else {
				for e.generation <= prev.generation && len(ppl) > 0 {
					ppl = ppl[:len(ppl)-1]
					if len(ppl) == 0 {
						return nil, fmt.Errorf("line %d: invalid person generation number", e.lineno)
					}
					prev = ppl[len(ppl)-1]
				}
				if e.generation == prev.generation+1 {
					// child
					if len(prev.person.Families) == 0 {
						// child of first family
						fam := &DescendantFamily{
							Children: []*DescendantPerson{e.person},
						}
						prev.person.Families = append(prev.person.Families, fam)
					} else {
						fam := prev.person.Families[len(prev.person.Families)-1]
						fam.Children = append(fam.Children, e.person)
					}

					// child is new current person entry
					ppl = append(ppl, e)
				} else {
					return nil, fmt.Errorf("line %d: expected person with generation number %d, got %d", e.lineno, e.generation+1, e.generation)
				}
			}
		}
	}

	return lin, nil
}

// parseDetails parses a person's details from a line
func (p *Parser) parseDetails(ctx context.Context, s string) []string {
	isDetailStart := func(s string) bool {
		return strings.HasPrefix(s, "(") ||
			strings.HasPrefix(s, "b.") ||
			strings.HasPrefix(s, "m.") ||
			strings.HasPrefix(s, "d.") ||
			strings.HasPrefix(s, "b:") ||
			strings.HasPrefix(s, "m:") ||
			strings.HasPrefix(s, "d:")
	}

	cleanLines := func(name, detail string) []string {
		if name != "" && detail == "" {
			br := strings.IndexByte(name, '(')
			if br == -1 {
				return []string{strings.TrimSpace(name)}
			}

			detail = name[br:]
			name = name[:br]
		}

		if strings.HasPrefix(detail, "(") && strings.HasSuffix(detail, ")") {
			detail = detail[1 : len(detail)-1]
		}

		lines := strings.Split(detail, ";")
		for i := range lines {
			lines[i] = strings.TrimSpace(lines[i])
		}

		name = strings.TrimSpace(name)
		if name == "" {
			return lines
		}

		ret := make([]string, 0, len(lines)+1)
		ret = append(ret, name)
		ret = append(ret, lines...)
		return ret
	}

	s = strings.TrimSpace(s)
	if isDetailStart(s) {
		return cleanLines("", s)
	}

	pos := 0
	sp := strings.IndexByte(s[pos:], ' ')
	for sp != -1 {
		pos += sp + 1

		if isDetailStart(s[pos:]) {
			return cleanLines(s[:pos-1], s[pos:])
		}

		sp = strings.IndexByte(s[pos:], ' ')
	}

	return cleanLines(s, "")
}
