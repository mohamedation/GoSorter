// Package service - tests
package service

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mohamedation/GoSorter/helpers"
	"github.com/mohamedation/GoSorter/model"
)

func writePNG(t *testing.T, path string, transparent bool) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if transparent {
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

func TestFileProcessor_ProcessDirectory(t *testing.T) {

	tempDir, err := os.MkdirTemp("", "gosorter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	testFiles := []string{"test1.jpg", "test2.txt", "test3.mp3"}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	config := &model.Config{
		MoveDuplicates: false,
		Verbose:        false,
		Silent:         true,
	}
	stats := &model.Stats{StartTime: time.Now()}

	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})
	ctx := context.Background()

	err = processor.ProcessDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ProcessDirectory failed: %v", err)
	}

	if stats.GetTotalFiles() != int64(len(testFiles)) {
		t.Errorf("Expected %d files processed, got %d", len(testFiles), stats.GetTotalFiles())
	}
}

func TestFileProcessor_ProcessDirectoryNonExistent(t *testing.T) {
	config := &model.Config{}
	stats := &model.Stats{StartTime: time.Now()}
	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})

	ctx := context.Background()
	err := processor.ProcessDirectory(ctx, "/nonexistent/directory")

	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestFileProcessor_ProcessFilesWithDuplicates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_test_duplicates")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	testContent := []byte("identical content for duplicate test")
	testFiles := []string{"duplicate1.txt", "duplicate2.txt", "original.txt"}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, testContent, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	config := &model.Config{
		MoveDuplicates: true,
		Verbose:        false,
		Silent:         true,
	}
	stats := &model.Stats{StartTime: time.Now()}

	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})
	ctx := context.Background()

	err = processor.ProcessDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ProcessDirectory with duplicates failed: %v", err)
	}

	if stats.GetTotalFiles() != int64(len(testFiles)) {
		t.Errorf("Expected %d files processed, got %d", len(testFiles), stats.GetTotalFiles())
	}

	if stats.GetDuplicatesMoved() == 0 {
		t.Error("Expected duplicates to be moved, but none were detected")
	}
}

func TestFileProcessor_ProcessFilesWithoutDuplicates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_test_no_duplicates")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	testFiles := map[string]string{
		"document.pdf": "PDF content",
		"image.jpg":    "JPEG content",
		"video.mp4":    "Video content",
		"unknown.xyz":  "Unknown extension content",
	}

	for file, content := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	config := &model.Config{
		MoveDuplicates: false,
		Verbose:        true,
		Silent:         false,
	}
	stats := &model.Stats{StartTime: time.Now()}

	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})
	ctx := context.Background()

	err = processor.ProcessDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ProcessDirectory without duplicates failed: %v", err)
	}

	if stats.GetTotalFiles() != int64(len(testFiles)) {
		t.Errorf("Expected %d files processed, got %d", len(testFiles), stats.GetTotalFiles())
	}

	if stats.GetFilesMoved() >= stats.GetTotalFiles() {
		t.Errorf("Expected some files to be skipped due to unknown extensions")
	}
}

func TestFileProcessor_TransparentPNGDetection(t *testing.T) {
	tempDir := t.TempDir()
	writePNG(t, filepath.Join(tempDir, "transparent.png"), true)
	writePNG(t, filepath.Join(tempDir, "opaque.png"), false)

	config := &model.Config{
		MoveDuplicates:        false,
		Silent:                true,
		DetectTransparentPNGs: true,
	}
	stats := &model.Stats{StartTime: time.Now()}
	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})

	if err := processor.ProcessDirectory(context.Background(), tempDir); err != nil {
		t.Fatalf("ProcessDirectory: %v", err)
	}

	if stats.GetTransparentPNGsMoved() != 1 {
		t.Fatalf("expected 1 transparent PNG moved, got %d", stats.GetTransparentPNGsMoved())
	}
	if _, err := os.Stat(filepath.Join(tempDir, "PNGs", "transparent.png")); err != nil {
		t.Fatalf("transparent PNG should be in PNGs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "Pictures", "opaque.png")); err != nil {
		t.Fatalf("opaque PNG should be in Pictures: %v", err)
	}
}

