// Package helpers - files
package helpers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mohamedation/GoSorter/model"
)

func FolderExists(folderPath string) bool {
	_, err := os.Stat(folderPath)
	return !os.IsNotExist(err)
}

// MoveFile moves src to dst. Tries rename first; on cross-device (or other
// rename failure that needs a copy), copies via a temp file in dst's directory
// so a failed copy never truncates an existing destination.
func MoveFile(src, dst string, cfg model.Config, logger Logger) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !shouldCopyFallback(err) {
		return err
	} else {
		logger.Log(cfg, Debug, fmt.Sprintf("Rename failed (%v), falling back to copy: %s -> %s\n", err, FormatPath(src, cfg), FormatPath(dst, cfg)))
	}

	return copyFileThenRemove(src, dst, cfg, logger)
}

// shouldCopyFallback reports whether a rename error should be handled by
// copy-then-remove (cross-device, Windows not-same-volume, exist/permission quirks).
func shouldCopyFallback(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EXDEV) {
		return true
	}
	// Windows ERROR_NOT_SAME_DEVICE
	var errno syscall.Errno
	if errors.As(err, &errno) && errno == 17 {
		return true
	}
	var linkErr *os.LinkError
	if errors.As(err, &linkErr) {
		return shouldCopyFallback(linkErr.Err)
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return shouldCopyFallback(pathErr.Err)
	}
	// keep previous fallback triggers (e.g. Windows dest-exists rename behavior)
	if os.IsExist(err) || os.IsPermission(err) {
		return true
	}
	return false
}

// copyFileThenRemove copies src to dst via a temp file, syncs, preserves mode
// and mtime, then removes src. Final dst is only replaced after a durable copy.
func copyFileThenRemove(src, dst string, cfg model.Config, logger Logger) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
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

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".gosorter-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	closed := false
	renamed := false
	defer func() {
		if !closed {
			_ = tmp.Close()
		}
		if !renamed {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(tmp, srcFile); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	closed = true

	// preserve permissions from source before publishing as dst
	if err := os.Chmod(tmpName, srcInfo.Mode()); err != nil {
		return err
	}

	if err := os.Rename(tmpName, dst); err != nil {
		return err
	}
	renamed = true

	// mtime is best-effort; xattrs are not preserved
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		logger.Log(cfg, Debug, fmt.Sprintf("could not preserve mtime for %s: %v\n", FormatPath(dst, cfg), err))
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

func MoveDuplicateFile(folderPath, fileName, originalPath string, cfg model.Config, logger Logger) error {
	duplicatesFolder := filepath.Join(folderPath, "Duplicates")
	if !FolderExists(duplicatesFolder) {
		if err := os.MkdirAll(duplicatesFolder, 0750); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to create folder %s: %v\n", duplicatesFolder, err))
			return err
		}
	}

	originalFileName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
	duplicateFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	newDuplicateFileName := fmt.Sprintf("%s_duplicate_of_%s%s", duplicateFileName, originalFileName, filepath.Ext(fileName))
	duplicateDstPath := filepath.Join(duplicatesFolder, newDuplicateFileName)
	srcPath := filepath.Join(folderPath, fileName)

	if err := MoveFile(srcPath, duplicateDstPath, cfg, logger); err != nil {
		logger.Log(cfg, Error, fmt.Sprintf("Failed to move duplicate file %s: %v\n", srcPath, err))
		return err
	}

	logger.Log(cfg, Info, fmt.Sprintf("Moved duplicate: %s -> %s\n", FormatPath(srcPath, cfg), FormatPath(duplicateDstPath, cfg)))
	return nil
}

func MoveExtractedArchive(folderPath, fileName string, cfg model.Config, logger Logger) error {
	extractedFolder := filepath.Join(folderPath, "Archives-Extracted")
	if !FolderExists(extractedFolder) {
		if err := os.MkdirAll(extractedFolder, 0750); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to create folder %s: %v\n", extractedFolder, err))
			return err
		}
	}

	extractedDstPath := filepath.Join(extractedFolder, fileName)
	srcPath := filepath.Join(folderPath, fileName)
	if err := MoveFile(srcPath, extractedDstPath, cfg, logger); err != nil {
		logger.Log(cfg, Error, fmt.Sprintf("Failed to move archive %s: %v\n", srcPath, err))
		return err
	}

	logger.Log(cfg, Info, fmt.Sprintf("Moved: %s -> %s\n", FormatPath(srcPath, cfg), FormatPath(extractedDstPath, cfg)))
	return nil
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func MoveFileToTargetFolder(folderPath, fileName, targetFolder string, cfg model.Config, logger Logger) error {
	targetPath := filepath.Join(folderPath, targetFolder)
	if !FolderExists(targetPath) {
		if err := os.MkdirAll(targetPath, 0750); err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to create folder %s: %v\n", targetPath, err))
			return err
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
			return err
		}
		logger.Log(cfg, Debug, fmt.Sprintf("Hashing destination file: %s\n", FormatPath(dstPath, cfg)))
		dstHash, err := HashFile(dstPath, maxBytes, cfg, logger)
		if err != nil {
			logger.Log(cfg, Error, fmt.Sprintf("Failed to hash destination file %s: %v\n", dstPath, err))
			return err
		}
		if srcHash != "" && srcHash == dstHash {
			if err := os.Remove(srcPath); err != nil {
				logger.Log(cfg, Error, fmt.Sprintf("Failed to remove duplicate file %s: %v\n", srcPath, err))
				return err
			}
			logger.Log(cfg, Info, fmt.Sprintf("Removed duplicate: %s (identical to %s)\n", FormatPath(srcPath, cfg), FormatPath(dstPath, cfg)))
			return nil
		}
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
	}

	if err := MoveFile(srcPath, dstPath, cfg, logger); err != nil {
		logger.Log(cfg, Error, fmt.Sprintf("Failed to move file %s: %v\n", srcPath, err))
		return err
	}
	logger.Log(cfg, Info, fmt.Sprintf("Moved: %s -> %s\n", FormatPath(srcPath, cfg), FormatPath(dstPath, cfg)))
	return nil
}
