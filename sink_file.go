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
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Публичные структуры
type SinkFile struct {
	bufWriter     *bufio.Writer
	currentSize   int64
	file          *os.File
	filename      string
	flushDone     chan struct{}
	flushInterval time.Duration
	flushTicker   *time.Ticker
	maxAge        int
	maxBackups    int
	maxSize       int64
	mutex         sync.Mutex
	once          sync.Once
	rotating      atomic.Bool
	wg            sync.WaitGroup
}

// Публичные конструкторы
func NewSinkFile(filename string, params ...fileParams) (*SinkFile, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	sinkFile := &SinkFile{
		bufWriter:     bufio.NewWriterSize(file, 64*1024),
		currentSize:   info.Size(),
		file:          file,
		filename:      filename,
		flushInterval: 1 * time.Second,
		maxAge:        30,
		maxBackups:    10,
		maxSize:       100 * 1024 * 1024,
	}
	for _, param := range params {
		param(sinkFile)
	}
	if sinkFile.flushInterval > 0 {
		sinkFile.flushDone = make(chan struct{})
		sinkFile.flushTicker = time.NewTicker(sinkFile.flushInterval)
		sinkFile.wg.Add(1)
		go func() {
			defer sinkFile.wg.Done()
			sinkFile.flushLoop()
		}()
	}
	return sinkFile, nil
}

// Публичные функции
func WithFileFlushInterval(interval time.Duration) fileParams {
	return func(sinkFile *SinkFile) {
		sinkFile.flushInterval = interval
	}
}
func WithFileMaxAge(days int) fileParams {
	return func(sinkFile *SinkFile) {
		sinkFile.maxAge = days
	}
}
func WithFileMaxBackups(count int) fileParams {
	return func(sinkFile *SinkFile) {
		sinkFile.maxBackups = count
	}
}
func WithFileMaxSize(sizeMB int) fileParams {
	return func(sinkFile *SinkFile) {
		sinkFile.maxSize = int64(sizeMB) * 1024 * 1024
	}
}

// Публичные методы
func (sinkFile *SinkFile) Close() error {
	var err error
	sinkFile.once.Do(func() {
		sinkFile.mutex.Lock()
		if sinkFile.flushTicker != nil {
			sinkFile.flushTicker.Stop()
		}
		if sinkFile.flushDone != nil {
			close(sinkFile.flushDone)
		}
		sinkFile.mutex.Unlock()
		done := make(chan struct{})
		go func() {
			sinkFile.wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(maxWaitFile):
			fmt.Fprintf(DefaultWriterErr, "ulog: SinkFile.Close() timeout waiting for background tasks\n")
		}
		sinkFile.mutex.Lock()
		defer sinkFile.mutex.Unlock()
		if sinkFile.bufWriter != nil {
			if flushErr := sinkFile.bufWriter.Flush(); flushErr != nil {
				err = flushErr
			}
			sinkFile.bufWriter = nil
		}
		if sinkFile.file != nil {
			if closeErr := sinkFile.file.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			sinkFile.file = nil
		}
	})
	return err
}
func (sinkFile *SinkFile) Flush() error {
	sinkFile.mutex.Lock()
	defer sinkFile.mutex.Unlock()
	if sinkFile.bufWriter == nil {
		return nil
	}
	return sinkFile.bufWriter.Flush()
}
func (sinkFile *SinkFile) Write(p []byte) (int, error) {
	sinkFile.mutex.Lock()
	needRotate := sinkFile.currentSize+int64(len(p)) > sinkFile.maxSize
	sinkFile.mutex.Unlock()
	if needRotate {
		if err := sinkFile.rotate(); err != nil {
			return 0, err
		}
	}
	for {
		sinkFile.mutex.Lock()
		if sinkFile.file != nil {
			break
		}
		sinkFile.mutex.Unlock()
		time.Sleep(time.Millisecond)
	}
	defer sinkFile.mutex.Unlock()
	if sinkFile.bufWriter == nil {
		return 0, fmt.Errorf("sink is closed")
	}
	n, err := sinkFile.bufWriter.Write(p)
	if err == nil {
		sinkFile.currentSize += int64(n)
	}
	return n, err
}
func (sinkFile *SinkFile) WriteWithAttributes(attributes writeAttributes, fields []Field) (int, error) {
	bufData := dataPool.Get().(*bytes.Buffer)
	bufData.Reset()
	defer dataPool.Put(bufData)
	switch attributes.typeFormat {
	case FormatJson:
		formatJson(bufData, attributes, fields)
	case FormatText:
		formatText(bufData, attributes, fields)
	default:
		return 0, fmt.Errorf("unsupported format: %v", attributes.typeFormat)
	}
	data := bufData.Bytes()
	sinkFile.mutex.Lock()
	needRotate := sinkFile.currentSize+int64(len(data)) > sinkFile.maxSize
	sinkFile.mutex.Unlock()
	if needRotate {
		if err := sinkFile.rotate(); err != nil {
			return 0, err
		}
	}
	for {
		sinkFile.mutex.Lock()
		if sinkFile.file != nil {
			break
		}
		sinkFile.mutex.Unlock()
		time.Sleep(time.Millisecond)
	}
	defer sinkFile.mutex.Unlock()
	if sinkFile.bufWriter == nil {
		return 0, fmt.Errorf("sink is closed")
	}
	n, err := sinkFile.bufWriter.Write(data)
	if err == nil {
		sinkFile.currentSize += int64(n)
	}
	return n, err
}

