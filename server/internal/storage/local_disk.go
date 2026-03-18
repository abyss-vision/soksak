package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalDiskProvider stores objects on the local filesystem under a base directory.
type LocalDiskProvider struct {
	baseDir string
}

// NewLocalDiskProvider creates a LocalDiskProvider rooted at baseDir.
func NewLocalDiskProvider(baseDir string) *LocalDiskProvider {
	return &LocalDiskProvider{baseDir: filepath.Clean(baseDir)}
}

// resolveKey validates and resolves a storage key to an absolute path within baseDir.
func (p *LocalDiskProvider) resolveKey(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("storage: key must not be empty")
	}
	// Normalize and reject traversal attempts.
	clean := filepath.Clean(key)
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", fmt.Errorf("storage: invalid key %q", key)
	}
	full := filepath.Join(p.baseDir, clean)
	// Double-check the resolved path stays within baseDir.
	if !strings.HasPrefix(full, p.baseDir+string(filepath.Separator)) && full != p.baseDir {
		return "", fmt.Errorf("storage: key %q escapes base directory", key)
	}
	return full, nil
}

// Put writes r to baseDir/key atomically using a temp file.
func (p *LocalDiskProvider) Put(ctx context.Context, key string, r io.Reader) error {
	full, err := p.resolveKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("storage.Put: mkdir: %w", err)
	}

	tmp := full + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("storage.Put: create temp: %w", err)
	}
	defer func() { _ = os.Remove(tmp) }()

	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return fmt.Errorf("storage.Put: write: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("storage.Put: close: %w", err)
	}
	if err := os.Rename(tmp, full); err != nil {
		return fmt.Errorf("storage.Put: rename: %w", err)
	}
	return nil
}

// Get returns a ReadCloser for the object at key.
func (p *LocalDiskProvider) Get(_ context.Context, key string) (io.ReadCloser, error) {
	full, err := p.resolveKey(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage.Get: object not found: %q", key)
		}
		return nil, fmt.Errorf("storage.Get: %w", err)
	}
	return f, nil
}

// Delete removes the object at key. Idempotent — not-found is not an error.
func (p *LocalDiskProvider) Delete(_ context.Context, key string) error {
	full, err := p.resolveKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage.Delete: %w", err)
	}
	return nil
}

// List returns all keys under prefix within baseDir.
func (p *LocalDiskProvider) List(_ context.Context, prefix string) ([]string, error) {
	searchDir := p.baseDir
	if prefix != "" {
		clean := filepath.Clean(prefix)
		if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
			return nil, fmt.Errorf("storage.List: invalid prefix %q", prefix)
		}
		searchDir = filepath.Join(p.baseDir, clean)
	}

	var keys []string
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			rel, relErr := filepath.Rel(p.baseDir, path)
			if relErr == nil {
				keys = append(keys, rel)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("storage.List: %w", err)
	}
	return keys, nil
}
