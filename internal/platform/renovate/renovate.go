// Package renovate runs Renovate's configuration validator.
package renovate

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// Version is the Renovate the kit's proofs ran with; tools.txt names the
// same.
const Version = "44.181.0"

var message = regexp.MustCompile(`"message": ("(?:[^"\\]|\\.)*")`)

// Validate checks file, a repository's Renovate config in dir, with
// --strict: a warning or a needed migration fails it too. It returns the
// validator's messages, none when the config is valid.
func Validate(ctx context.Context, dir, file string) ([]string, error) {
	out, err := run.Command(ctx, dir, nil, "npx", "--yes", "--package", "renovate@"+Version,
		"renovate-config-validator", "--strict", "--no-global", file)
	if err != nil {
		return nil, err
	}
	if out.Code == 0 {
		return nil, nil
	}
	var msgs []string
	for _, m := range message.FindAllStringSubmatch(string(out.Stdout), -1) {
		if s, err := strconv.Unquote(m[1]); err == nil {
			msgs = append(msgs, s)
		}
	}
	if out.Code != 1 || len(msgs) == 0 {
		return nil, fmt.Errorf("renovate-config-validator %s: exit %d: %s", Version, out.Code, strings.TrimSpace(string(out.Stderr)+string(out.Stdout)))
	}
	return msgs, nil
}
