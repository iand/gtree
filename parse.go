package gtree

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var reLine = regexp.MustCompile(`^(\s*)(\d+|sp)\.\s*(.+)$`)

// A Parser parses a textual descendent list.
//
// A descendant list consists of a list of person entries, one per line. A person
// entry line consists of a prefix followed by detail text.
//
// The prefix of each line denotes the relationship of the person to an earlier person.
// A prefix consists of zero or more leading whitespace characters and some context
// text followed by a dot (period).
//
// A prefix may include leading whitespace which indicates the relationship
// of the person to the preceding persons.
//
// The context text may be a number which indicates that the person is a child of the
// first preceding person with less leading whitespace. The number value is not used.
// People in a family group are placed in the order the lines are read from the input.
//
// Alternatively the context text may be the two characters 'sp' which indicates
// that the person is the spouse of the first preceding person with equal or less
// leading whitespace.
//
// The remaining text after the prefix is parsed as follows to build a list of detail
// strings for the person:
//
//   - If no matching parentheses or semicolons are found in the line then the detail
//     list will consist of a single entry containing the text up to the end of the
//     line.
//   - Otherwise, the text is divided into segments by scanning left to right. A
//     semicolon starts a new segment. A left paranthesis also starts a new segment
//     that includes all text up to the matching right paranthesis(including semicolons
//     and other parantheses). The delimiting semicolons or parantheses do not form
//     part of the text segment.
//
// All entries in the detail list are trimmed to remove leading and trailing
// whitespace.
//
// Identifiers are assigned using the line number of the person's entry.
type Parser struct{}

func (p *Parser) Parse(ctx context.Context, r io.Reader) (*DescendantChart, error) {
	s := bufio.NewScanner(r)
	lineno := 0

	lin := new(DescendantChart)
	ppl := []*DescendantPerson{}
	for s.Scan() {
		lineno++
		line := s.Text()

		matches := reLine.FindStringSubmatch(line)
		if len(matches) != 4 {
			return nil, fmt.Errorf("line %d: malformed entry", lineno)
		}

		person := &DescendantPerson{ID: lineno, Details: p.parseDetails(ctx, matches[3])}

		// indent := len(matches[1])
		ctext := matches[2]
		if ctext == "sp" {
			if len(ppl) == 0 {
				return nil, fmt.Errorf("line %d: spouse seen before person", lineno)
			}
			// start a new family
			// TODO: check indentation

			fam := &DescendantFamily{
				Other: person,
			}
			ppl[len(ppl)-1].Families = append(ppl[len(ppl)-1].Families, fam)
		} else {
			_, err := strconv.Atoi(matches[2])
			if err != nil {
				return nil, fmt.Errorf("line %d: malformed sequence number: %w", lineno, err)
			}
			// TODO: check indentation
			if len(ppl) == 0 {
				if lin.Root == nil {
					lin.Root = person
				}
				ppl = append(ppl, person)
			} else {
				if len(ppl[len(ppl)-1].Families) == 0 {
					ppl[len(ppl)-1].Families = []*DescendantFamily{{}}
				}
				fam := ppl[len(ppl)-1].Families[len(ppl[len(ppl)-1].Families)-1]
				fam.Children = append(fam.Children, person)
			}

		}

	}
	if s.Err() != nil {
		return nil, s.Err()
	}

	return lin, nil
}

// parseDetails parses a person's details from a line
func (p *Parser) parseDetails(ctx context.Context, s string) []string {
	details := []string{}
	rs := []rune(s)
	last := 0
	for cur := 0; cur < len(rs); cur++ {
		if rs[cur] == ';' {
			// end of segment
			if last < cur {
				details = append(details, string(rs[last:cur]))
			}
			last = cur + 1
		} else if rs[cur] == '(' {
			// end of segment
			if last < cur {
				details = append(details, string(rs[last:cur]))
				last = cur + 1
			}

			// maybe start of a new segment if parantheses are balanced
			depth := 0
			for i := cur + 1; i < len(rs); i++ {
				if rs[i] == ')' {
					if depth == 0 {
						details = append(details, string(rs[cur+1:i]))
						cur = i + 1
						last = i + 1
						break
					}
					depth--
				} else if rs[i] == '(' {
					depth++
				}
			}
		}
	}
	if last < len(rs)-1 {
		details = append(details, string(rs[last:]))
	}

	for i := range details {
		details[i] = strings.TrimSpace(details[i])
	}
	return details
}