func TestFileProcessor_TransparentPNGDetectionDisabled(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_test_png_disabled")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	// AI generated, but maybe not needed
	simplePNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x62, 0xF8, 0x0F, 0x00, 0x00,
		0xFF, 0xFF, 0x01, 0x00, 0x01, 0x00, 0x01, 0x02,
		0x7E, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	testFiles := []string{"test1.png", "test2.png"}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, simplePNG, 0644); err != nil {
			t.Fatalf("Failed to create test PNG file %s: %v", file, err)
		}
	}

	config := &model.Config{
		MoveDuplicates:        false,
		Verbose:               false,
		Silent:                true,
		DetectTransparentPNGs: false,
	}
	stats := &model.Stats{StartTime: time.Now()}

	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})
	ctx := context.Background()

	err = processor.ProcessDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ProcessDirectory with PNG detection disabled failed: %v", err)
	}

	if stats.GetTotalFiles() != int64(len(testFiles)) {
		t.Errorf("Expected %d files processed, got %d", len(testFiles), stats.GetTotalFiles())
	}

	if stats.GetTransparentPNGsMoved() != 0 {
		t.Errorf("Expected no transparent PNGs counted when detection disabled, got %d", stats.GetTransparentPNGsMoved())
	}
}

func TestFileProcessor_ZipFileHandling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_test_zip")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	zipFileName := "archive.zip"
	extractedDirName := "archive"

	zipFilePath := filepath.Join(tempDir, zipFileName)
	extractedDirPath := filepath.Join(tempDir, extractedDirName)

	if err := os.WriteFile(zipFilePath, []byte("ZIP content"), 0644); err != nil {
		t.Fatalf("Failed to create test zip file: %v", err)
	}

	if err := os.Mkdir(extractedDirPath, 0750); err != nil {
		t.Fatalf("Failed to create extracted directory: %v", err)
	}

	config := &model.Config{
		MoveDuplicates: false,
		Verbose:        false,
		Silent:         true,
	}
	stats := &model.Stats{StartTime: time.Now()}

	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})
	ctx := context.Background()

	err = processor.ProcessDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ProcessDirectory with zip handling failed: %v", err)
	}

	if stats.GetTotalFiles() == 0 {
		t.Error("Expected at least one file to be processed")
	}
}

func TestFileProcessor_FailedMoveNotCountedAsSuccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_test_failed_move")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	// a file named Documents blocks the target folder, so the move must fail
	documentsPath := filepath.Join(tempDir, "Documents")
	if err := os.WriteFile(documentsPath, []byte("not a folder"), 0644); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}
	notePath := filepath.Join(tempDir, "note.txt")
	if err := os.WriteFile(notePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := &model.Config{
		MoveDuplicates: false,
		Verbose:        false,
		Silent:         true,
	}
	stats := &model.Stats{StartTime: time.Now()}

	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})
	ctx := context.Background()

	err = processor.ProcessDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ProcessDirectory failed: %v", err)
	}

	if stats.GetFilesMoved() != 0 {
		t.Errorf("Expected 0 files moved on failure, got %d", stats.GetFilesMoved())
	}
	if stats.GetErrorsCount() == 0 {
		t.Error("Expected failed move to increment error count")
	}
	if _, err := os.Stat(notePath); err != nil {
		t.Errorf("Source file should still exist after failed move: %v", err)
	}
}

