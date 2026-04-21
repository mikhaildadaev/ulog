package ulog

import (
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
type FileSink struct {
	currentSize int64
	file        *os.File
	filename    string
	maxAge      int
	maxBackups  int
	maxSize     int64
	mutex       sync.Mutex
	rotateMutex sync.Mutex
	rotating    atomic.Bool
	wg          sync.WaitGroup
}

// Публичные конструкторы
func NewFileSink(filename string, params ...fileParams) (*FileSink, error) {
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
	fileSink := &FileSink{
		currentSize: info.Size(),
		file:        file,
		filename:    filename,
		maxAge:      30,
		maxBackups:  10,
		maxSize:     100 * 1024 * 1024,
	}
	for _, param := range params {
		param(fileSink)
	}
	return fileSink, nil
}

// Публичные функции
func WithFileMaxAge(days int) fileParams {
	return func(fileSink *FileSink) {
		fileSink.maxAge = days
	}
}
func WithFileMaxBackups(count int) fileParams {
	return func(fileSink *FileSink) {
		fileSink.maxBackups = count
	}
}
func WithFileMaxSize(sizeMB int) fileParams {
	return func(fileSink *FileSink) {
		fileSink.maxSize = int64(sizeMB) * 1024 * 1024
	}
}

// Публичные методы
func (fileSink *FileSink) Close() error {
	fileSink.wg.Wait()
	fileSink.mutex.Lock()
	defer fileSink.mutex.Unlock()
	if fileSink.file != nil {
		return fileSink.file.Close()
	}
	return nil
}
func (fileSink *FileSink) Sync() error {
	fileSink.mutex.Lock()
	defer fileSink.mutex.Unlock()
	if fileSink.file != nil {
		return fileSink.file.Sync()
	}
	return nil
}
func (fileSink *FileSink) Write(p []byte) (n int, err error) {
	fileSink.mutex.Lock()
	needRotate := fileSink.currentSize+int64(len(p)) > fileSink.maxSize
	fileSink.mutex.Unlock()
	if needRotate {
		if err := fileSink.getRotateFile(); err != nil {
			return 0, err
		}
	}
	fileSink.mutex.Lock()
	defer fileSink.mutex.Unlock()
	if fileSink.file == nil {
		return 0, fmt.Errorf("file is nil")
	}
	n, err = fileSink.file.Write(p)
	if err == nil {
		fileSink.currentSize += int64(n)
	}
	return n, err
}
func (fileSink *FileSink) WriteWithAttributes(attributes writeAttributes, fields []Field) (n int, err error) {
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
	fileSink.mutex.Lock()
	needRotate := fileSink.currentSize+int64(len(data)) > fileSink.maxSize
	fileSink.mutex.Unlock()
	if needRotate {
		if err := fileSink.getRotateFile(); err != nil {
			return 0, err
		}
	}
	fileSink.mutex.Lock()
	defer fileSink.mutex.Unlock()
	if fileSink.file == nil {
		return 0, fmt.Errorf("file is nil")
	}
	n, err = fileSink.file.Write(data)
	if err == nil {
		fileSink.currentSize += int64(n)
	}
	return n, err
}

// Приватные структуры
type fileParams func(*FileSink)

// Приватные функции
func (fileSink *FileSink) cleanupBackups() error {
	pattern := fileSink.getBackupPattern() + ".gz"
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
	if fileSink.maxBackups > 0 && len(files) > fileSink.maxBackups {
		for _, file := range files[fileSink.maxBackups:] {
			if err := os.Remove(file); err != nil {
				fmt.Fprintf(defaultWriterErr, "failed to remove old backup %s: %v\n", file, err)
			}
		}
		files = files[:fileSink.maxBackups]
	}
	if fileSink.maxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -fileSink.maxAge)
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				if err := os.Remove(file); err != nil {
					fmt.Fprintf(defaultWriterErr, "failed to remove old backup %s: %v\n", file, err)
				}
			}
		}
	}
	return nil
}
func (fileSink *FileSink) getBackupName(timestamp string) string {
	ext := filepath.Ext(fileSink.filename)
	if ext == "" {
		return fmt.Sprintf("%s-%s.log", fileSink.filename, timestamp)
	}
	nameWithoutExt := fileSink.filename[:len(fileSink.filename)-len(ext)]
	return fmt.Sprintf("%s-%s%s", nameWithoutExt, timestamp, ext)
}
func (fileSink *FileSink) getBackupPattern() string {
	base := filepath.Base(fileSink.filename)
	dir := filepath.Dir(fileSink.filename)
	ext := filepath.Ext(fileSink.filename)
	if ext == "" {
		return filepath.Join(dir, base+"-*.log*")
	}
	nameWithoutExt := base[:len(base)-len(ext)]
	return filepath.Join(dir, nameWithoutExt+"-*.log*")
}
func (fileSink *FileSink) getCompressFile(filename string) error {
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
	defer dst.Close()
	defer func() {
		if err != nil {
			os.Remove(tmpName)
		}
	}()
	gz := gzip.NewWriter(dst)
	defer gz.Close()
	if _, err = io.Copy(gz, src); err != nil {
		return err
	}
	if err = gz.Close(); err != nil {
		return err
	}
	if err = dst.Sync(); err != nil {
		return err
	}
	gzName := filename + ".gz"
	if err = os.Rename(tmpName, gzName); err != nil {
		return err
	}
	return os.Remove(filename)
}
func (fileSink *FileSink) getRotateFile() error {
	if !fileSink.rotating.CompareAndSwap(false, true) {
		return nil
	}
	defer fileSink.rotating.Store(false)
	fileSink.mutex.Lock()
	if fileSink.file != nil {
		fileSink.file.Close()
		fileSink.file = nil
	}
	fileSink.mutex.Unlock()
	timestamp := time.Now().Format("20060102-150405")
	backupName := fileSink.getBackupName(timestamp)
	if err := os.Rename(fileSink.filename, backupName); err != nil {
		return err
	}
	fileSink.wg.Add(1)
	go func() {
		defer fileSink.wg.Done()
		if err := fileSink.getCompressFile(backupName); err != nil {
			fmt.Fprintf(defaultWriterErr, "failed to compress %s: %v\n", backupName, err)
		}
	}()
	newFile, err := os.OpenFile(fileSink.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fileSink.mutex.Lock()
	fileSink.file = newFile
	fileSink.currentSize = 0
	fileSink.mutex.Unlock()
	fileSink.wg.Add(1)
	go func() {
		defer fileSink.wg.Done()
		if err := fileSink.cleanupBackups(); err != nil {
			fmt.Fprintf(defaultWriterErr, "failed to cleanup backups: %v\n", err)
		}
	}()
	return nil
}
