// Package helpers - hashing
package helpers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/mohamedation/GoSorter/model"
)

// HashFile - size limit is for speed
func HashFile(filePath string, maxBytes int64, cfg model.Config, logger Logger) (string, error) {
	if logger != nil {
		logger.Log(cfg, Debug, fmt.Sprintf("[DEBUG] Opening file for hashing: %s", filePath))
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if logger != nil {
			logger.Log(cfg, Error, fmt.Sprintf("[ERROR] Stat failed for %s: %v", filePath, err))
		}
		return "", model.NewHashError(filePath, err)
	}
	if fileInfo.Size() > maxBytes {
		if logger != nil {
			logger.Log(cfg, Info, fmt.Sprintf("Skipping hash for %s: file size %d bytes exceeds max allowed %d bytes", filePath, fileInfo.Size(), maxBytes))
		}
		return "", nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		if logger != nil {
			logger.Log(cfg, Error, fmt.Sprintf("[ERROR] Failed to open file for hashing: %s, error: %v", filePath, err))
		}
		return "", model.NewHashError(filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil && logger != nil {
			logger.Log(cfg, Error, fmt.Sprintf("[ERROR] error closing file %s: %v", filePath, err))
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		if logger != nil {
			logger.Log(cfg, Error, fmt.Sprintf("[ERROR] Failed to hash file %s: %v", filePath, err))
		}
		return "", model.NewHashError(filePath, err)
	}

	result := fmt.Sprintf("%x", hash.Sum(nil))
	if result == "" && logger != nil {
		logger.Log(cfg, Debug, fmt.Sprintf("[DEBUG] Hash for file %s is empty, skipping", filePath))
	}

	if logger != nil {
		logger.Log(cfg, Debug, fmt.Sprintf("[DEBUG] Finished hashing file: %s", filePath))
	}

	return result, nil
}

// PartialHashFile - need to test if its accurate enough and if it actually saves time or just adds an extra step
func PartialHashFile(filePath string, numBytes int64, cfg model.Config, logger Logger) (string, error) {
	if logger != nil {
		logger.Log(cfg, Debug, fmt.Sprintf("[DEBUG] Starting partial hash for file: %s (first %d bytes)", filePath, numBytes))
	}
	file, err := os.Open(filePath)
	if err != nil {
		if logger != nil {
			logger.Log(cfg, Error, fmt.Sprintf("[ERROR] Failed to open file for partial hash: %s, error: %v", filePath, err))
		}
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil && logger != nil {
			logger.Log(cfg, Error, fmt.Sprintf("[ERROR] error closing file %s: %v", filePath, err))
		}
	}()

	buf := make([]byte, numBytes)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		if logger != nil {
			logger.Log(cfg, Error, fmt.Sprintf("[ERROR] Error reading file for partial hash: %s, error: %v", filePath, err))
		}
		return "", err
	}
	result := fmt.Sprintf("%x", sha256.Sum256(buf[:n]))
	if result == "" && logger != nil {
		logger.Log(cfg, Debug, fmt.Sprintf("[DEBUG] Partial hash for file %s is empty, skipping", filePath))
	}
	if logger != nil {
		logger.Log(cfg, Debug, fmt.Sprintf("[DEBUG] Finished partial hash for file: %s", filePath))
	}
	return result, nil
}
