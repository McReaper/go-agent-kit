package seccase

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"math/rand/v2"
	_ "net/http/pprof"
	"os"
	"os/exec"
)

func Sum(b []byte) [16]byte {
	return md5.Sum(b)
}

func Token() int {
	return rand.IntN(1 << 30)
}

func Insecure() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}

func Run(ctx context.Context, name string) error {
	return exec.CommandContext(ctx, name).Run()
}

func Git(ctx context.Context, arg string) error {
	return exec.CommandContext(ctx, "git", arg).Run()
}

func Load(path string) ([]byte, error) {
	return os.ReadFile(path)
}

const Warning = "safe‮txt.exe"

func Cafè() int {
	return 1
}
