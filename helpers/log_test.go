package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mohamedation/GoSorter/model"
)

func TestCLILogger_Levels(t *testing.T) {
	logger := &CLILogger{}
	logPath := filepath.Join(t.TempDir(), "out.log")

	cfg := model.Config{Verbose: true, LogFilePath: logPath}
	logger.Log(cfg, Normal, "normal\n")
	logger.Log(cfg, Info, "info\n")
	logger.Log(cfg, Debug, "debug\n")
	logger.Log(cfg, Error, "error\n")
	logger.Log(cfg, Raw, "raw\n")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	content := string(data)
	for _, want := range []string{"normal", "[INFO]", "[DEBUG]", "[ERROR]", "raw"} {
		if !strings.Contains(content, want) {
			t.Fatalf("log missing %q: %s", want, content)
		}
	}
}

func TestCLILogger_SilentAndNonVerbose(t *testing.T) {
	logger := &CLILogger{}
	logPath := filepath.Join(t.TempDir(), "silent.log")

	silent := model.Config{Silent: true, LogFilePath: logPath}
	logger.Log(silent, Info, "silent-info\n")
	logger.Log(silent, Error, "silent-error\n")

	quiet := model.Config{Verbose: false}
	logger.Log(quiet, Debug, "should-not-print\n")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "silent-info") {
		t.Fatalf("expected silent path to still write log file: %s", data)
	}
}

func TestWriteLogFile_NoPath(t *testing.T) {
	writeLogFile(model.Config{}, "noop", &CLILogger{})
}
