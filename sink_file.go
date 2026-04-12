package ulog

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Публичные структуры
type FileSink struct {
	file       *os.File
	filename   string
	maxAge     int
	maxBackups int
	maxSize    int64
	mutex      sync.Mutex
}
type FileOption func(*FileSink)

// Публичные конструкторы
func NewFileSink(filename string, options ...FileOption) (*FileSink, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	fileSink := &FileSink{
		file:       file,
		filename:   filename,
		maxAge:     30,
		maxBackups: 10,
		maxSize:    100 * 1024 * 1024,
	}
	for _, option := range options {
		option(fileSink)
	}
	return fileSink, nil
}

// Публичные функции
func WithFileMaxAge(days int) FileOption {
	return func(fileSink *FileSink) {
		fileSink.maxAge = days
	}
}
func WithFileMaxBackups(count int) FileOption {
	return func(fileSink *FileSink) {
		fileSink.maxBackups = count
	}
}
func WithFileMaxSize(sizeMB int) FileOption {
	return func(fileSink *FileSink) {
		fileSink.maxSize = int64(sizeMB) * 1024 * 1024
	}
}

// Публичные методы
func (fileSink *FileSink) Close() error {
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
	defer fileSink.mutex.Unlock()
	info, err := fileSink.file.Stat()
	if err != nil {
		return 0, err
	}
	if info.Size()+int64(len(p)) > fileSink.maxSize {
		if err := fileSink.rotate(); err != nil {
			return 0, err
		}
	}
	return fileSink.file.Write(p)
}

// Приватные функции
func (fileSink *FileSink) cleanupBackups() error {
	pattern := fileSink.getBackupPattern()
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
				fmt.Fprintf(os.Stderr, "failed to remove old backup %s: %v\n", file, err)
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
					fmt.Fprintf(os.Stderr, "failed to remove old backup %s: %v\n", file, err)
				}
			}
		}
	}
	return nil
}
func (fileSink *FileSink) compress(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	gzFilename := filename + ".gz"
	gzFile, err := os.Create(gzFilename)
	if err != nil {
		return err
	}
	defer gzFile.Close()
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()
	if _, err := io.Copy(gzWriter, file); err != nil {
		return err
	}
	return os.Remove(filename)
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
func (fileSink *FileSink) rotate() error {
	if err := fileSink.file.Close(); err != nil {
		return err
	}
	timestamp := time.Now().Format("20060102-150405")
	backupName := fileSink.getBackupName(timestamp)
	if err := os.Rename(fileSink.filename, backupName); err != nil {
		return err
	}
	go func() {
		if err := fileSink.compress(backupName); err != nil {
			fmt.Fprintf(os.Stderr, "failed to compress %s: %v\n", backupName, err)
		}
	}()
	file, err := os.OpenFile(fileSink.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fileSink.file = file
	go func() {
		if err := fileSink.cleanupBackups(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to cleanup backups: %v\n", err)
		}
	}()
	return nil
}
