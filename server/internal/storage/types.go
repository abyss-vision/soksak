package storage

import (
	"context"
	"io"
)

// StorageProvider is the interface for object storage backends.
type StorageProvider interface {
	// Put stores a reader's content at the given key.
	Put(ctx context.Context, key string, r io.Reader) error
	// Get retrieves the content at the given key.
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes the object at the given key.
	Delete(ctx context.Context, key string) error
	// List returns all keys with the given prefix.
	List(ctx context.Context, prefix string) ([]string, error)
}