// Приватные структуры
type fileParams func(*SinkFile)

// Приватные функции
func (sinkFile *SinkFile) cleanup() error {
	pattern := sinkFile.getBackupPattern() + ".gz"
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})
	if sinkFile.maxBackups > 0 && len(files) > sinkFile.maxBackups {
		for _, file := range files[sinkFile.maxBackups:] {
			if err := os.Remove(file); err != nil {
				fmt.Fprintf(DefaultWriterErr, "failed to remove old backup %s: %v\n", file, err)
			}
		}
		files = files[:sinkFile.maxBackups]
	}
	if sinkFile.maxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -sinkFile.maxAge)
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				if err := os.Remove(file); err != nil {
					fmt.Fprintf(DefaultWriterErr, "failed to remove old backup %s: %v\n", file, err)
				}
			}
		}
	}
	return nil
}
func (fileSink *SinkFile) compress(filename string) error {
	src, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer src.Close()
	tmpName := filename + ".gz.tmp"
	dst, err := os.Create(tmpName)
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(dst)
	if _, err = io.Copy(gz, src); err != nil {
		gz.Close()
		dst.Close()
		os.Remove(tmpName)
		return err
	}
	if err = gz.Close(); err != nil {
		dst.Close()
		os.Remove(tmpName)
		return err
	}
	if err = dst.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	gzName := filename + ".gz"
	if _, statErr := os.Stat(gzName); statErr == nil {
		if err = os.Remove(gzName); err != nil {
			return err
		}
	}
	if err = os.Rename(tmpName, gzName); err != nil {
		return err
	}
	err = os.Remove(filename)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
func (sinkFile *SinkFile) flushLoop() {
	ticker := sinkFile.flushTicker
	for {
		select {
		case <-ticker.C:
			sinkFile.Flush()
		case <-sinkFile.flushDone:
			return
		}
	}
}
func (sinkFile *SinkFile) getBackupName(timestamp string) string {
	ext := filepath.Ext(sinkFile.filename)
	if ext == "" {
		return fmt.Sprintf("%s-%s.log", sinkFile.filename, timestamp)
	}
	nameWithoutExt := sinkFile.filename[:len(sinkFile.filename)-len(ext)]
	return fmt.Sprintf("%s-%s%s", nameWithoutExt, timestamp, ext)
}
func (sinkFile *SinkFile) getBackupPattern() string {
	base := filepath.Base(sinkFile.filename)
	dir := filepath.Dir(sinkFile.filename)
	ext := filepath.Ext(sinkFile.filename)
	if ext == "" {
		return filepath.Join(dir, base+"-*.log*")
	}
	nameWithoutExt := base[:len(base)-len(ext)]
	return filepath.Join(dir, nameWithoutExt+"-*.log*")
}
func (sinkFile *SinkFile) rotate() error {
	if !sinkFile.rotating.CompareAndSwap(false, true) {
		return nil
	}
	defer sinkFile.rotating.Store(false)
	sinkFile.mutex.Lock()
	if sinkFile.bufWriter != nil {
		if err := sinkFile.bufWriter.Flush(); err != nil {
			fmt.Fprintf(DefaultWriterErr, "ulog: failed to flush before rotation: %v\n", err)
		}
	}
	if sinkFile.file != nil {
		sinkFile.file.Close()
		sinkFile.file = nil
	}
	sinkFile.mutex.Unlock()
	timestamp := time.Now().Format("20060102-150405.000000")
	backupName := sinkFile.getBackupName(timestamp)
	if err := os.Rename(sinkFile.filename, backupName); err != nil {
		newFile, openErr := os.OpenFile(sinkFile.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if openErr != nil {
			return fmt.Errorf("rotate failed and reopen failed: %w (original: %v)", openErr, err)
		}
		sinkFile.mutex.Lock()
		sinkFile.file = newFile
		sinkFile.bufWriter = bufio.NewWriterSize(newFile, 64*1024)
		sinkFile.currentSize = 0
		sinkFile.mutex.Unlock()
		return err
	}
	sinkFile.wg.Add(1)
	go func() {
		defer sinkFile.wg.Done()
		if err := sinkFile.compress(backupName); err != nil {
			fmt.Fprintf(DefaultWriterErr, "failed to compress %s: %v\n", backupName, err)
		}
	}()
	newFile, err := os.OpenFile(sinkFile.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	sinkFile.mutex.Lock()
	sinkFile.file = newFile
	sinkFile.bufWriter = bufio.NewWriterSize(newFile, 64*1024)
	sinkFile.currentSize = 0
	sinkFile.mutex.Unlock()
	sinkFile.wg.Add(1)
	go func() {
		defer sinkFile.wg.Done()
		if err := sinkFile.cleanup(); err != nil {
			fmt.Fprintf(DefaultWriterErr, "failed to cleanup backups: %v\n", err)
		}
	}()
	return nil
}
