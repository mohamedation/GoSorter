// Package model - extensions
package model

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ExtensionConfig struct {
	ExtensionToFolder      map[string]string `json:"extension_to_folder"`
	ArchiveExtractedFolder string            `json:"archives_extracted_folder"`
	DuplicatesFolder       string            `json:"duplicates_folder"`
	TransparentPNGFolder   string            `json:"transparent_png_folder"`
}

func DefaultExtensionConfig() *ExtensionConfig {
	return &ExtensionConfig{
		ExtensionToFolder: map[string]string{
			// Archives
			".zip": "Archives",
			".rar": "Archives",
			".tar": "Archives",
			".gz":  "Archives",

			// Audio/Music
			".mp3":  "Music",
			".wav":  "Music",
			".flac": "Music",
			".aac":  "Music",
			".ogg":  "Music",
			".m4a":  "Music",
			".wma":  "Music",
			".opus": "Music",
			".m4b":  "Music",
			".m4p":  "Music",

			// Design/Graphics
			".ai":   "Illustrator",
			".indd": "InDesign",
			".psd":  "Photoshop",

			// Documents
			".doc":  "Documents",
			".docx": "Documents",
			".odt":  "Documents",
			".txt":  "Documents",

			// Development
			".py": "Development",

			// Ebooks
			".mobi": "Ebooks",
			".epub": "Ebooks",
			".azw3": "Ebooks",

			// Executables/Apps
			".apk": "AndroidApps",
			".exe": "Executables",
			".deb": "Packages",

			// Images
			".bmp":  "Pictures",
			".gif":  "GIFs",
			".heic": "Pictures",
			".heif": "Pictures",
			".jpg":  "Pictures",
			".jpeg": "Pictures",
			".png":  "Pictures",
			".raw":  "RawImages",
			".svg":  "SVGs",
			".tiff": "Pictures",
			".tif":  "Pictures",
			".webp": "WebP",

			// Presentations
			".odp":  "Presentations",
			".ppt":  "Presentations",
			".pptx": "Presentations",

			// Sheets/Spreadsheets
			".csv":  "Sheets",
			".ods":  "Sheets",
			".xls":  "Sheets",
			".xlsx": "Sheets",

			// dmg and iso
			".dmg": "DiskImages",
			".iso": "ISOs",

			// Text/Data Files
			".json": "JSONs",
			".xml":  "XMLs",

			// Torrents
			".torrent": "Torrents",

			// Videos
			".avi":  "Videos",
			".mkv":  "Videos",
			".mp4":  "Videos",
			".mpg":  "Videos",
			".mpeg": "Videos",
			".webm": "Videos",

			// Virtual Machines
			".ova": "VirtualMachines",

			// VPN/Network Configs
			".ovpn": "Configs",

			// 3D Files
			".stl":   "3D/STLs",
			".3mf":   "3D/3MFs",
			".obj":   "3D/Objects",
			".gcode": "3D/GCode",

			// PDFs
			".pdf": "PDFs",
		},
		ArchiveExtractedFolder: "Archives-Extracted",
		DuplicatesFolder:       "Duplicates",
		TransparentPNGFolder:   "PNGs",
	}
}

func (ec *ExtensionConfig) GetTargetFolder(extension string) (string, bool) {
	folder, exists := ec.ExtensionToFolder[extension]
	return folder, exists
}

func LoadExtensionConfig() *ExtensionConfig {
	// user config first
	if config, err := loadUserExtensionConfig(); err == nil {
		return config
	}

	// default
	return DefaultExtensionConfig()
}

// config ~/.config/GoSorter/extension.json
// to do: test on linux. already tested on mac
func loadUserExtensionConfig() (*ExtensionConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".config", "GoSorter", "extension.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ExtensionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return mergeWithDefaults(&config), nil
}

// default values for missing configs
func mergeWithDefaults(userConfig *ExtensionConfig) *ExtensionConfig {
	defaultConfig := DefaultExtensionConfig()

	if len(userConfig.ExtensionToFolder) == 0 {
		userConfig.ExtensionToFolder = defaultConfig.ExtensionToFolder
	} else {
		merged := make(map[string]string)
		for ext, folder := range defaultConfig.ExtensionToFolder {
			merged[ext] = folder
		}
		for ext, folder := range userConfig.ExtensionToFolder {
			merged[ext] = folder
		}
		userConfig.ExtensionToFolder = merged
	}

	if userConfig.ArchiveExtractedFolder == "" {
		userConfig.ArchiveExtractedFolder = defaultConfig.ArchiveExtractedFolder
	}
	if userConfig.DuplicatesFolder == "" {
		userConfig.DuplicatesFolder = defaultConfig.DuplicatesFolder
	}
	if userConfig.TransparentPNGFolder == "" {
		userConfig.TransparentPNGFolder = defaultConfig.TransparentPNGFolder
	}

	return userConfig
}

func SaveExtensionConfig(config *ExtensionConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "GoSorter")
	configPath := filepath.Join(configDir, "extension.json")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
