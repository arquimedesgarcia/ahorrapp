package ports

import "context"

type Cache interface {
	Ping(ctx context.Context) error
}
