package ports

import "context"

type DependencyChecker interface {
	Name() string
	Check(ctx context.Context) error
}
