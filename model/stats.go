// Package model - stats
package model

import (
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	StartTime            time.Time
	EndTime              time.Time
	TimeElapsed          time.Duration
	FilesMoved           int64
	DuplicatesMoved      int64
	TotalFiles           int64
	ErrorsCount          int64
	TransparentPNGsMoved int64
	UnknownExtensions    int64
	UnknownExtMap        sync.Map
}

func (s *Stats) IncrementFilesMoved() {
	atomic.AddInt64(&s.FilesMoved, 1)
}

func (s *Stats) IncrementDuplicatesMoved() {
	atomic.AddInt64(&s.DuplicatesMoved, 1)
}

func (s *Stats) IncrementTotalFiles() {
	atomic.AddInt64(&s.TotalFiles, 1)
}

func (s *Stats) IncrementErrors() {
	atomic.AddInt64(&s.ErrorsCount, 1)
}

func (s *Stats) IncrementTransparentPNGsMoved() {
	atomic.AddInt64(&s.TransparentPNGsMoved, 1)
}

func (s *Stats) IncrementUnknownExtensions(ext string) {
	atomic.AddInt64(&s.UnknownExtensions, 1)
	// Track count for specific extension
	val, _ := s.UnknownExtMap.LoadOrStore(ext, new(int64))
	counter := val.(*int64)
	atomic.AddInt64(counter, 1)
}

func (s *Stats) GetFilesMoved() int64 {
	return atomic.LoadInt64(&s.FilesMoved)
}

func (s *Stats) GetDuplicatesMoved() int64 {
	return atomic.LoadInt64(&s.DuplicatesMoved)
}

func (s *Stats) GetTotalFiles() int64 {
	return atomic.LoadInt64(&s.TotalFiles)
}

func (s *Stats) GetErrorsCount() int64 {
	return atomic.LoadInt64(&s.ErrorsCount)
}

func (s *Stats) GetTransparentPNGsMoved() int64 {
	return atomic.LoadInt64(&s.TransparentPNGsMoved)
}

func (s *Stats) GetUnknownExtensions() int64 {
	return atomic.LoadInt64(&s.UnknownExtensions)
}

func (s *Stats) GetUnknownExtMap() map[string]int64 {
	result := make(map[string]int64)
	s.UnknownExtMap.Range(func(key, value interface{}) bool {
		ext := key.(string)
		counter := value.(*int64)
		count := atomic.LoadInt64(counter)
		result[ext] = count
		return true
	})
	return result
}
