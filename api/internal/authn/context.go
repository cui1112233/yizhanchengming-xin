package authn

import "context"

type currentUserKey struct{}

func WithCurrentUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, currentUserKey{}, user)
}

func CurrentUser(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(currentUserKey{}).(User)
	return user, ok && user.ID > 0
}
