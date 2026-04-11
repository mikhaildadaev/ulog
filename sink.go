// Этот файл (sink.go) находится в стадии активной разработки.
// API может изменяться до выхода версии v1.0.0.
//
// Планируется добавить:
// - Batch [пакетирование]
// - Circuit Breaker [защита от перегрузки]
// - Retry [повторные отправки]
package ulog

import (
	"fmt"
	"io"
	"sync"
)

// Публичные структуры
type MultiSink struct {
	mutex   sync.RWMutex
	writers []io.Writer
}
type Sink = io.Writer

// Публичные конструкторы
func NewMultiSink(writers ...io.Writer) *MultiSink {
	return &MultiSink{
		writers: writers,
	}
}

// Публичные методы
func (multiSink *MultiSink) Add(w io.Writer) {
	multiSink.mutex.Lock()
	defer multiSink.mutex.Unlock()
	multiSink.writers = append(multiSink.writers, w)
}
func (multiSink *MultiSink) Close() error {
	multiSink.mutex.Lock()
	defer multiSink.mutex.Unlock()
	var errors []error
	for i, w := range multiSink.writers {
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
func (multiSink *MultiSink) Len() int {
	multiSink.mutex.RLock()
	defer multiSink.mutex.RUnlock()
	return len(multiSink.writers)
}
func (multiSink *MultiSink) Remove(index int) error {
	multiSink.mutex.Lock()
	defer multiSink.mutex.Unlock()
	if index < 0 || index >= len(multiSink.writers) {
		return fmt.Errorf("index out of range: %d", index)
	}
	multiSink.writers = append(multiSink.writers[:index], multiSink.writers[index+1:]...)
	return nil
}
func (multiSink *MultiSink) Write(p []byte) (n int, err error) {
	multiSink.mutex.RLock()
	defer multiSink.mutex.RUnlock()
	if len(multiSink.writers) == 0 {
		return 0, nil
	}
	var errors []error
	for _, w := range multiSink.writers {
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
