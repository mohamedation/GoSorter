package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mohamedation/GoSorter/model"
)

func TestRunHelp(t *testing.T) {
	code := run([]string{"-h"})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
}

func TestRunBadFlag(t *testing.T) {
	code := run([]string{"-bogus"})
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

func TestRunInvalidConfig(t *testing.T) {
	code := run([]string{"-v", "-s"})
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
}

func TestRunMissingDirectory(t *testing.T) {
	code := run([]string{"/nonexistent/gosorter-dir"})
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
}

func TestRunSortDirectory(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "note.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	code := run([]string{"-s", tempDir})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "Documents", "note.txt")); err != nil {
		t.Fatalf("expected file to be sorted into Documents: %v", err)
	}
}

func TestRunDuplicatesOnly(t *testing.T) {
	tempDir := t.TempDir()
	content := []byte("same")
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(tempDir, name), content, 0644); err != nil {
			t.Fatalf("create file: %v", err)
		}
	}

	code := run([]string{"-do", "-s", tempDir})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}

	entries, err := os.ReadDir(filepath.Join(tempDir, "Duplicates"))
	if err != nil {
		t.Fatalf("expected Duplicates folder: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one duplicate to be moved")
	}
}

func TestRunParseMaxHashSize(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	cases := [][]string{
		{"-d", "-S", "2G", "-s", tempDir},
		{"-d", "-S", "512M", "-s", tempDir},
		{"-S", "2G", "-s", tempDir}, // -S without -d should warn and continue
		{"-d", "-S", "10X", "-s", tempDir},
		{"-d", "-S", "notasize", "-s", tempDir},
	}
	for _, args := range cases {
		if code := run(args); code != 0 {
			t.Fatalf("run(%v) = %d, want 0", args, code)
		}
	}
}

func TestRunVerboseStats(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "pic.jpg"), []byte("jpg"), 0644); err != nil {
		t.Fatalf("create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "weird.xyz"), []byte("x"), 0644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	code := run([]string{"-v", tempDir})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
}

func TestPrintHelpContainsVersion(t *testing.T) {
	// smoke: printHelp should not panic
	printHelp()
}

func TestPrintStatsBranches(t *testing.T) {
	stats := &model.Stats{StartTime: time.Now().Add(-time.Second)}
	stats.IncrementTotalFiles()
	stats.IncrementFilesMoved()
	stats.IncrementDuplicatesMoved()
	stats.IncrementTransparentPNGsMoved()
	stats.IncrementUnknownExtensions(".xyz")
	stats.IncrementUnknownExtensions("")
	stats.IncrementErrors()

	logPath := filepath.Join(t.TempDir(), "stats.log")
	cfg := model.Config{
		Verbose:               true,
		MoveDuplicates:        true,
		DetectTransparentPNGs: true,
		LogFilePath:           logPath,
	}
	printStats(cfg, stats)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "GoSorter Stats") {
		t.Fatalf("log missing stats: %q", data)
	}

	// early return when neither verbose nor log
	printStats(model.Config{}, stats)
}
