// Package service - file processor service
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mohamedation/GoSorter/helpers"
	"github.com/mohamedation/GoSorter/model"
)

type FileProcessor struct {
	config    *model.Config
	stats     *model.Stats
	extConfig *model.ExtensionConfig
	Logger    helpers.Logger
}

// move a file to its appropriate target folder
func (fp *FileProcessor) moveFileToTargetFolder(folderPath string, file model.FileDetail) error {
	return fp.moveFileToFolder(folderPath, file, fp.extConfig)
}

// NewFileProcessor -  file processor instance
func NewFileProcessor(config *model.Config, stats *model.Stats, logger helpers.Logger) *FileProcessor {
	return &FileProcessor{
		config:    config,
		stats:     stats,
		extConfig: model.LoadExtensionConfig(),
		Logger:    logger,
	}
}

// ProcessDirectory - processes all files in the directory
func (fp *FileProcessor) ProcessDirectory(ctx context.Context, folderPath string) error {
	if !helpers.FolderExists(folderPath) {
		return fmt.Errorf("directory '%s' does not exist", folderPath)
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	return fp.processFiles(ctx, folderPath, entries)
}

// file processing logic
func (fp *FileProcessor) processFiles(ctx context.Context, folderPath string, entries []os.DirEntry) error {
	if fp.config.MoveDuplicates {
		return fp.processFilesWithDuplicates(ctx, folderPath, entries)
	}
	return fp.processFilesWithoutDuplicates(ctx, folderPath, entries)
}

// duplicate detection enabled
func (fp *FileProcessor) processFilesWithDuplicates(ctx context.Context, folderPath string, entries []os.DirEntry) error {
	// group by size to check duplicates
	fp.Logger.Log(*fp.config, helpers.Debug, "[DEBUG] Grouping files by size\n")
	sizeGroups := make(map[int64][]os.DirEntry)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(folderPath, entry.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			fp.stats.IncrementErrors()
			fp.Logger.Log(*fp.config, helpers.Error, fmt.Sprintf("Failed to stat file %s: %v\n", filePath, err))
			continue
		}
		sizeGroups[info.Size()] = append(sizeGroups[info.Size()], entry)
	}

	fileHashes := make(map[string][]model.FileDetail)
	fileDetails := []model.FileDetail{}

	type hashResult struct {
		detail model.FileDetail
		err    error
	}

	partialHashSize := int64(4096) // 4KB only partial hash (still reconsidering if needed or just an extra step)
	maxBytes := fp.config.MaxHashFileSizeMB
	if maxBytes <= 0 {
		maxBytes = 1024 // i made this redundant during refactoring?
	}
	maxBytes = maxBytes * 1024 * 1024

	for size, group := range sizeGroups {
		if len(group) == 1 {
			entry := group[0]
			filePath := filepath.Join(folderPath, entry.Name())
			uniqueHash := "size-" + fmt.Sprint(size)
			detail := model.FileDetail{
				Name: entry.Name(),
				Path: filePath,
				Hash: uniqueHash,
				Ext:  strings.ToLower(filepath.Ext(entry.Name())),
			}
			fileDetails = append(fileDetails, detail)
			fileHashes[uniqueHash] = append(fileHashes[uniqueHash], detail)
			fp.stats.IncrementTotalFiles()
			continue
		}

		fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("[DEBUG] Processing %d files of size %d bytes\n", len(group), size))

		// part. hash (still reconsidering if needed or just an extra step)
		partialHashes := make(map[string][]os.DirEntry)
		for _, entry := range group {
			filePath := filepath.Join(folderPath, entry.Name())
			partialHash, err := helpers.PartialHashFile(filePath, partialHashSize, *fp.config, fp.Logger)
			if err != nil {
				fp.stats.IncrementErrors()
				fp.Logger.Log(*fp.config, helpers.Error, fmt.Sprintf("Failed to partial hash file %s: %v\n", filePath, err))
				continue
			}
			partialHashes[partialHash] = append(partialHashes[partialHash], entry)
		}

		// full hash as last resort
		for pHash, pGroup := range partialHashes {
			if len(pGroup) == 1 {
				entry := pGroup[0]
				filePath := filepath.Join(folderPath, entry.Name())
				detail := model.FileDetail{
					Name: entry.Name(),
					Path: filePath,
					Ext:  strings.ToLower(filepath.Ext(entry.Name())),
				}
				fileDetails = append(fileDetails, detail)
				fileHashes["size-partial-"+fmt.Sprint(size)+"-"+pHash] = append(fileHashes["size-partial-"+fmt.Sprint(size)+"-"+pHash], detail)
				fp.stats.IncrementTotalFiles()
				continue
			}

			fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("[DEBUG] Processing %d files with size %d bytes and partial hash %s\n", len(pGroup), size, pHash))

			numWorkers := runtime.NumCPU()
			jobs := make(chan os.DirEntry, len(pGroup))
			results := make(chan hashResult, len(pGroup))
			var wg sync.WaitGroup

			worker := func() {
				defer wg.Done()
				for entry := range jobs {
					filePath := filepath.Join(folderPath, entry.Name())
					hash, err := helpers.HashFile(filePath, maxBytes, *fp.config, fp.Logger)
					detail := model.FileDetail{
						Name: entry.Name(),
						Path: filePath,
						Hash: hash,
						Ext:  strings.ToLower(filepath.Ext(entry.Name())),
					}
					results <- hashResult{detail: detail, err: err}
				}
			}

			for range numWorkers {
				wg.Add(1)
				go worker()
			}

			for _, entry := range pGroup {
				select {
				case jobs <- entry:
				case <-ctx.Done():
					close(jobs)
					return ctx.Err()
				}
			}
			close(jobs)

			go func() {
				wg.Wait()
				close(results)
			}()

			for res := range results {
				if res.err != nil {
					fp.stats.IncrementErrors()
					fp.Logger.Log(*fp.config, helpers.Error, fmt.Sprintf("Failed to hash file %s: %v\n", res.detail.Path, res.err))
					continue
				}
				fileDetails = append(fileDetails, res.detail)
				fileHashes[res.detail.Hash] = append(fileHashes[res.detail.Hash], res.detail)
				fp.stats.IncrementTotalFiles()
			}
		}
	}

	// process
	return fp.processFileGroups(folderPath, fileDetails, fileHashes)
}

