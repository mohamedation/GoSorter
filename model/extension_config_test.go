// Package model - extension configuration tests
package model

import (
	"os"
	"runtime"
	"testing"
)

func TestDefaultExtensionConfig(t *testing.T) {
	config := DefaultExtensionConfig()

	if config.ExtensionToFolder == nil {
		t.Error("ExtensionToFolder should not be nil")
	}

	if len(config.ExtensionToFolder) == 0 {
		t.Error("ExtensionToFolder should not be empty")
	}

	if config.ExtensionToFolder[".jpg"] != "Pictures" {
		t.Errorf("Expected .jpg to map to Pictures, got %s", config.ExtensionToFolder[".jpg"])
	}

	if config.ExtensionToFolder[".png"] != "Pictures" {
		t.Errorf("Expected .png to map to Pictures, got %s", config.ExtensionToFolder[".png"])
	}

	if config.ExtensionToFolder[".pdf"] != "PDFs" {
		t.Errorf("Expected .pdf to map to PDFs, got %s", config.ExtensionToFolder[".pdf"])
	}

	if config.ArchiveExtractedFolder != "Archives-Extracted" {
		t.Errorf("Expected ArchiveExtractedFolder to be Archives-Extracted, got %s", config.ArchiveExtractedFolder)
	}

	if config.DuplicatesFolder != "Duplicates" {
		t.Errorf("Expected DuplicatesFolder to be Duplicates, got %s", config.DuplicatesFolder)
	}

	if config.TransparentPNGFolder != "PNGs" {
		t.Errorf("Expected TransparentPNGFolder to be PNGs, got %s", config.TransparentPNGFolder)
	}
}

func TestGetTargetFolder(t *testing.T) {
	config := DefaultExtensionConfig()

	folder, exists := config.GetTargetFolder(".jpg")
	if !exists {
		t.Error("Expected .jpg to exist in config")
	}
	if folder != "Pictures" {
		t.Errorf("Expected .jpg to map to Pictures, got %s", folder)
	}

	folder, exists = config.GetTargetFolder(".unknown")
	if exists {
		t.Error("Expected .unknown to not exist in config")
	}
	if folder != "" {
		t.Errorf("Expected empty folder for unknown extension, got %s", folder)
	}
}

// setTestHome points os.UserHomeDir at dir for this test (HOME on Unix, USERPROFILE on Windows).
func setTestHome(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
		return
	}
	t.Setenv("HOME", dir)
}

func TestLoadExtensionConfig_NoUserConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	setTestHome(t, tempDir)

	config := LoadExtensionConfig()

	if config.ExtensionToFolder[".jpg"] != "Pictures" {
		t.Errorf("Expected .jpg to map to Pictures, got %s", config.ExtensionToFolder[".jpg"])
	}
}

func TestSaveAndLoadExtensionConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gosorter_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	setTestHome(t, tempDir)

	customConfig := &ExtensionConfig{
		ExtensionToFolder: map[string]string{
			".custom": "CustomFolder",
			".jpg":    "MyPictures",
		},
		ArchiveExtractedFolder: "MyArchives",
		DuplicatesFolder:       "MyDuplicates",
		TransparentPNGFolder:   "MyPNGs",
	}

	err = SaveExtensionConfig(customConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedConfig := LoadExtensionConfig()

	if loadedConfig.ExtensionToFolder[".custom"] != "CustomFolder" {
		t.Errorf("Expected .custom to map to CustomFolder, got %s", loadedConfig.ExtensionToFolder[".custom"])
	}

	if loadedConfig.ExtensionToFolder[".jpg"] != "MyPictures" {
		t.Errorf("Expected .jpg to map to MyPictures, got %s", loadedConfig.ExtensionToFolder[".jpg"])
	}

	if loadedConfig.ExtensionToFolder[".pdf"] != "PDFs" {
		t.Errorf("Expected .pdf to map to PDFs (default), got %s", loadedConfig.ExtensionToFolder[".pdf"])
	}

	if loadedConfig.ArchiveExtractedFolder != "MyArchives" {
		t.Errorf("Expected ArchiveExtractedFolder to be MyArchives, got %s", loadedConfig.ArchiveExtractedFolder)
	}
}

func TestMergeWithDefaults(t *testing.T) {
	userConfig := &ExtensionConfig{
		ExtensionToFolder: map[string]string{
			".custom": "CustomFolder",
		},
	}

	merged := mergeWithDefaults(userConfig)

	if merged.ExtensionToFolder[".custom"] != "CustomFolder" {
		t.Errorf("Expected .custom to map to CustomFolder, got %s", merged.ExtensionToFolder[".custom"])
	}

	if merged.ExtensionToFolder[".jpg"] != "Pictures" {
		t.Errorf("Expected .jpg to map to Pictures (default), got %s", merged.ExtensionToFolder[".jpg"])
	}

	if merged.ArchiveExtractedFolder != "Archives-Extracted" {
		t.Errorf("Expected ArchiveExtractedFolder to be Archives-Extracted (default), got %s", merged.ArchiveExtractedFolder)
	}
}
