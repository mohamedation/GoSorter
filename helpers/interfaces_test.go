package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mohamedation/GoSorter/model"
)

func TestOSFileOperations(t *testing.T) {
	ops := NewOSFileOperations()
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !ops.Exists(file) {
		t.Fatal("Exists should be true for file")
	}
	if ops.Exists(filepath.Join(dir, "missing.txt")) {
		t.Fatal("Exists should be false for missing file")
	}
	if !ops.FolderExists(dir) {
		t.Fatal("FolderExists should be true for directory")
	}

	sub := filepath.Join(dir, "sub")
	if err := ops.CreateFolder(sub); err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}

	cfg := model.Config{Silent: true}
	hash, err := ops.HashFile(file, cfg, &CLILogger{})
	if err != nil || hash == "" {
		t.Fatalf("HashFile: %v %q", err, hash)
	}

	dst := filepath.Join(dir, "b.txt")
	if err := ops.MoveFile(file, dst, cfg, &CLILogger{}); err != nil {
		t.Fatalf("MoveFile: %v", err)
	}
	if ops.Exists(file) || !ops.Exists(dst) {
		t.Fatal("MoveFile did not relocate file")
	}
}

func TestFileMover(t *testing.T) {
	dir := t.TempDir()
	cfg := &model.Config{Silent: true, MaxHashFileSizeMB: 1}
	mover := NewFileMover(NewOSFileOperations(), cfg)

	if err := os.WriteFile(filepath.Join(dir, "doc.txt"), []byte("doc"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := mover.MoveToTargetFolder(dir, "doc.txt", "Documents"); err != nil {
		t.Fatalf("MoveToTargetFolder: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "dup.txt"), []byte("dup"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := mover.MoveToDuplicates(dir, "dup.txt", filepath.Join(dir, "original.txt")); err != nil {
		t.Fatalf("MoveToDuplicates: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "arch.zip"), []byte("zip"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := mover.MoveExtractedArchive(dir, "arch.zip"); err != nil {
		t.Fatalf("MoveExtractedArchive: %v", err)
	}
}
