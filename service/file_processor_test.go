// Package service - tests
package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mohamedation/GoSorter/helpers"
	"github.com/mohamedation/GoSorter/model"
)

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

// test for transparent PNG detection, but needs valid PNG data so still WIP
// func TestFileProcessor_TransparentPNGDetection(t *testing.T) {
// 	tempDir, err := os.MkdirTemp("", "gosorter_test_png")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp dir: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	// generated with AI, but still not valid, yet
// 	// Transparent PNG (1x1 pixel with alpha=0)
// 	transparentPNG := []byte{
// 		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
// 		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
// 		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 pixels
// 		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, // RGBA, no interlace
// 		0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41, // IDAT chunk
// 		0x54, 0x78, 0x9C, 0x62, 0x00, 0x02, 0x00, 0x00, // Compressed data (transparent pixel)
// 		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
// 		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, // IEND chunk
// 		0x42, 0x60, 0x82,
// 	}

// 	// Opaque PNG (1x1 pixel with alpha=255)
// 	opaquePNG := []byte{
// 		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
// 		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
// 		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 pixels
// 		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, // RGBA, no interlace
// 		0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41, // IDAT chunk
// 		0x54, 0x78, 0x9C, 0x62, 0xF8, 0x0F, 0x00, 0x00, // Compressed data (opaque pixel)
// 		0xFF, 0xFF, 0x01, 0x00, 0x01, 0x00, 0x01, 0x02,
// 		0x7E, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, // IEND chunk
// 		0x44, 0xAE, 0x42, 0x60, 0x82,
// 	}

// 	transparentPath := filepath.Join(tempDir, "transparent.png")
// 	opaquePath := filepath.Join(tempDir, "opaque.png")

// 	if err := os.WriteFile(transparentPath, transparentPNG, 0644); err != nil {
// 		t.Fatalf("Failed to create transparent PNG file: %v", err)
// 	}

// 	if err := os.WriteFile(opaquePath, opaquePNG, 0644); err != nil {
// 		t.Fatalf("Failed to create opaque PNG file: %v", err)
// 	}

// 	config := &model.Config{
// 		MoveDuplicates:        false,
// 		Verbose:               true,
// 		Silent:                false,
// 		DetectTransparentPNGs: true,
// 	}
// 	stats := &model.Stats{StartTime: time.Now()}

// 	processor := NewFileProcessor(config, stats)
// 	ctx := context.Background()

// 	err = processor.ProcessDirectory(ctx, tempDir)
// 	if err != nil {
// 		t.Errorf("ProcessDirectory with PNG detection failed: %v", err)
// 	}

// 	if stats.GetTotalFiles() != 2 {
// 		t.Errorf("Expected 2 files processed, got %d", stats.GetTotalFiles())
// 	}
// }

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