func TestFileProcessor_IdenticalContentKeepsDestination(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_test_identical")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	content := []byte("identical content")
	picturesDir := filepath.Join(tempDir, "Pictures")
	if err := os.Mkdir(picturesDir, 0750); err != nil {
		t.Fatalf("Failed to create Pictures folder: %v", err)
	}

	dstPath := filepath.Join(picturesDir, "photo.jpg")
	srcPath := filepath.Join(tempDir, "photo.jpg")
	if err := os.WriteFile(dstPath, content, 0644); err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	oldTime := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(dstPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set destination mtime: %v", err)
	}

	config := &model.Config{
		MoveDuplicates: false,
		Verbose:        false,
		Silent:         true,
	}
	stats := &model.Stats{StartTime: time.Now()}

	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})
	ctx := context.Background()

	err = processor.ProcessDirectory(ctx, tempDir)
	if err != nil {
		t.Errorf("ProcessDirectory failed: %v", err)
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("Expected source file to be removed as a duplicate of the destination")
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Destination file should still exist: %v", err)
	}
	if !info.ModTime().Before(stats.StartTime) {
		t.Error("Destination file should have been kept, not replaced")
	}
}

func TestFileProcessor_DuplicatesOnly(t *testing.T) {
	tempDir := t.TempDir()
	content := []byte("same-bytes")
	for _, name := range []string{"keep-me.txt", "dup-a.txt", "dup-b.txt"} {
		if err := os.WriteFile(filepath.Join(tempDir, name), content, 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(tempDir, "unique.jpg"), []byte("jpg"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	config := &model.Config{
		MoveDuplicates: true,
		DuplicatesOnly: true,
		Silent:         true,
	}
	stats := &model.Stats{StartTime: time.Now()}
	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})

	if err := processor.ProcessDirectory(context.Background(), tempDir); err != nil {
		t.Fatalf("ProcessDirectory: %v", err)
	}

	if stats.GetDuplicatesMoved() < 1 {
		t.Fatalf("expected duplicates moved, got %d", stats.GetDuplicatesMoved())
	}
	// unique file should stay put in duplicates-only mode
	if _, err := os.Stat(filepath.Join(tempDir, "unique.jpg")); err != nil {
		t.Fatalf("unique file should remain in place: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "Pictures")); !os.IsNotExist(err) {
		t.Fatal("duplicates-only should not create extension folders for unique files")
	}
}

func TestFileProcessor_SameSizeDifferentContent(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("AAAA"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "b.txt"), []byte("BBBB"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	config := &model.Config{MoveDuplicates: true, Silent: true}
	stats := &model.Stats{StartTime: time.Now()}
	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})

	if err := processor.ProcessDirectory(context.Background(), tempDir); err != nil {
		t.Fatalf("ProcessDirectory: %v", err)
	}
	if stats.GetDuplicatesMoved() != 0 {
		t.Fatalf("different content should not be duplicates, got %d", stats.GetDuplicatesMoved())
	}
	if stats.GetFilesMoved() != 2 {
		t.Fatalf("expected both files sorted, got %d", stats.GetFilesMoved())
	}
}

func TestFileProcessor_ZipWithoutExtractedFolder(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "bundle.zip"), []byte("zip"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	config := &model.Config{Silent: true}
	stats := &model.Stats{StartTime: time.Now()}
	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})

	if err := processor.ProcessDirectory(context.Background(), tempDir); err != nil {
		t.Fatalf("ProcessDirectory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "Archives", "bundle.zip")); err != nil {
		t.Fatalf("zip without extracted folder should go to Archives: %v", err)
	}
}

func TestFileProcessor_ContextCancelled(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := &model.Config{Silent: true}
	stats := &model.Stats{StartTime: time.Now()}
	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})

	err := processor.ProcessDirectory(ctx, tempDir)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestFileProcessor_SkipsDirectories(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, "subdir"), 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "ok.pdf"), []byte("pdf"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	config := &model.Config{Silent: true}
	stats := &model.Stats{StartTime: time.Now()}
	processor := NewFileProcessor(config, stats, &helpers.CLILogger{})

	if err := processor.ProcessDirectory(context.Background(), tempDir); err != nil {
		t.Fatalf("ProcessDirectory: %v", err)
	}
	if stats.GetTotalFiles() != 1 {
		t.Fatalf("expected only files counted, got %d", stats.GetTotalFiles())
	}
}
