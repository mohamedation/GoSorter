package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mohamedation/GoSorter/model"
)

func TestHashFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := model.Config{Silent: true}
	hash, err := HashFile(path, 1024, cfg, &CLILogger{})
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	hash2, err := HashFile(path, 1024, cfg, nil)
	if err != nil || hash2 != hash {
		t.Fatalf("nil logger hash mismatch: %v %q", err, hash2)
	}
}

func TestHashFile_TooLarge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.txt")
	if err := os.WriteFile(path, []byte("0123456789"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := model.Config{Silent: true}
	hash, err := HashFile(path, 4, cfg, &CLILogger{})
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if hash != "" {
		t.Fatalf("expected empty hash for oversized file, got %q", hash)
	}
}

func TestHashFile_Missing(t *testing.T) {
	cfg := model.Config{Silent: true}
	_, err := HashFile(filepath.Join(t.TempDir(), "missing.txt"), 1024, cfg, &CLILogger{})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPartialHashFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("partial-hash-content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg := model.Config{Silent: true}
	hash, err := PartialHashFile(path, 8, cfg, &CLILogger{})
	if err != nil {
		t.Fatalf("PartialHashFile: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty partial hash")
	}
}

func TestPartialHashFile_Missing(t *testing.T) {
	cfg := model.Config{Silent: true}
	_, err := PartialHashFile(filepath.Join(t.TempDir(), "missing.txt"), 8, cfg, &CLILogger{})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
