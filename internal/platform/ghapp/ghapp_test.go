package ghapp_test

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/McReaper/go-agent-kit/internal/platform/ghapp"
)

func newKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestLoadKeyReadsAPKCS1PEM(t *testing.T) {
	key := newKey(t)
	path := filepath.Join(t.TempDir(), "app.pem")
	data := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ghapp.LoadKey(path)
	if err != nil || !got.Equal(key) {
		t.Fatalf("LoadKey = %v, want the key written", err)
	}
}

func TestLoadKeyRefusesAFileWithNoPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.pem")
	if err := os.WriteFile(path, []byte("not a key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ghapp.LoadKey(path); !errors.Is(err, ghapp.ErrNoKey) {
		t.Fatalf("err = %v, want ErrNoKey", err)
	}
}

func TestTheJWTIsSignedByTheKeyAndNamesTheIssuer(t *testing.T) {
	key := newKey(t)
	now := time.Unix(1_800_000_000, 0)
	tok, err := ghapp.JWT("Iv23example", key, now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT has %d parts", len(parts))
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, digest[:], sig); err != nil {
		t.Fatalf("signature: %v", err)
	}
	claims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf(`{"iat":%d,"exp":%d,"iss":"Iv23example"}`, now.Unix()-60, now.Unix()+540)
	if string(claims) != want {
		t.Errorf("claims = %s, want %s", claims, want)
	}
}

func reply(t *testing.T, w io.Writer, body string) {
	t.Helper()
	if _, err := io.WriteString(w, body); err != nil {
		t.Error(err)
	}
}

func fakeGitHub(t *testing.T) ghapp.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer the-jwt" && !strings.HasPrefix(r.URL.Path, "/users/") {
			http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
			return
		}
		switch r.Method + " " + r.URL.Path {
		case "GET /app/installations":
			reply(t, w, `[{"id":42,"account":{"login":"McReaper"},"repository_selection":"selected"}]`)
		case "POST /app/installations/42/access_tokens":
			reply(t, w, `{"token":"ghs_secret","expires_at":"2026-10-31T02:00:00Z"}`)
		case "GET /app":
			reply(t, w, `{"slug":"t7-companion-maintainer"}`)
		case "GET /users/t7-companion-maintainer[bot]":
			reply(t, w, `{"id":350539711}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return ghapp.Client{BaseURL: srv.URL, HTTP: srv.Client()}
}

func TestTheClientListsInstallationsAndMintsATokenWithTheJWT(t *testing.T) {
	c := fakeGitHub(t)
	ctx := context.Background()
	list, err := c.Installations(ctx, "the-jwt")
	if err != nil || len(list) != 1 || list[0].ID != 42 || list[0].Account.Login != "McReaper" {
		t.Fatalf("Installations = %+v, %v", list, err)
	}
	tok, err := c.Token(ctx, "the-jwt", 42)
	if err != nil || tok.Value != "ghs_secret" || tok.ExpiresAt.IsZero() {
		t.Fatalf("Token = %v, %v", tok.ExpiresAt, err)
	}
	slug, err := c.Slug(ctx, "the-jwt")
	if err != nil || slug != "t7-companion-maintainer" {
		t.Fatalf("Slug = %q, %v", slug, err)
	}
	id, err := c.UserID(ctx, slug+"[bot]")
	if err != nil || id != 350539711 {
		t.Fatalf("UserID = %d, %v", id, err)
	}
}

func TestABadJWTIsAnErrorThatSaysWhy(t *testing.T) {
	_, err := fakeGitHub(t).Installations(context.Background(), "stale")
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Bad credentials") {
		t.Fatalf("err = %v, want the status and GitHub's message", err)
	}
}

func TestTheTokenNeverPrintsOrLogs(t *testing.T) {
	tok := ghapp.Token{Value: "ghs_secret"}
	var logs bytes.Buffer
	slog.New(slog.NewTextHandler(&logs, nil)).InfoContext(t.Context(), "minted", slog.Any("token", tok))
	for _, out := range []string{fmt.Sprint(tok), fmt.Sprintf("%v", tok), logs.String()} {
		if strings.Contains(out, "ghs_secret") {
			t.Errorf("the token leaked: %s", out)
		}
	}
}

func TestTheEnvGivesGitTheTokenThroughAHelperAndTheBotAsAuthor(t *testing.T) {
	env := ghapp.Env([]string{"PATH=x"}, ghapp.Token{Value: "ghs_secret"}, ghapp.Identity{Login: "app[bot]", ID: 7})
	for _, want := range []string{
		"PATH=x", "GH_TOKEN=ghs_secret", "GIT_CONFIG_KEY_1=credential.helper",
		"GIT_AUTHOR_NAME=app[bot]", "GIT_COMMITTER_EMAIL=7+app[bot]@users.noreply.github.com",
	} {
		if !slices.Contains(env, want) {
			t.Errorf("env lacks %s", want)
		}
	}
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "ghs_secret") != 2 {
		t.Errorf("the token is in the env once for gh and once for GITHUB_TOKEN, nowhere else: %s", data)
	}
}
