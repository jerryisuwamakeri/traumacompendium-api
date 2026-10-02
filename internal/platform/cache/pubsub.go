package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Publisher adapts a real *redis.Client to the minimal, go-redis-free
// Publish(ctx, channel, message) error shape services like
// safereachservice.Publisher program against — so a service package
// doesn't need to import go-redis just to publish an event, and can be
// unit-tested with a trivial fake instead of a real Redis connection.
type Publisher struct {
	Client *redis.Client
}

func (p Publisher) Publish(ctx context.Context, channel string, message any) error {
	return p.Client.Publish(ctx, channel, message).Err()
}
