package observability

import "context"

type userCorrelationKey struct{}

type userCorrelationState struct {
	userID int64
}

func WithUserCorrelation(ctx context.Context) context.Context {
	return context.WithValue(ctx, userCorrelationKey{}, &userCorrelationState{})
}

func SetUserID(ctx context.Context, userID int64) {
	if userID <= 0 {
		return
	}
	state, ok := ctx.Value(userCorrelationKey{}).(*userCorrelationState)
	if ok && state != nil {
		state.userID = userID
	}
}

func UserID(ctx context.Context) int64 {
	state, ok := ctx.Value(userCorrelationKey{}).(*userCorrelationState)
	if !ok || state == nil {
		return 0
	}
	return state.userID
}
