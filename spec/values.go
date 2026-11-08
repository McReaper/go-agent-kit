package spec

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// WHY: what a Then step expects is a quoted string or a number in its text,
// its doc string, or a cell of its table below the header: each is a value
// Rumblestrip alters once.
type value struct {
	raw   string
	alter func() Step
}

var (
	quoted = regexp.MustCompile(`"([^"]*)"`)
	number = regexp.MustCompile(`-?\b\d+(\.\d+)?\b`)
)

func values(st Step) []value {
	var out []value
	inQuotes := quoted.FindAllStringSubmatchIndex(st.Text, -1)
	for _, m := range inQuotes {
		out = append(out, textValue(st, m[2], m[3]))
	}
	for _, m := range number.FindAllStringIndex(st.Text, -1) {
		if !within(m[0], inQuotes) {
			out = append(out, textValue(st, m[0], m[1]))
		}
	}
	if st.DocString != nil {
		out = append(out, value{*st.DocString, func() Step {
			s := st
			doc := altered(*st.DocString)
			s.DocString = &doc
			return s
		}})
	}
	first := 1
	if len(st.Table) == 1 {
		first = 0
	}
	for r := first; r < len(st.Table); r++ {
		for c := range st.Table[r] {
			out = append(out, cellValue(st, r, c))
		}
	}
	return out
}

func textValue(st Step, start, end int) value {
	return value{st.Text[start:end], func() Step {
		s := st
		s.Text = st.Text[:start] + altered(st.Text[start:end]) + st.Text[end:]
		return s
	}}
}

func cellValue(st Step, r, c int) value {
	return value{st.Table[r][c], func() Step {
		s := st
		s.Table = make([][]string, len(st.Table))
		for i, row := range st.Table {
			s.Table[i] = append([]string(nil), row...)
		}
		s.Table[r][c] = altered(st.Table[r][c])
		return s
	}}
}

func within(at int, spans [][]int) bool {
	for _, s := range spans {
		if at >= s[0] && at < s[1] {
			return true
		}
	}
	return false
}

// WHY: the altered value keeps the original's shape, its length and the
// class of its last letter or digit, so a step whose pattern takes the one
// still takes the other: the scenario then fails on what the step checks,
// not on a step it no longer finds.
func altered(s string) string {
	r := []rune(s)
	for i := len(r) - 1; i >= 0; i-- {
		switch c := r[i]; {
		case c >= '0' && c <= '9':
			r[i] = '0' + (c-'0'+1)%10
			return string(r)
		case c >= 'a' && c <= 'z':
			r[i] = 'a' + (c-'a'+1)%26
			return string(r)
		case c >= 'A' && c <= 'Z':
			r[i] = 'A' + (c-'A'+1)%26
			return string(r)
		case unicode.IsLetter(c):
			r[i] = 'x'
			return string(r)
		}
	}
	return s + "x"
}

// Gherkin writes s as a feature file of its own, the one scenario run when
// Rumblestrip proves it.
func (s Scenario) Gherkin() string {
	var b strings.Builder
	b.WriteString("Feature: " + s.File + "\n\n")
	if len(s.Tags) > 0 {
		b.WriteString("  " + strings.Join(s.Tags, " ") + "\n")
	}
	b.WriteString("  Scenario: " + s.Name + "\n")
	for _, st := range s.Steps {
		writeStep(&b, st)
	}
	return b.String()
}

func keyword(k Kind) string {
	switch k {
	case When:
		return "When"
	case Then:
		return "Then"
	case Given:
	}
	return "Given"
}

func writeStep(b *strings.Builder, st Step) {
	fmt.Fprintf(b, "    %s %s\n", keyword(st.Kind), st.Text)
	if st.DocString != nil {
		fence := `"""`
		if strings.Contains(*st.DocString, fence) {
			fence = "```"
		}
		b.WriteString("      " + fence + "\n")
		for line := range strings.Lines(*st.DocString + "\n") {
			b.WriteString("      " + line)
		}
		b.WriteString("      " + fence + "\n")
	}
	for _, row := range st.Table {
		b.WriteString("      |")
		for _, c := range row {
			b.WriteString(" " + escapeCell(c) + " |")
		}
		b.WriteString("\n")
	}
}

func escapeCell(c string) string {
	return strings.NewReplacer(`\`, `\\`, "|", `\|`, "\n", `\n`).Replace(c)
}
