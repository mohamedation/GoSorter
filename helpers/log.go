// Package helpers - logger
package helpers

import (
	"fmt"
	"os"

	"github.com/mohamedation/GoSorter/model"
)

type LogLevel int

const (
	Normal LogLevel = iota
	Info
	Debug
	Error
	Raw // New log level for raw output
)

type Logger interface {
	Log(cfg model.Config, level LogLevel, message string)
}

type CLILogger struct{}

func (l *CLILogger) Log(cfg model.Config, level LogLevel, message string) {
	if level == Raw {
		fmt.Print(message)
		writeLogFile(cfg, message, l)
		return
	}
	if level == Error {
		if !cfg.Silent {
			fmt.Printf("\033[38;5;210m[ERROR] %s\033[0m", message)
		}
		writeLogFile(cfg, "[ERROR] "+message, l)
		return
	}

	if cfg.Silent {
		writeLogFile(cfg, message, l)
		return
	}

	if level == Debug && !cfg.Verbose {
		return
	}

	var colorCode, prefix string
	switch level {
	case Normal:
		colorCode = "\033[0m"
		prefix = ""
	case Info:
		colorCode = "\033[38;5;121m"
		prefix = "[INFO] "
	case Debug:
		colorCode = "\033[38;5;229m"
		prefix = "[DEBUG] "
	}

	logMessage := fmt.Sprintf("%s%s%s\033[0m", colorCode, prefix, message)
	fmt.Print(logMessage)
	writeLogFile(cfg, prefix+message, l)
}

func writeLogFile(cfg model.Config, message string, logger Logger) {
	if cfg.LogFilePath == "" {
		return
	}
	f, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		logger.Log(cfg, Error, "error opening log file: "+err.Error())
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Log(cfg, Error, "error closing log file: "+err.Error())
		}
	}()
	if _, err := f.WriteString(message + "\n"); err != nil {
		logger.Log(cfg, Error, "error writing to log file: "+err.Error())
	}
}
