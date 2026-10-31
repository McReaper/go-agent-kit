package logcase

import (
	"context"
	"log/slog"
)

func Saved(ctx context.Context, log *slog.Logger, id string) {
	log.InfoContext(ctx, "note saved", slog.String("id", id))
}

func Imported(ctx context.Context, log *slog.Logger, id string) {
	log.InfoContext(ctx, "note saved", slog.String("id", id))
}

func Deleted(ctx context.Context, log *slog.Logger, id string) {
	log.InfoContext(ctx, "note deleted", slog.String("id", id))
}
