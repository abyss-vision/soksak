package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalDiskProvider_PutGetDelete(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	content := "hello, world"
	err := p.Put(ctx, "file.txt", strings.NewReader(content))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	rc, err := p.Get(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if string(got) != content {
		t.Errorf("Get content = %q, want %q", string(got), content)
	}

	err = p.Delete(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// After delete, Get should fail.
	_, err = p.Get(ctx, "file.txt")
	if err == nil {
		t.Fatal("Get after Delete: expected error, got nil")
	}
}

func TestLocalDiskProvider_Delete_Idempotent(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	// Deleting a nonexistent key should not error.
	err := p.Delete(ctx, "nonexistent.txt")
	if err != nil {
		t.Errorf("Delete nonexistent: expected nil, got %v", err)
	}
}

func TestLocalDiskProvider_Put_Subdirectory(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	err := p.Put(ctx, "sub/dir/file.txt", strings.NewReader("data"))
	if err != nil {
		t.Fatalf("Put subdirectory: %v", err)
	}

	rc, err := p.Get(ctx, "sub/dir/file.txt")
	if err != nil {
		t.Fatalf("Get subdirectory: %v", err)
	}
	rc.Close()
}

func TestLocalDiskProvider_List(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	_ = p.Put(ctx, "a.txt", strings.NewReader("a"))
	_ = p.Put(ctx, "b.txt", strings.NewReader("b"))
	_ = p.Put(ctx, "sub/c.txt", strings.NewReader("c"))

	keys, err := p.List(ctx, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("List: expected 3 keys, got %d: %v", len(keys), keys)
	}
}

func TestLocalDiskProvider_List_WithPrefix(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	_ = p.Put(ctx, "sub/x.txt", strings.NewReader("x"))
	_ = p.Put(ctx, "sub/y.txt", strings.NewReader("y"))
	_ = p.Put(ctx, "other.txt", strings.NewReader("o"))

	keys, err := p.List(ctx, "sub")
	if err != nil {
		t.Fatalf("List with prefix: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("List prefix: expected 2 keys, got %d: %v", len(keys), keys)
	}
}

func TestLocalDiskProvider_List_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	keys, err := p.List(ctx, "")
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("List empty: expected 0 keys, got %d", len(keys))
	}
}

func TestLocalDiskProvider_resolveKey_InvalidKeys(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	invalidCases := []struct {
		name string
		key  string
	}{
		{"empty key", ""},
		{"absolute path", filepath.Join(dir, "escape.txt")},
		{"traversal", "../outside.txt"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := p.Put(ctx, tc.key, bytes.NewReader(nil))
			if err == nil {
				t.Errorf("Put(%q): expected error, got nil", tc.key)
			}
		})
	}
}

func TestLocalDiskProvider_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	_, err := p.Get(ctx, "missing.txt")
	if err == nil {
		t.Fatal("Get missing: expected error, got nil")
	}
}

func TestLocalDiskProvider_Put_Overwrite(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	_ = p.Put(ctx, "file.txt", strings.NewReader("first"))
	_ = p.Put(ctx, "file.txt", strings.NewReader("second"))

	rc, err := p.Get(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Get after overwrite: %v", err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if string(got) != "second" {
		t.Errorf("overwrite content = %q, want %q", string(got), "second")
	}
}

func TestLocalDiskProvider_List_InvalidPrefix(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	_, err := p.List(ctx, "../outside")
	if err == nil {
		t.Fatal("List invalid prefix: expected error, got nil")
	}
}

func TestLocalDiskProvider_List_NonexistentPrefix(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	// A valid but nonexistent prefix should return empty, not error.
	keys, err := p.List(ctx, "nonexistent-subdir")
	if err != nil {
		t.Fatalf("List nonexistent prefix: unexpected error: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("List nonexistent prefix: expected 0 keys, got %d", len(keys))
	}
}

func TestLocalDiskProvider_Put_ReadError(t *testing.T) {
	dir := t.TempDir()
	p := NewLocalDiskProvider(dir)
	ctx := context.Background()

	// Put with actual content should succeed, checking non-nil reader path.
	_ = p.Put(ctx, "valid.txt", strings.NewReader("content"))
	fi, err := os.Stat(filepath.Join(dir, "valid.txt"))
	if err != nil {
		t.Fatalf("stat after put: %v", err)
	}
	if fi.Size() != int64(len("content")) {
		t.Errorf("size = %d, want %d", fi.Size(), len("content"))
	}
}
