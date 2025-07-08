package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mohamedation/GoSorter/helpers"
	"github.com/mohamedation/GoSorter/model"
	"github.com/mohamedation/GoSorter/service"
)

const (
	Version = "0.0.1"
	Author  = "mohamedation"
	Website = "https://mohamedation.com"
)

func main() {
	stats := &model.Stats{StartTime: time.Now()}
	var cfg model.Config

	logToFile := flag.Bool("l", false, "Enable logging to a file in the current directory")
	showHelp := flag.Bool("h", false, "Show help message")

	// arguments
	flag.BoolVar(&cfg.MoveDuplicates, "d", false, "Move duplicate files to the Duplicates folder")
	flag.BoolVar(&cfg.DuplicatesOnly, "do", false, "Only detect and move duplicates, no sorting")
	flag.BoolVar(&cfg.Verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&cfg.Silent, "s", false, "silent output")
	flag.BoolVar(&cfg.DetectTransparentPNGs, "t", false, "Check transparent PNGs (slower, but sorts PNGs with transparent backgrounds into PNGs folder)")

	// max hash file size (2048M or 2G)
	var maxHashSizeStr string
	flag.StringVar(&maxHashSizeStr, "S", "1024M", "Maximum file size to hash for duplicate detection (e.g. 1024M, 2G). Only used with -d or -do.")

	flag.Usage = func() {
		printHelp()
	}

	flag.Parse()

	if cfg.DuplicatesOnly {
		cfg.MoveDuplicates = true
	}

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	if *logToFile {
		timestamp := time.Now().Format("20060102-150405")
		cfg.LogFilePath = fmt.Sprintf("gosorter-%s.log", timestamp)
	}

	// check configuration
	if err := cfg.Validate(); err != nil {
		logger := &helpers.CLILogger{}
		logger.Log(cfg, helpers.Error, fmt.Sprintf("Configuration error: %v\n", err))
		os.Exit(1)
	}

	// max hash size
	cfg.MaxHashFileSizeMB = 1024 // default 1GB
	if maxHashSizeStr != "" {
		var sizeMB int64
		var unit string
		_, err := fmt.Sscanf(maxHashSizeStr, "%d%s", &sizeMB, &unit)
		if err == nil {
			switch unit {
			case "M", "m":
				cfg.MaxHashFileSizeMB = sizeMB
			case "G", "g":
				cfg.MaxHashFileSizeMB = sizeMB * 1024
			default:
				fmt.Fprintf(os.Stderr, "Unknown size unit for -S: %s (use M or G)\n", unit)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Invalid value for -S: %s (use e.g. 1024M or 2G)\n", maxHashSizeStr)
		}
	}

	// allow -S if -d or -do is set..maybe for the feature ignore it?
	if cfg.MaxHashFileSizeMB != 1024 && !cfg.MoveDuplicates && !cfg.DuplicatesOnly {
		fmt.Fprintf(os.Stderr, "-S is only used with -d or -do (duplicate detection modes)\n")
		cfg.MaxHashFileSizeMB = 1024
	}

	// directory path
	folderPath := "."
	if len(flag.Args()) > 0 {
		folderPath = flag.Arg(0)
	}

	processor := service.NewFileProcessor(&cfg, stats, &helpers.CLILogger{})

	ctx := context.Background()
	if err := processor.ProcessDirectory(ctx, folderPath); err != nil {
		fmt.Printf("Error processing directory: %v\n", err)
		os.Exit(1)
	}

	// statistics
	printStats(cfg, stats)
}

func printStats(cfg model.Config, stats *model.Stats) {
	// no stats
	if !cfg.Verbose && cfg.LogFilePath == "" {
		return
	}

	stats.EndTime = time.Now()
	stats.TimeElapsed = stats.EndTime.Sub(stats.StartTime)

	// needs some work
	statsContent := "\n====================[ GoSorter Stats ]====================\n"
	statsContent += fmt.Sprintf("%-25s %s\n", "Started at:", stats.StartTime.Format(time.RFC1123))
	statsContent += fmt.Sprintf("%-25s %s\n", "Finished at:", stats.EndTime.Format(time.RFC1123))
	statsContent += fmt.Sprintf("%-25s %s\n", "Duration:", stats.TimeElapsed)
	statsContent += "------------------------------------------------------------\n"
	statsContent += fmt.Sprintf("%-25s %d\n", "Total files processed:", stats.GetTotalFiles())
	statsContent += fmt.Sprintf("%-25s %d\n", "Files moved:", stats.GetFilesMoved())
	if cfg.MoveDuplicates {
		statsContent += fmt.Sprintf("%-25s %d\n", "Duplicate files moved:", stats.GetDuplicatesMoved())
	}
	if cfg.DetectTransparentPNGs {
		statsContent += fmt.Sprintf("%-25s %d\n", "Transparent PNGs moved:", stats.GetTransparentPNGsMoved())
	}
	if stats.GetUnknownExtensions() > 0 {
		statsContent += fmt.Sprintf("%-25s %d\n", "Unknown extensions:", stats.GetUnknownExtensions())
	}
	if stats.GetErrorsCount() > 0 {
		statsContent += fmt.Sprintf("%-25s %d\n", "Errors encountered:", stats.GetErrorsCount())
	}
	statsContent += "============================================================"

	// unknown extensions
	if stats.GetUnknownExtensions() > 0 {
		statsContent += "\n\n=================[ Unknown Extensions ]===================\n"
		unknownExtMap := stats.GetUnknownExtMap()
		for ext, count := range unknownExtMap {
			if ext == "" {
				ext = "(no extension)"
			}
			statsContent += fmt.Sprintf("%-25s %d files\n", ext+":", count)
		}
		statsContent += "===========================================================\n"
		statsContent += "Consider adding these extensions to your configuration file:\n"
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".config", "GoSorter", "extension.json")
		statsContent += configPath
	}

	// verbose
	if cfg.Verbose {
		fmt.Print(statsContent)
	}

	if cfg.LogFilePath != "" {
		logger := &helpers.CLILogger{}
		logger.Log(cfg, helpers.Info, statsContent)
	}
}

