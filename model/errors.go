// Package model - errors (custom errors)
package model

import "fmt"

type FileError struct {
	Op   string
	Path string
	Err  error
}

func (e *FileError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

type HashError struct {
	*FileError
}

type MoveError struct {
	*FileError
	Dest string
}

func (e *MoveError) Error() string {
	return fmt.Sprintf("move %s to %s: %v", e.Path, e.Dest, e.Err)
}

func NewHashError(path string, err error) *HashError {
	return &HashError{
		FileError: &FileError{
			Op:   "hash",
			Path: path,
			Err:  err,
		},
	}
}

func NewMoveError(src, dest string, err error) *MoveError {
	return &MoveError{
		FileError: &FileError{
			Op:   "move",
			Path: src,
			Err:  err,
		},
		Dest: dest,
	}
}
