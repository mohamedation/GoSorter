package helpers

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/mohamedation/GoSorter/model"
)

func TestMoveFile_RenameSameDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_movefile_rename")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	src := filepath.Join(tempDir, "src.txt")
	dst := filepath.Join(tempDir, "dst.txt")
	content := []byte("hello move")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("Failed to create src: %v", err)
	}

	cfg := model.Config{Silent: true}
	if err := MoveFile(src, dst, cfg, &CLILogger{}); err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("Expected source to be gone after rename")
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("Failed to read dest: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Expected dest content %q, got %q", content, got)
	}
}

func TestCopyFileThenRemove_SuccessPreservesContentAndMode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_copy_ok")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	src := filepath.Join(tempDir, "src.txt")
	dst := filepath.Join(tempDir, "dst.txt")
	content := []byte("copied content")
	if err := os.WriteFile(src, content, 0600); err != nil {
		t.Fatalf("Failed to create src: %v", err)
	}

	cfg := model.Config{Silent: true}
	if err := copyFileThenRemove(src, dst, cfg, &CLILogger{}); err != nil {
		t.Fatalf("copyFileThenRemove failed: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("Expected source to be removed after copy")
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("Failed to read dest: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Expected dest content %q, got %q", content, got)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Failed to stat dest: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected mode 0600, got %v", info.Mode().Perm())
	}
}

func TestCopyFileThenRemove_FailedCopyLeavesDestUntouched(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_copy_fail")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// opening a directory succeeds on Unix, but reading it fails — copy aborts
	srcDir := filepath.Join(tempDir, "srcdir")
	if err := os.Mkdir(srcDir, 0750); err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}

	dst := filepath.Join(tempDir, "dst.txt")
	original := []byte("do not truncate me")
	if err := os.WriteFile(dst, original, 0644); err != nil {
		t.Fatalf("Failed to create dest: %v", err)
	}

	cfg := model.Config{Silent: true}
	err = copyFileThenRemove(srcDir, dst, cfg, &CLILogger{})
	if err == nil {
		t.Fatal("Expected copyFileThenRemove to fail when source is a directory")
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("Dest should still be readable: %v", err)
	}
	if string(got) != string(original) {
		t.Errorf("Dest was modified on failed copy: got %q, want %q", got, original)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}
	for _, e := range entries {
		if len(e.Name()) >= 10 && e.Name()[:10] == ".gosorter-" {
			t.Errorf("Temp file left behind: %s", e.Name())
		}
	}
}

func TestShouldCopyFallback_EXDEV(t *testing.T) {
	if !shouldCopyFallback(syscall.EXDEV) {
		t.Error("Expected EXDEV to trigger copy fallback")
	}
	if !shouldCopyFallback(&os.LinkError{Err: syscall.EXDEV}) {
		t.Error("Expected LinkError(EXDEV) to trigger copy fallback")
	}
	if shouldCopyFallback(errors.New("random error")) {
		t.Error("Expected unrelated error not to trigger copy fallback")
	}
	if !shouldCopyFallback(os.ErrExist) {
		t.Error("Expected ErrExist to trigger copy fallback")
	}
}