// processes files normally
func (fp *FileProcessor) processFilesWithoutDuplicates(ctx context.Context, folderPath string, entries []os.DirEntry) error {
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(folderPath, entry.Name())
		detail := model.FileDetail{
			Name: entry.Name(),
			Path: filePath,
			Ext:  strings.ToLower(filepath.Ext(entry.Name())),
		}

		fp.stats.IncrementTotalFiles()

		// unknown extension
		if _, ok := fp.extConfig.ExtensionToFolder[detail.Ext]; !ok {
			fp.stats.IncrementUnknownExtensions(detail.Ext)
			fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("Skipping file with unknown extension: %s (%s)\n", detail.Name, detail.Ext))
			continue
		}

		if err := fp.moveFileToTargetFolder(folderPath, detail); err != nil {
			fp.stats.IncrementErrors()
			fp.Logger.Log(*fp.config, helpers.Error, fmt.Sprintf("Failed to move file: %v\n", err))
			continue
		}
		fp.stats.IncrementFilesMoved()
	}
	return nil
}

// handles file groups
func (fp *FileProcessor) processFileGroups(folderPath string, fileDetails []model.FileDetail, fileHashes map[string][]model.FileDetail) error {
	processedHashes := make(map[string]bool)

	for _, detail := range fileDetails {
		if detail.Hash == "" {
			fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("Skipping file %s due to empty hash\n", detail.Path))
			continue
		}
		if processedHashes[detail.Hash] {
			continue
		}

		files := fileHashes[detail.Hash]
		if len(files) == 0 {
			fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("No files found for hash %s, skipping\n", detail.Hash))
			continue
		}

		if len(files) == 1 {
			if fp.config.DuplicatesOnly {
				continue
			}
			original := files[0]
			if _, ok := fp.extConfig.ExtensionToFolder[original.Ext]; !ok {
				fp.stats.IncrementUnknownExtensions(original.Ext)
				fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("Skipping file with unknown extension: %s (%s)\n", original.Name, original.Ext))
				continue
			}
			if err := fp.moveFileToFolder(folderPath, original, fp.extConfig); err != nil {
				fp.stats.IncrementErrors()
				fp.Logger.Log(*fp.config, helpers.Error, fmt.Sprintf("Failed to move file: %v\n", err))
			} else {
				fp.stats.IncrementFilesMoved()
			}
			processedHashes[detail.Hash] = true
			continue
		}

		original := files[0]
		for _, file := range files {
			if len(file.Name) < len(original.Name) {
				original = file
			}
		}

		for _, file := range files {
			if file.Path == original.Path {
				continue
			}
			helpers.MoveDuplicateFile(folderPath, file.Name, original.Path, *fp.config, fp.Logger)
			fp.stats.IncrementDuplicatesMoved()
		}

		if fp.config.DuplicatesOnly {
			fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("Duplicates-only mode: leaving original file %s in place\n", original.Name))
		} else {
			// Normal mode: move original file to appropriate folder
			if _, ok := fp.extConfig.ExtensionToFolder[original.Ext]; !ok {
				fp.stats.IncrementUnknownExtensions(original.Ext)
				fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("Skipping file with unknown extension: %s (%s)\n", original.Name, original.Ext))
				processedHashes[detail.Hash] = true
				continue
			}
			if err := fp.moveFileToFolder(folderPath, original, fp.extConfig); err != nil {
				fp.stats.IncrementErrors()
				fp.Logger.Log(*fp.config, helpers.Error, fmt.Sprintf("Failed to move file: %v\n", err))
			} else {
				fp.stats.IncrementFilesMoved()
			}
		}
		processedHashes[detail.Hash] = true
	}
	return nil
}

