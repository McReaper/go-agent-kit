---
name: secrets
description: How this app holds a credential (an API token, a password, a key) so it never reaches a log, an error, an output or the journal. Use when the app first reads a credential from the environment or a file, or when a log, an error or the journal could show one.
---

# Credentials are a `secret.Secret`

A credential is held in a `secret.Secret` from the moment it is read until the call that needs it. Printed, logged, encoded to JSON or written to the journal, it reads `[redacted]`; only `Reveal` gives the value, at the one place that sends it.

The first credential adds the package to the platform, with its test, in the pull request that needs it. The kit's proofs compile both and run them through every gate.

```go internal/platform/secret/secret.go
// Package secret holds credentials so they never reach a log, an error,
// an output or the journal by accident.
package secret

import "log/slog"

const redacted = "[redacted]"

// Secret is a credential. Printed, logged or encoded, it reads
// "[redacted]"; only Reveal gives its value.
type Secret struct {
	value string
}

// New holds value as a secret.
func New(value string) Secret {
	return Secret{value: value}
}

// Reveal returns the value, for the one call that sends it.
func (s Secret) Reveal() string {
	return s.value
}

// IsZero reports whether the secret is empty.
func (s Secret) IsZero() bool {
	return s.value == ""
}

// String keeps the value out of fmt's verbs.
func (Secret) String() string {
	return redacted
}

// GoString keeps the value out of %#v.
func (Secret) GoString() string {
	return redacted
}

// LogValue keeps the value out of slog.
func (Secret) LogValue() slog.Value {
	return slog.StringValue(redacted)
}

// MarshalText keeps the value out of JSON, the journal included.
func (Secret) MarshalText() ([]byte, error) {
	return []byte(redacted), nil
}
```

```go internal/platform/secret/secret_test.go
package secret_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/secret"
)

const value = "ghp_example0123456789"

func TestASecretShowsOnlyThroughReveal(t *testing.T) {
	s := secret.New(value)
	if s.Reveal() != value || s.IsZero() || !secret.New("").IsZero() {
		t.Fatalf("Reveal = %q, IsZero = %v", s.Reveal(), s.IsZero())
	}
	type config struct {
		Token secret.Secret `json:"token"`
	}
	encoded, err := json.Marshal(config{Token: s})
	if err != nil {
		t.Fatal(err)
	}
	var logged bytes.Buffer
	slog.New(slog.NewJSONHandler(&logged, nil)).InfoContext(t.Context(), "loaded", slog.Any("token", s))
	shown := map[string]string{
		"%v":    fmt.Sprintf("token=%v", s),
		"%s":    fmt.Sprintf("token=%s", s),
		"%+v":   fmt.Sprintf("config=%+v", config{Token: s}),
		"%#v":   fmt.Sprintf("token=%#v", s),
		"error": fmt.Errorf("login with %v: %w", s, errors.New("refused")).Error(),
		"json":  string(encoded),
		"slog":  logged.String(),
	}
	for how, text := range shown {
		if strings.Contains(text, value) || !strings.Contains(text, "[redacted]") {
			t.Errorf("%s shows %q", how, text)
		}
	}
}
```

## Where it comes from

- Read it once, at startup, in `internal/platform/config`: `Token secret.Secret`, from the environment (`<APP>_TOKEN`), never from a flag (a flag shows in the process list).
- Pass the `Secret` down; call `Reveal` only in the infra call that sends it.
- An operation's `Args` never holds a credential: arguments come from the agent and go to the journal.
