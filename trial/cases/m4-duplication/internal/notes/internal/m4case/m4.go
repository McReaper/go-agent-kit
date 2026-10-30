package m4case

import "strings"

func TitleWords(s string) []string {
	var out []string
	for w := range strings.FieldsSeq(s) {
		w = strings.Trim(w, ".,;:!?")
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		switch {
		case len(w) > 3:
			w = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		case strings.ToUpper(w) == w:
			w = strings.ToUpper(w)
		default:
			w = strings.ToLower(w)
		}
		out = append(out, strings.ReplaceAll(w, "_", " "))
	}
	return out
}

func HeadingWords(s string) []string {
	var out []string
	for w := range strings.FieldsSeq(s) {
		w = strings.Trim(w, ".,;:!?")
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		switch {
		case len(w) > 3:
			w = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		case strings.ToUpper(w) == w:
			w = strings.ToUpper(w)
		default:
			w = strings.ToLower(w)
		}
		out = append(out, strings.ReplaceAll(w, "_", " "))
	}
	return out
}
