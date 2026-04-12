package ulog

import (
	"fmt"
	"io"
	"sync"
)

// Публичные структуры
type TeeSink struct {
	mutex   sync.RWMutex
	writers []io.Writer
}
type Sink = io.Writer

// Публичные конструкторы
func NewTeeSink(writers ...io.Writer) *TeeSink {
	return &TeeSink{
		writers: writers,
	}
}

// Публичные методы
func (teeSink *TeeSink) Add(w io.Writer) {
	teeSink.mutex.Lock()
	defer teeSink.mutex.Unlock()
	teeSink.writers = append(teeSink.writers, w)
}
func (teeSink *TeeSink) Close() error {
	teeSink.mutex.Lock()
	defer teeSink.mutex.Unlock()
	var errors []error
	for i, w := range teeSink.writers {
		if closer, ok := w.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				errors = append(errors, fmt.Errorf("writer[%d]: %w", i, err))
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("close errors: %v", errors)
	}
	return nil
}
func (teeSink *TeeSink) Len() int {
	teeSink.mutex.RLock()
	defer teeSink.mutex.RUnlock()
	return len(teeSink.writers)
}
func (teeSink *TeeSink) Remove(index int) error {
	teeSink.mutex.Lock()
	defer teeSink.mutex.Unlock()
	if index < 0 || index >= len(teeSink.writers) {
		return fmt.Errorf("index out of range: %d", index)
	}
	teeSink.writers = append(teeSink.writers[:index], teeSink.writers[index+1:]...)
	return nil
}
func (teeSink *TeeSink) Replace(index int, w io.Writer) error {
	teeSink.mutex.Lock()
	defer teeSink.mutex.Unlock()
	if index < 0 || index >= len(teeSink.writers) {
		return fmt.Errorf("index out of range: %d", index)
	}
	if closer, ok := teeSink.writers[index].(io.Closer); ok {
		_ = closer.Close()
	}
	teeSink.writers[index] = w
	return nil
}
func (teeSink *TeeSink) Write(p []byte) (n int, err error) {
	teeSink.mutex.RLock()
	defer teeSink.mutex.RUnlock()
	if len(teeSink.writers) == 0 {
		return 0, nil
	}
	var errors []error
	for _, w := range teeSink.writers {
		if _, err := w.Write(p); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return len(p), fmt.Errorf("multi sink errors: %v", errors)
	}
	return len(p), nil
}

// Приватные константы
const (
	maxDiscordMessageLen  = 2000
	maxSlackMessageLen    = 4000
	maxTelegramMessageLen = 4096
)
