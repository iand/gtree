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
//
// The entry text may wrap onto subsequent lines until a line with a generation number or spouse prefix is
// encountered.
//
// The text after the prefix is the person's name followed by optional tags and detail text
// used for additional information such as birth, death, marriage, and other life events.
//
// If the SurnameSeparateLine field is true then the name will be parsed to detect
// a surname, which will be placed on a seperate heading line. If the name ends in
// one or more words delimted by slashes '/' then these will be used as the surname,
// otherwise the surname will be taken to be the last whole word after a space.
//
// All text up to the first tag delimiter or detail delimiter is to be the name of the person.
//
// Tags may be specified by prefixing words with a hash '#'. Multiple tags may be specified.
// Any tags must be occur between the name and the detail text delimiter.
//
// Detail text is delimited by parantheses '(' and ')'. All text between the parantheses is
// assumed to be the detail text.
//
// Any text after the closing detail paranthesis is ignored.
//
// The name and the detail text are trimmed to remove leading and trailing whitespace. Outer
// matching parantheses are removed from the detail text before trimming.
//
// Any semicolons ';' within the detail text are treated as line breaks, resulting in
// multiple lines of text.
//
// Identifiers are assigned using the line number of the person's entry. People in a family
// group are placed in the order the lines are read from the input.
type Parser struct {
	SurnameSeparateLine bool // if true the parser puts the surname on a second header line
}

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
			headings, details, tags := p.parseDetails(ctx, strings.TrimSpace(matches[3]))

			cur = &entry{
				lineno: lineno,
				indent: len(matches[1]),
				text:   strings.TrimSpace(matches[3]),
				person: &DescendantPerson{
					ID:       len(entries) + 1,
					Headings: headings,
					Details:  details,
					Tags:     tags,
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
func (p *Parser) parseDetails(ctx context.Context, s string) ([]string, []string, []string) {
	maybeSplitName := func(name string) []string {
		name = strings.TrimSpace(name)
		if !p.SurnameSeparateLine {
			return []string{name}
		}

		if strings.HasSuffix(name, "/") {
			sl := strings.IndexByte(name, '/')
			if sl != -1 {
				return []string{strings.TrimSpace(name[:sl]), name[sl+1 : len(name)-1]}
			}
		}

		sp := strings.LastIndexByte(name, ' ')
		if sp == -1 {
			return []string{name}
		}
		return []string{strings.TrimSpace(name[:sp]), name[sp:]}
	}

	cleanLines := func(name, detail string) ([]string, []string) {
		if name != "" && detail == "" {
			br := strings.IndexByte(name, '(')
			if br == -1 {
				return maybeSplitName(name), []string{}
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
			return []string{}, lines
		}

		return maybeSplitName(name), lines
	}

	var nametext, detailtext string
	var headings, details, tags []string

	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "(") {
		headings, details = cleanLines("", s)
		return headings, details, tags
	}

	pos := 0
	sp := strings.IndexByte(s[pos:], ' ')
	for sp != -1 {
		pos += sp + 1

		if strings.HasPrefix(s[pos:], "#") {
			if nametext == "" {
				nametext = s[:pos-1]
			}
			sp = strings.IndexByte(s[pos:], ' ')
			if sp == -1 {
				tags = append(tags, s[pos+1:])
				break
			}
			tags = append(tags, s[pos+1:pos+sp])
			continue
		}

		if strings.HasPrefix(s[pos:], "(") {
			if nametext == "" {
				nametext = s[:pos-1]
			}
			open := 1
			cl := pos + 1
			for ; cl < len(s); cl++ {
				if strings.HasPrefix(s[cl:], "(") {
					open++
					continue
				}
				if strings.HasPrefix(s[cl:], ")") {
					open--
					if open == 0 {
						break
					}
				}
			}

			if open == 0 {
				detailtext = s[pos+1 : cl]
			}
			headings, details = cleanLines(nametext, detailtext)
			return headings, details, tags
		}

		sp = strings.IndexByte(s[pos:], ' ')
	}

	if nametext == "" {
		nametext = s
	}

	headings, details = cleanLines(nametext, "")
	return headings, details, tags
}
