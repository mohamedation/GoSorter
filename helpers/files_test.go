package helpers

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	// Windows does not preserve Unix permission bits the same way
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
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

func TestFolderExistsAndFileExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !FolderExists(dir) {
		t.Fatal("FolderExists should be true for directory")
	}
	if FolderExists(filepath.Join(dir, "missing")) {
		t.Fatal("FolderExists should be false for missing path")
	}
	if !FileExists(file) {
		t.Fatal("FileExists should be true for file")
	}
	if FileExists(dir) {
		t.Fatal("FileExists should be false for directory")
	}
	if FileExists(filepath.Join(dir, "missing.txt")) {
		t.Fatal("FileExists should be false for missing file")
	}
}

func TestFormatPath(t *testing.T) {
	verbose := model.Config{Verbose: true}
	if got := FormatPath("photo.jpg", verbose); got != "photo.jpg" {
		t.Fatalf("verbose short = %q", got)
	}
	if got := FormatPath(filepath.Join("dir", "photo.jpg"), verbose); got != filepath.Join("dir", "photo.jpg") {
		t.Fatalf("verbose with dir = %q", got)
	}

	short := FormatPath("shortname.jpg", model.Config{})
	if short != "shortname.jpg" {
		t.Fatalf("short name = %q", short)
	}

	longName := strings.Repeat("a", 40) + ".jpg"
	got := FormatPath(longName, model.Config{})
	if !strings.Contains(got, "...") || !strings.HasSuffix(got, ".jpg") {
		t.Fatalf("expected truncated name, got %q", got)
	}
}

func TestMoveDuplicateFileAndExtractedArchive(t *testing.T) {
	dir := t.TempDir()
	cfg := model.Config{Silent: true}
	logger := &CLILogger{}

	if err := os.WriteFile(filepath.Join(dir, "copy.txt"), []byte("dup"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := MoveDuplicateFile(dir, "copy.txt", filepath.Join(dir, "orig.txt"), cfg, logger); err != nil {
		t.Fatalf("MoveDuplicateFile: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "Duplicates"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one duplicate file, err=%v entries=%v", err, entries)
	}

	if err := os.WriteFile(filepath.Join(dir, "pack.zip"), []byte("zip"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := MoveExtractedArchive(dir, "pack.zip", cfg, logger); err != nil {
		t.Fatalf("MoveExtractedArchive: %v", err)
	}
	if !FileExists(filepath.Join(dir, "Archives-Extracted", "pack.zip")) {
		t.Fatal("expected archive in Archives-Extracted")
	}
}

func TestMoveFileToTargetFolder_NewAndConflict(t *testing.T) {
	dir := t.TempDir()
	cfg := model.Config{Silent: true, MaxHashFileSizeMB: 1}
	logger := &CLILogger{}

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := MoveFileToTargetFolder(dir, "a.txt", "Documents", cfg, logger); err != nil {
		t.Fatalf("move: %v", err)
	}
	if !FileExists(filepath.Join(dir, "Documents", "a.txt")) {
		t.Fatal("expected a.txt in Documents")
	}

	// different content conflict -> rename with (n)
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("two"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := MoveFileToTargetFolder(dir, "a.txt", "Documents", cfg, logger); err != nil {
		t.Fatalf("conflict move: %v", err)
	}
	if !FileExists(filepath.Join(dir, "Documents", "a(1).txt")) {
		t.Fatal("expected renamed conflict file a(1).txt")
	}

	// identical content -> remove source
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := MoveFileToTargetFolder(dir, "a.txt", "Documents", cfg, logger); err != nil {
		t.Fatalf("identical move: %v", err)
	}
	if FileExists(filepath.Join(dir, "a.txt")) {
		t.Fatal("identical source should be removed")
	}
}

func TestShouldCopyFallback_MoreCases(t *testing.T) {
	if shouldCopyFallback(nil) {
		t.Fatal("nil should not fallback")
	}
	if !shouldCopyFallback(&os.PathError{Err: syscall.EXDEV}) {
		t.Fatal("PathError(EXDEV) should fallback")
	}
	if !shouldCopyFallback(os.ErrPermission) {
		t.Fatal("permission should fallback")
	}
	if shouldCopyFallback(syscall.ENOENT) {
		t.Fatal("ENOENT should not fallback")
	}
}

func TestCopyFileThenRemove_MissingSource(t *testing.T) {
	dir := t.TempDir()
	err := copyFileThenRemove(filepath.Join(dir, "missing"), filepath.Join(dir, "dst"), model.Config{Silent: true}, &CLILogger{})
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}
