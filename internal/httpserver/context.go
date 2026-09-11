package httpserver

import (
	"context"

	"github.com/VitoMinheere/TrackDate/internal/store"
)

type ctxKey int

const userCtxKey ctxKey = 0

func withUserContext(ctx context.Context, u store.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// userFromContext returns the logged-in user and whether one is present.
func userFromContext(ctx context.Context) (store.User, bool) {
	u, ok := ctx.Value(userCtxKey).(store.User)
	return u, ok
}
