package lr

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

func step() error { return errors.New("x") }

func Twice(ctx context.Context, log *slog.Logger) error {
	if err := step(); err != nil {
		log.ErrorContext(ctx, "step failed", slog.Any("err", err)) // want `M7: err is logged and returned`
		return fmt.Errorf("step: %w", err)
	}
	return nil
}

func Global(ctx context.Context) error {
	err := step()
	if err != nil {
		slog.ErrorContext(ctx, "failed", "err", err) // want `M7: err is logged and returned`
		return err
	}
	return nil
}

func Handled(ctx context.Context, log *slog.Logger) {
	if err := step(); err != nil {
		log.WarnContext(ctx, "step failed, retrying later", slog.Any("err", err))
	}
}

func Wrapped() error {
	if err := step(); err != nil {
		return fmt.Errorf("step: %w", err)
	}
	return nil
}