// move files to their appropriate folder
func (fp *FileProcessor) moveFileToFolder(folderPath string, file model.FileDetail, config *model.ExtensionConfig) error {
	targetFolder, ok := config.ExtensionToFolder[file.Ext]
	if !ok {
		fp.stats.IncrementUnknownExtensions(file.Ext)
		fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("Skipping file with unknown extension: %s (%s)\n", file.Name, file.Ext))
		return fmt.Errorf("unknown extension: %s", file.Ext)
	}

	switch file.Ext {
	case ".zip":
		dirName := strings.TrimSuffix(file.Name, file.Ext)
		dirPath := filepath.Join(folderPath, dirName)
		if helpers.FolderExists(dirPath) {
			helpers.MoveExtractedArchive(folderPath, file.Name, *fp.config, fp.Logger)
			return nil
		}
		helpers.MoveFileToTargetFolder(folderPath, file.Name, "Archives", *fp.config, fp.Logger)
	case ".png":
		if fp.config.DetectTransparentPNGs {
			hasTransparency, err := helpers.HasTransparency(file.Path, *fp.config, fp.Logger)
			if err != nil {
				fp.stats.IncrementErrors()
				fp.Logger.Log(*fp.config, helpers.Error, fmt.Sprintf("Failed to check transparency for file %s: %v\n", file.Path, err))
				hasTransparency = false
			}
			if hasTransparency {
				fp.stats.IncrementTransparentPNGsMoved()
				fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("PNG file %s has transparency, moving to %s\n", file.Name, config.TransparentPNGFolder))
				helpers.MoveFileToTargetFolder(folderPath, file.Name, config.TransparentPNGFolder, *fp.config, fp.Logger)
			} else {
				fp.Logger.Log(*fp.config, helpers.Debug, fmt.Sprintf("PNG file %s has no transparency, moving to %s\n", file.Name, targetFolder))
				helpers.MoveFileToTargetFolder(folderPath, file.Name, targetFolder, *fp.config, fp.Logger)
			}
		} else {
			helpers.MoveFileToTargetFolder(folderPath, file.Name, targetFolder, *fp.config, fp.Logger)
		}
	default:
		helpers.MoveFileToTargetFolder(folderPath, file.Name, targetFolder, *fp.config, fp.Logger)
	}
	return nil
}
