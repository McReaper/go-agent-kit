package lib

import "context"

func Work() context.Context {
	ctx := context.Background() // want `M11: context.Background starts a new root`
	_ = context.TODO()          // want `M11: context.TODO starts a new root`
	return ctx
}

func Pass(ctx context.Context) context.Context { return ctx }
