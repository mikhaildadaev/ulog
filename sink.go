// ULOG (Universal Logger Observability Gateway)
// Copyright (C) 2026 Mikhail Dadaev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package ulog

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Публичные структуры
type TeeSink struct {
	closeErr  error
	closeOnce sync.Once
	mutex     sync.Mutex
	pointer   atomic.Pointer[[]Sink]
	writers   []io.Writer
}
type Sink = io.Writer

// Публичные конструкторы
func NewTeeSink(writers ...Sink) *TeeSink {
	teeSink := &TeeSink{}
	teeSink.pointer.Store(&writers)
	return teeSink
}

// Публичные методы
func (teeSink *TeeSink) Add(sink Sink) {
	teeSink.mutex.Lock()
	defer teeSink.mutex.Unlock()
	current := teeSink.snapshot()
	next := make([]Sink, len(current)+1)
	copy(next, current)
	next[len(current)] = sink
	teeSink.pointer.Store(&next)
}
func (teeSink *TeeSink) Close() error {
	teeSink.closeOnce.Do(func() {
		writers := teeSink.snapshot()
		var errors []error
		for i, w := range writers {
			if closer, ok := w.(io.Closer); ok {
				if err := closer.Close(); err != nil {
					errors = append(errors, fmt.Errorf("tee[%d]: %w", i, err))
				}
			}
		}
		if len(errors) > 0 {
			teeSink.closeErr = fmt.Errorf("close errors: %v", errors)
		}
	})
	return teeSink.closeErr
}
func (teeSink *TeeSink) Flush() error {
	writers := teeSink.snapshot()
	var errors []error
	for i, w := range writers {
		if flusher, ok := w.(interface{ Flush() error }); ok {
			if err := flusher.Flush(); err != nil {
				errors = append(errors, fmt.Errorf("tee[%d]: %w", i, err))
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("flush errors: %v", errors)
	}
	return nil
}
func (teeSink *TeeSink) Len() int {
	return len(teeSink.snapshot())
}
func (teeSink *TeeSink) Remove(index int) error {
	teeSink.mutex.Lock()
	defer teeSink.mutex.Unlock()
	current := teeSink.snapshot()
	if index < 0 || index >= len(current) {
		return fmt.Errorf("index out of range: %d", index)
	}
	next := make([]Sink, len(current)-1)
	copy(next, current[:index])
	copy(next[index:], current[index+1:])
	teeSink.pointer.Store(&next)
	return nil
}
func (teeSink *TeeSink) Replace(index int, sink Sink) error {
	teeSink.mutex.Lock()
	defer teeSink.mutex.Unlock()
	current := teeSink.snapshot()
	if index < 0 || index >= len(current) {
		return fmt.Errorf("index out of range: %d", index)
	}
	next := make([]Sink, len(current))
	copy(next, current)
	next[index] = sink
	teeSink.pointer.Store(&next)
	return nil
}
func (teeSink *TeeSink) Write(p []byte) (n int, err error) {
	writers := teeSink.snapshot()
	if len(writers) == 0 {
		return 0, nil
	}
	var errors []error
	for i, w := range writers {
		if _, err := w.Write(p); err != nil {
			errors = append(errors, fmt.Errorf("tee[%d]: %w", i, err))
		}
	}
	if len(errors) > 0 {
		return len(p), fmt.Errorf("write errors: %v", errors)
	}
	return len(p), nil
}
func (teeSink *TeeSink) WriteWithAttributes(attributes writeAttributes, fields []Field) (n int, err error) {
	writers := teeSink.snapshot()
	if len(writers) == 0 {
		return 0, nil
	}
	var formatted []byte
	var buf *bytes.Buffer
	for _, writer := range writers {
		if _, ok := writer.(SinkWriter); !ok {
			buf = dataPool.Get().(*bytes.Buffer)
			buf.Reset()
			switch attributes.typeFormat {
			case FormatText:
				formatText(buf, attributes, fields)
			default:
				formatJson(buf, attributes, fields)
			}
			formatted = buf.Bytes()
			break
		}
	}
	defer func() {
		if buf != nil {
			dataPool.Put(buf)
		}
	}()
	var errors []error
	total := 0
	for i, writer := range writers {
		if sink, ok := writer.(SinkWriter); ok {
			w, err := sink.WriteWithAttributes(attributes, fields)
			total += w
			if err != nil {
				errors = append(errors, fmt.Errorf("tee[%d]: %w", i, err))
			}
			continue
		}
		if formatted == nil {
			continue
		}
		w, err := writer.Write(formatted)
		total += w
		if err != nil {
			errors = append(errors, fmt.Errorf("tee[%d]: %w", i, err))
		}
	}
	if len(errors) > 0 {
		return total, fmt.Errorf("write errors: %v", errors)
	}
	return total, nil
}

// Приватные методы
func (teeSink *TeeSink) snapshot() []Sink {
	if snap := teeSink.pointer.Load(); snap != nil {
		return *snap
	}
	return nil
}
