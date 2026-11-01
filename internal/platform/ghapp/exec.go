package ghapp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Identity is the app's bot user, who commits and pushes.
type Identity struct {
	Login string
	ID    int64
}

// Email is the bot's no-reply address, which GitHub ties to its account.
func (i Identity) Email() string {
	return strconv.FormatInt(i.ID, 10) + "+" + i.Login + "@users.noreply.github.com"
}

// Env returns base with the token for gh and git, and the bot as author
// and committer.
//
// WHY: git reads the credential through a helper that echoes $GH_TOKEN,
// so the token sits in the child's environment only, never on a command
// line or in a git config file.
func Env(base []string, tok Token, id Identity) []string {
	helper := `!f() { echo username=x-access-token; echo "password=$GH_TOKEN"; }; f`
	return append(base,
		"GH_TOKEN="+tok.Value,
		"GITHUB_TOKEN="+tok.Value,
		"GIT_CONFIG_COUNT=2",
		"GIT_CONFIG_KEY_0=credential.helper", "GIT_CONFIG_VALUE_0=",
		"GIT_CONFIG_KEY_1=credential.helper", "GIT_CONFIG_VALUE_1="+helper,
		"GIT_AUTHOR_NAME="+id.Login, "GIT_AUTHOR_EMAIL="+id.Email(),
		"GIT_COMMITTER_NAME="+id.Login, "GIT_COMMITTER_EMAIL="+id.Email(),
	)
}

// Exec runs argv as the app, its standard streams the caller's.
func Exec(ctx context.Context, argv []string, tok Token, id Identity) error {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = Env(os.Environ(), tok, id)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", argv[0], err)
	}
	return nil
}
