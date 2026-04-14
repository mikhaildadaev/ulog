// Этот файл (sink_http.go) находится в стадии активной разработки.
// API может изменяться
//
// Планируется добавить:
// - Circuit Breaker [защита от перегрузки]
package ulog

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Публичные структуры
type HttpSink struct {
	batchBuffer  [][]byte
	batchChan    chan struct{}
	batchMutex   sync.Mutex
	batchSize    int
	batchTicker  *time.Ticker
	client       *http.Client
	endPoint     string
	formatter    func(level TypeLevel, p []byte) ([]byte, error)
	headers      map[string]string
	levelMin     TypeLevel
	method       string
	retryAfter   time.Duration
	retryBackoff time.Duration
	retryMax     int
}
type HttpOption func(*HttpSink)

// Публичные конструкторы
func NewHttpSink(endPoint string, options ...HttpOption) *HttpSink {
	httpSink := &HttpSink{
		batchChan:    make(chan struct{}),
		client:       &http.Client{Timeout: 10 * time.Second},
		endPoint:     endPoint,
		formatter:    defaultformatter,
		headers:      make(map[string]string),
		levelMin:     LevelError,
		method:       "POST",
		retryBackoff: time.Second,
		retryMax:     0,
	}
	for _, option := range options {
		option(httpSink)
	}
	return httpSink
}

// Публичные функции
func WithHttpBatch(size int, flushInterval time.Duration) HttpOption {
	return func(httpSink *HttpSink) {
		httpSink.batchSize = size
		httpSink.batchTicker = time.NewTicker(flushInterval)
		go httpSink.batchLoop()
	}
}
func WithHttpFormatter(formatter func(level TypeLevel, p []byte) ([]byte, error)) HttpOption {
	return func(httpSink *HttpSink) {
		httpSink.formatter = formatter
	}
}
func WithHttpHeader(key, value string) HttpOption {
	return func(httpSink *HttpSink) {
		httpSink.headers[key] = value
	}
}
func WithHttpLevelMin(level TypeLevel) HttpOption {
	return func(httpSink *HttpSink) {
		httpSink.levelMin = level
	}
}
func WithHttpMethod(method string) HttpOption {
	return func(httpSink *HttpSink) {
		httpSink.method = method
	}
}
func WithHttpRetry(maxRetries int, backoff time.Duration) HttpOption {
	return func(httpSink *HttpSink) {
		httpSink.retryMax = maxRetries
		httpSink.retryBackoff = backoff
	}
}
func WithHttpTimeout(timeout time.Duration) HttpOption {
	return func(httpSink *HttpSink) {
		httpSink.client.Timeout = timeout
	}
}

// Публичные методы
func (httpSink *HttpSink) Close() error {
	if httpSink.batchSize > 0 {
		close(httpSink.batchChan)
		httpSink.batchTicker.Stop()
	}
	httpSink.client.CloseIdleConnections()
	return nil
}
func (httpSink *HttpSink) Sync() error {
	if httpSink.batchSize > 0 {
		return httpSink.flush()
	}
	return nil
}
func (httpSink *HttpSink) Write(p []byte) (n int, err error) {
	return httpSink.WriteWithLevel(LevelDebug, p)
}
func (httpSink *HttpSink) WriteWithLevel(level TypeLevel, p []byte) (n int, err error) {
	if level < httpSink.levelMin {
		return len(p), nil
	}
	body, err := httpSink.formatter(level, p)
	if err != nil {
		return len(p), fmt.Errorf("formatter error: %w", err)
	}
	if httpSink.batchSize > 0 {
		httpSink.batchMutex.Lock()
		httpSink.batchBuffer = append(httpSink.batchBuffer, body)
		needFlush := len(httpSink.batchBuffer) >= httpSink.batchSize
		httpSink.batchMutex.Unlock()
		if needFlush {
			go httpSink.flush()
		}
		return len(p), nil
	}
	return len(p), httpSink.sendWithRetry(body)
}

// Приватные функции
func defaultformatter(level TypeLevel, p []byte) ([]byte, error) {
	return p, nil
}

// Приватные методы
func (httpSink *HttpSink) batchLoop() {
	for {
		select {
		case <-httpSink.batchTicker.C:
			httpSink.flush()
		case <-httpSink.batchChan:
			httpSink.flush()
			return
		}
	}
}
func (httpSink *HttpSink) flush() error {
	httpSink.batchMutex.Lock()
	if len(httpSink.batchBuffer) == 0 {
		httpSink.batchMutex.Unlock()
		return nil
	}
	batch := make([][]byte, len(httpSink.batchBuffer))
	copy(batch, httpSink.batchBuffer)
	httpSink.batchBuffer = httpSink.batchBuffer[:0]
	httpSink.batchMutex.Unlock()
	var body []byte
	if len(batch) == 1 {
		body = batch[0]
	} else {
		var parts [][]byte
		for _, b := range batch {
			parts = append(parts, b)
		}
		body = bytes.Join(parts, []byte{'\n'})
	}
	return httpSink.sendWithRetry(body)
}
func (httpSink *HttpSink) send(body []byte) error {
	req, err := http.NewRequest(httpSink.method, httpSink.endPoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range httpSink.headers {
		req.Header.Set(k, v)
	}
	resp, err := httpSink.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				httpSink.retryAfter = time.Duration(seconds) * time.Second
			} else if t, err := http.ParseTime(retryAfter); err == nil {
				httpSink.retryAfter = time.Until(t)
			}
		}
		if httpSink.retryAfter == 0 {
			httpSink.retryAfter = 5 * time.Second
		}
		return fmt.Errorf("rate limited")
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		httpSink.retryAfter = 0
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	return nil
}
func (httpSink *HttpSink) sendWithRetry(body []byte) error {
	var lastErr error
	for i := 0; i <= httpSink.retryMax; i++ {
		err := httpSink.send(body)
		if err == nil {
			return nil
		}
		lastErr = err
		if i == httpSink.retryMax {
			break
		}
		var sleepDuration time.Duration
		if strings.Contains(err.Error(), "rate limited") {
			sleepDuration = httpSink.retryAfter
			if sleepDuration == 0 {
				sleepDuration = 5 * time.Second
			}
		} else {
			sleepDuration = httpSink.retryBackoff * time.Duration(1<<i)
		}
		time.Sleep(sleepDuration)
	}
	return fmt.Errorf("failed after %d retries: %w", httpSink.retryMax, lastErr)
}
