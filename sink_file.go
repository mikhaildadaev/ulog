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
type SinkFile struct {
	currentSize int64
	file        *os.File
	filename    string
	maxAge      int
	maxBackups  int
	maxSize     int64
	mutex       sync.Mutex
	rotating    atomic.Bool
	wg          sync.WaitGroup
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
		currentSize: info.Size(),
		file:        file,
		filename:    filename,
		maxAge:      30,
		maxBackups:  10,
		maxSize:     100 * 1024 * 1024,
	}
	for _, param := range params {
		param(sinkFile)
	}
	return sinkFile, nil
}

// Публичные функции
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
	sinkFile.wg.Wait()
	sinkFile.mutex.Lock()
	defer sinkFile.mutex.Unlock()
	if sinkFile.file != nil {
		return sinkFile.file.Close()
	}
	return nil
}
func (sinkFile *SinkFile) Sync() error {
	sinkFile.mutex.Lock()
	defer sinkFile.mutex.Unlock()
	if sinkFile.file != nil {
		return sinkFile.file.Sync()
	}
	return nil
}
func (sinkFile *SinkFile) Write(p []byte) (n int, err error) {
	sinkFile.mutex.Lock()
	needRotate := sinkFile.currentSize+int64(len(p)) > sinkFile.maxSize
	sinkFile.mutex.Unlock()
	if needRotate {
		if err := sinkFile.getRotateFile(); err != nil {
			return 0, err
		}
	}
	for {
		sinkFile.mutex.Lock()
		if sinkFile.file != nil {
			break
		}
		sinkFile.mutex.Unlock()
		time.Sleep(time.Microsecond)
	}
	defer sinkFile.mutex.Unlock()
	n, err = sinkFile.file.Write(p)
	if err == nil {
		sinkFile.currentSize += int64(n)
	}
	return n, err
}
func (sinkFile *SinkFile) WriteWithAttributes(attributes writeAttributes, fields []Field) (n int, err error) {
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
		if err := sinkFile.getRotateFile(); err != nil {
			return 0, err
		}
	}
	for {
		sinkFile.mutex.Lock()
		if sinkFile.file != nil {
			break
		}
		sinkFile.mutex.Unlock()
		time.Sleep(time.Microsecond)
	}
	defer sinkFile.mutex.Unlock()
	n, err = sinkFile.file.Write(data)
	if err == nil {
		sinkFile.currentSize += int64(n)
	}
	return n, err
}

// Приватные структуры
type fileParams func(*SinkFile)

// Приватные функции
func (sinkFile *SinkFile) cleanupBackups() error {
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
func (fileSink *SinkFile) getCompressFile(filename string) error {
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
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	err = os.Remove(filename)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
func (sinkFile *SinkFile) getRotateFile() error {
	if !sinkFile.rotating.CompareAndSwap(false, true) {
		return nil
	}
	defer sinkFile.rotating.Store(false)
	sinkFile.mutex.Lock()
	if sinkFile.file != nil {
		sinkFile.file.Close()
		sinkFile.file = nil
	}
	sinkFile.mutex.Unlock()
	timestamp := time.Now().Format("20060102-150405")
	backupName := sinkFile.getBackupName(timestamp)
	if err := os.Rename(sinkFile.filename, backupName); err != nil {
		return err
	}
	sinkFile.wg.Add(1)
	go func() {
		defer sinkFile.wg.Done()
		if err := sinkFile.getCompressFile(backupName); err != nil {
			fmt.Fprintf(DefaultWriterErr, "failed to compress %s: %v\n", backupName, err)
		}
	}()
	newFile, err := os.OpenFile(sinkFile.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	sinkFile.mutex.Lock()
	sinkFile.file = newFile
	sinkFile.currentSize = 0
	sinkFile.mutex.Unlock()
	sinkFile.wg.Add(1)
	go func() {
		defer sinkFile.wg.Done()
		if err := sinkFile.cleanupBackups(); err != nil {
			fmt.Fprintf(DefaultWriterErr, "failed to cleanup backups: %v\n", err)
		}
	}()
	return nil
}
