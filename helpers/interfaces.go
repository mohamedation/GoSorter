// Package helpers - interfaces
package helpers

import (
	"os"

	"github.com/mohamedation/GoSorter/model"
)

type FileOperations interface {
	Exists(path string) bool
	FolderExists(path string) bool
	MoveFile(src, dst string, cfg model.Config, logger Logger) error
	HashFile(path string, cfg model.Config, logger Logger) (string, error)
	CreateFolder(path string) error
}

type OSFileOperations struct{}

func NewOSFileOperations() *OSFileOperations {
	return &OSFileOperations{}
}

func (o *OSFileOperations) Exists(path string) bool {
	return FileExists(path)
}

func (o *OSFileOperations) FolderExists(path string) bool {
	return FolderExists(path)
}

func (o *OSFileOperations) MoveFile(src, dst string, cfg model.Config, logger Logger) error {
	return MoveFile(src, dst, cfg, logger)
}

func (o *OSFileOperations) HashFile(path string, cfg model.Config, logger Logger) (string, error) {
	return HashFile(path, 1024*1024*1024, cfg, logger)
}

func (o *OSFileOperations) CreateFolder(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

type FileMover struct {
	fileOps FileOperations
	config  *model.Config
}

func NewFileMover(fileOps FileOperations, config *model.Config) *FileMover {
	return &FileMover{
		fileOps: fileOps,
		config:  config,
	}
}

func (fm *FileMover) MoveToTargetFolder(folderPath, fileName, targetFolder string) error {
	MoveFileToTargetFolder(folderPath, fileName, targetFolder, *fm.config, &CLILogger{})
	return nil
}

func (fm *FileMover) MoveToDuplicates(folderPath, fileName, originalPath string) error {
	MoveDuplicateFile(folderPath, fileName, originalPath, *fm.config, &CLILogger{})
	return nil
}

func (fm *FileMover) MoveExtractedArchive(folderPath, fileName string) error {
	MoveExtractedArchive(folderPath, fileName, *fm.config, &CLILogger{})
	return nil
}
