package check

import (
	"context"
	"fmt"
	"path"
	"slices"

	"github.com/McReaper/rumblestrip-go/internal/platform/psscriptanalyzer"
)

func psscriptanalyzerGate(ctx context.Context, o Options) Result {
	files, err := tracked(ctx, o.Dir)
	if err != nil {
		return failed(err)
	}
	scripts := slices.DeleteFunc(files, func(f string) bool {
		return !slices.Contains([]string{".ps1", ".psm1", ".psd1"}, path.Ext(f))
	})
	if len(scripts) == 0 {
		return Result{}
	}
	fs, err := psscriptanalyzer.Analyze(ctx, o.Dir, scripts)
	if err != nil {
		return failed(err)
	}
	lines := make([]string, 0, len(fs))
	for _, f := range fs {
		lines = append(lines, fmt.Sprintf("%s:%d: %s: %s", f.File, f.Line, f.Rule, f.Message))
	}
	return found(lines)
}
