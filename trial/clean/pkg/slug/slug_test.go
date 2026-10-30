package slug_test

import (
	"strings"
	"testing"

	"example.com/notes/pkg/slug"
	"pgregory.net/rapid"
)

func TestMake(t *testing.T) {
	cases := map[string]string{
		"Hello World":         "hello-world",
		"  spaced  out  ":     "spaced-out",
		"Café crème":          "cafe-creme",
		"already-a-slug":      "already-a-slug",
		"Zone 51: the Return": "zone-51-the-return",
		"!!!":                 "",
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			if got := slug.Make(in); got != want {
				t.Errorf("Make(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

func TestMakeKeepsOnlyLowercaseASCIIAndSingleHyphens(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		got := slug.Make(rapid.String().Draw(t, "s"))
		if strings.Contains(got, "--") || strings.HasPrefix(got, "-") || strings.HasSuffix(got, "-") {
			t.Fatalf("hyphens out of place in %q", got)
		}
		for _, r := range got {
			if !(r == '-' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
				t.Fatalf("%q holds %q", got, r)
			}
		}
	})
}

func TestMakeIsIdempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		once := slug.Make(rapid.String().Draw(t, "s"))
		if twice := slug.Make(once); twice != once {
			t.Fatalf("Make(%q) = %q, not a fixed point", once, twice)
		}
	})
}
