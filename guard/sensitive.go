package guard

import (
	"path"
	"strings"
)

func sensitive(f File) []Finding {
	if f.Status == 'D' || !isSensitive(path.Base(f.Path)) {
		return nil
	}
	return []Finding{{
		"sensitive-file", f.Path, 0,
		"S1: keys, certificates and env files are never committed: load them at run time from the environment or a secret store",
	}}
}

func isSensitive(name string) bool {
	lower := strings.ToLower(name)
	switch {
	case lower == ".env", strings.HasPrefix(lower, ".env.") && lower != ".env.example":
		return true
	case strings.HasPrefix(lower, "id_rsa"), strings.HasPrefix(lower, "id_dsa"),
		strings.HasPrefix(lower, "id_ecdsa"), strings.HasPrefix(lower, "id_ed25519"):
		return !strings.HasSuffix(lower, ".pub")
	}
	switch path.Ext(lower) {
	case ".pem", ".key", ".pfx", ".p12", ".jks", ".keystore", ".ppk":
		return true
	}
	return false
}
