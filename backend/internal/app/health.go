package app

import "context"

type databaseHealthProbe interface {
	PingContext(context.Context) error
}

func databaseHealth(db databaseHealthProbe) func(context.Context) error {
	return func(ctx context.Context) error {
		return db.PingContext(ctx)
	}
}
