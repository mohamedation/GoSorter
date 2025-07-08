// Package helpers - files
package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mohamedation/GoSorter/model"
)

func FolderExists(folderPath string) bool {
	_, err := os.Stat(folderPath)
	return !os.IsNotExist(err)
}

func MoveFile(src, dst string, cfg model.Config, logger Logger) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !os.IsExist(err) && !os.IsPermission(err) {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("error closing srcFile: %v", err))
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("error closing dstFile: %v", err))
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Remove(src)
}

func FormatPath(path string, cfg model.Config) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	if cfg.Verbose {
		if dir == "." {
			return base
		}
		return filepath.Join(dir, base)
	}

	if len(name) <= 32 {
		return base
	}
	return fmt.Sprintf("%s...%s%s", name[:3], name[len(name)-4:], ext)
}

func MoveDuplicateFile(folderPath, fileName, originalPath string, cfg model.Config, logger Logger) {
	duplicatesFolder := filepath.Join(folderPath, "Duplicates")
	if !FolderExists(duplicatesFolder) {
		if err := os.MkdirAll(duplicatesFolder, os.ModePerm); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to create folder %s: %v\n", duplicatesFolder, err))
			return
		}
	}

	originalFileName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
	duplicateFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	newDuplicateFileName := fmt.Sprintf("%s_duplicate_of_%s%s", duplicateFileName, originalFileName, filepath.Ext(fileName))
	duplicateDstPath := filepath.Join(duplicatesFolder, newDuplicateFileName)
	srcPath := filepath.Join(folderPath, fileName)

	if err := MoveFile(srcPath, duplicateDstPath, cfg, logger); err != nil {
		logger.Log(cfg, Error, fmt.Sprintf("Failed to move duplicate file %s: %v\n", srcPath, err))
		return
	}

	logger.Log(cfg, Info, fmt.Sprintf("Moved duplicate: %s -> %s\n", FormatPath(srcPath, cfg), FormatPath(duplicateDstPath, cfg)))
}

func MoveExtractedArchive(folderPath, fileName string, cfg model.Config, logger Logger) {
	extractedFolder := filepath.Join(folderPath, "Archives-Extracted")
	if !FolderExists(extractedFolder) {
		if err := os.MkdirAll(extractedFolder, os.ModePerm); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to create folder %s: %v\n", extractedFolder, err))
			return
		}
	}

	extractedDstPath := filepath.Join(extractedFolder, fileName)
	srcPath := filepath.Join(folderPath, fileName)
	if err := MoveFile(srcPath, extractedDstPath, cfg, logger); err != nil {
		logger.Log(cfg, Error, fmt.Sprintf("Failed to move archive %s: %v\n", srcPath, err))
		return
	}

	logger.Log(cfg, Info, fmt.Sprintf("Moved: %s -> %s\n", FormatPath(srcPath, cfg), FormatPath(extractedDstPath, cfg)))
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func MoveFileToTargetFolder(folderPath, fileName, targetFolder string, cfg model.Config, logger Logger) {
	targetPath := filepath.Join(folderPath, targetFolder)
	if !FolderExists(targetPath) {
		if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to create folder %s: %v\n", targetPath, err))
			return
		}
	}

	srcPath := filepath.Join(folderPath, fileName)
	dstPath := filepath.Join(targetPath, fileName)

	if FileExists(dstPath) {
		maxBytes := cfg.MaxHashFileSizeMB
		if maxBytes <= 0 {
			maxBytes = 1024
		}
		maxBytes = maxBytes * 1024 * 1024 // MB to bytes

		logger.Log(cfg, Debug, fmt.Sprintf("Hashing source file: %s\n", FormatPath(srcPath, cfg)))
		srcHash, err := HashFile(srcPath, maxBytes, cfg, logger)
		if err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to hash source file %s: %v\n", srcPath, err))
			return
		}
		logger.Log(cfg, Debug, fmt.Sprintf("Hashing destination file: %s\n", FormatPath(dstPath, cfg)))
		dstHash, err := HashFile(dstPath, maxBytes, cfg, logger)
		if err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to hash destination file %s: %v\n", dstPath, err))
			return
		}
		if srcHash != dstHash {
			ext := filepath.Ext(fileName)
			name := strings.TrimSuffix(fileName, ext)
			i := 1
			for {
				newName := fmt.Sprintf("%s(%d)%s", name, i, ext)
				newDstPath := filepath.Join(targetPath, newName)
				if !FileExists(newDstPath) {
					dstPath = newDstPath
					break
				}
				i++
			}
			logger.Log(cfg, Debug, fmt.Sprintf("File conflict: %s exists, renaming to %s\n", FormatPath(dstPath, cfg), FormatPath(dstPath, cfg)))
		} else {
			if err := os.Remove(dstPath); err != nil {
				logger.Log(cfg, Error, fmt.Sprintf("Failed to overwrite file %s: %v\n", dstPath, err))
				return
			}
			logger.Log(cfg, Debug, fmt.Sprintf("Overwriting file: %s\n", FormatPath(dstPath, cfg)))
		}
	}

	if err := MoveFile(srcPath, dstPath, cfg, logger); err != nil {
		logger.Log(cfg, Error, fmt.Sprintf("Failed to move file %s: %v\n", srcPath, err))
		return
	}
	logger.Log(cfg, Info, fmt.Sprintf("Moved: %s -> %s\n", FormatPath(srcPath, cfg), FormatPath(dstPath, cfg)))
}
