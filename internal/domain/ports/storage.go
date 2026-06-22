package ports

import "context"

type StorageProvider interface {
	Upload(ctx context.Context, objectName string, data []byte) (string, error)
}
