// Package model - configuration
package model

import "fmt"

type Config struct {
	MoveDuplicates        bool
	DuplicatesOnly        bool
	Verbose               bool
	Silent                bool
	LogFilePath           string
	DetectTransparentPNGs bool
	MaxHashFileSizeMB     int64
	MaxHashFileSize       int64
}

const (
	DefaultWorkerCount = 8
	DefaultBufferSize  = 100
)

func (c *Config) Validate() error {
	if c.Verbose && c.Silent {
		return fmt.Errorf("verbose and silent modes cannot be enabled simultaneously, otherwise, GoSorter might take a selfie")
	}
	if c.MaxHashFileSizeMB < 0 {
		return fmt.Errorf("max hash file size must be >= 0")
	}
	return nil
}