func printHelp() {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "GoSorter", "extension.json")
	progName := filepath.Base(os.Args[0])

	logger := helpers.CLILogger{}
	cfg := model.Config{} // config
	logger.Log(cfg, helpers.Normal, fmt.Sprintf("  GoSorter v%s\n", Version))
	logger.Log(cfg, helpers.Normal, fmt.Sprintf("  Created by: %s\n", Author))
	logger.Log(cfg, helpers.Normal, fmt.Sprintf("  Website: %s\n", Website))
	logger.Log(cfg, helpers.Normal, fmt.Sprintf("\nUSAGE:\n  %s [options] [directory]\n\n", progName))
	logger.Log(cfg, helpers.Normal, "A high-performance file organizer that sorts files into folders based on their extensions.\n")
	logger.Log(cfg, helpers.Normal, "\nOPTIONS:\n")
	options := []struct{ flag, desc string }{
		{"-h", "Show this help message"},
		{"-d", "Move duplicate files to the Duplicates folder"},
		{"-do", "Only detect and move duplicates, no extension-based sorting"},
		{"-v", "Enable verbose output with detailed statistics"},
		{"-s", "Enable silent output"},
		{"-l", "Enable logging to a file in the current directory"},
		{"-t", "Check transparent PNGs (slower, but sorts PNGs with transparent backgrounds)"},
	}
	for _, opt := range options {
		logger.Log(cfg, helpers.Normal, fmt.Sprintf("  %-4s %s\n", opt.flag, opt.desc))
	}
	logger.Log(cfg, helpers.Normal, "\nEXAMPLES:\n")
	examples := []struct{ cmd, desc string }{
		{progName, "Sort files in current directory"},
		{progName + " ~/Downloads", "Sort files in Downloads folder"},
		{progName + " -d -v ~/Documents", "Sort with duplicate detection and verbose output"},
		{progName + " -do ~/Documents", "Only detect and move duplicates, skip sorting"},
		{progName + " -t ~/Pictures", "Sort with transparent PNG detection"},
	}
	for _, ex := range examples {
		logger.Log(cfg, helpers.Normal, fmt.Sprintf("  %-30s # %s\n", ex.cmd, ex.desc))
	}
	logger.Log(cfg, helpers.Normal, fmt.Sprintf("\nCONFIGURATION:\n  Custom extension mappings can be configured in:\n  %s\n\n", configPath))
	logger.Log(cfg, helpers.Normal, "  Example config file:\n")
	configExample := `  {
    "extension_to_folder": {
      ".jpg": "MyPhotos",
      ".pdf": "Documents",
      ".custom": "CustomFolder"
    },
    "archives_extracted_folder": "Archives-Extracted",
    "duplicates_folder": "Duplicates",
    "transparent_png_folder": "PNGs"
  }`
	logger.Log(cfg, helpers.Normal, configExample+"\n")
}
