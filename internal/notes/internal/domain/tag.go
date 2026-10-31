package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

const MaxTagRunes = 32

var ErrEmptyTag = errors.New("a tag is empty")

type Tag string

// WHY: tags are case-insensitive, so the one stored form is lowercase and two
// spellings of a tag can never file a note apart.
func ParseTag(s string) (Tag, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return "", ErrEmptyTag
	}
	if n := utf8.RuneCountInString(t); n > MaxTagRunes {
		return "", fmt.Errorf("the tag %q has %d characters, the limit is %d", t, n, MaxTagRunes)
	}
	return Tag(t), nil
}

func ParseTags(ss []string) ([]Tag, error) {
	var tags []Tag
	for _, s := range ss {
		t, err := ParseTag(s)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(tags, t) {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

func (n Note) HasTag(t Tag) bool {
	return slices.Contains(n.Tags, t)
}
