// Этот файл (sink_http.go) находится в стадии активной разработки.
// API может изменяться
//
// Планируется добавить:
// - Circuit Breaker [защита от перегрузки]
package ulog

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Публичные структуры
type HttpSink struct {
	batchBuffer     [][]byte
	batchChan       chan struct{}
	batchMutex      sync.Mutex
	batchSize       int
	batchTicker     *time.Ticker
	client          *http.Client
	dedupCache      sync.Map
	dedupTTL        time.Duration
	dedupWindow     time.Duration
	endPoint        string
	formatter       func(attributes writeAttributes, p []byte) ([]byte, error)
	headers         map[string]string
	levelMin        TypeLevel
	method          string
	retryBackoff    time.Duration
	retryMax        int
	sampleCounter   int32
	sampleLastReset time.Time
	sampleMutex     sync.Mutex
	sampleRate      int32
	sampleWindow    time.Duration
	typeFilter      TypeData
}
type HttpParams func(*HttpSink)

// Публичные конструкторы
func NewHttpSink(endPoint string, params ...HttpParams) *HttpSink {
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
		typeFilter:   TypeData(defaultType),
	}
	for _, param := range params {
		param(httpSink)
	}
	return httpSink
}

// Публичные функции
func WithHttpBatch(size int, flushInterval time.Duration) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.batchSize = size
		httpSink.batchTicker = time.NewTicker(flushInterval)
		go httpSink.batchLoop()
	}
}
func WithHttpDedupWindow(window time.Duration) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.dedupWindow = window
	}
}
func WithHttpFormatter(formatter func(attributes writeAttributes, p []byte) ([]byte, error)) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.formatter = formatter
	}
}
func WithHttpHeader(key, value string) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.headers[key] = value
	}
}
func WithHttpLevelMin(level TypeLevel) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.levelMin = level
	}
}
func WithHttpMethod(method string) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.method = method
	}
}
func WithHttpTypeFilter(typeData TypeData) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.typeFilter = typeData
	}
}
func WithHttpRetry(maxRetries int, backoff time.Duration) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.retryMax = maxRetries
		httpSink.retryBackoff = backoff
	}
}
func WithHttpSampleRate(rate int32) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.sampleRate = rate
	}
}
func WithHttpSampleWindow(window time.Duration) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.sampleWindow = window
	}
}
func WithHttpTimeout(timeout time.Duration) HttpParams {
	return func(httpSink *HttpSink) {
		httpSink.client.Timeout = timeout
	}
}

// Публичные методы
func (rateLimitError *rateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", rateLimitError.retryAfter)
}
func (httpSink *HttpSink) Close() error {
	httpSink.batchMutex.Lock()
	batchSize := httpSink.batchSize
	ticker := httpSink.batchTicker
	httpSink.batchMutex.Unlock()
	if batchSize > 0 {
		close(httpSink.batchChan)
		if ticker != nil {
			ticker.Stop()
		}
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
	attributes := writeAttributes{
		typeData:  DataLog,
		typeLevel: LevelDebug,
	}
	return httpSink.WriteWithAttributes(attributes, p)
}
func (httpSink *HttpSink) WriteWithAttributes(attributes writeAttributes, p []byte) (n int, err error) {
	if attributes.typeLevel < httpSink.levelMin {
		return len(p), nil
	}
	if attributes.typeData != httpSink.typeFilter && httpSink.typeFilter >= 0 {
		return len(p), nil
	}
	if attributes.typeLevel != LevelError && attributes.typeLevel != LevelFatal {
		if !httpSink.shouldSample() {
			return len(p), nil
		}
		if httpSink.isDuplicate(p) {
			return len(p), nil
		}
	}
	body, err := httpSink.formatter(attributes, p)
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

// Приватные структуры
type rateLimitError struct {
	retryAfter time.Duration
}

// Приватные функции
func defaultformatter(attributes writeAttributes, p []byte) ([]byte, error) {
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
func (httpSink *HttpSink) isDuplicate(p []byte) bool {
	if httpSink.dedupWindow <= 0 {
		return false
	}
	hash := httpSink.hashMessage(p)
	if lastSeen, ok := httpSink.dedupCache.Load(hash); ok {
		if time.Since(lastSeen.(time.Time)) < httpSink.dedupWindow {
			return true
		}
	}
	httpSink.dedupCache.Store(hash, time.Now())
	return false
}
func (httpSink *HttpSink) hashMessage(p []byte) uint64 {
	hash := fnv.New64a()
	hash.Write(p)
	return hash.Sum64()
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
		var retryAfter time.Duration
		if retryAfterHeader := resp.Header.Get("Retry-After"); retryAfterHeader != "" {
			if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			} else if t, err := http.ParseTime(retryAfterHeader); err == nil {
				retryAfter = time.Until(t)
			}
		}
		if retryAfter == 0 {
			retryAfter = 5 * time.Second
		}
		return &rateLimitError{retryAfter: retryAfter}
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
		if rateErr, ok := err.(*rateLimitError); ok {
			sleepDuration = rateErr.retryAfter
		} else {
			sleepDuration = httpSink.retryBackoff * time.Duration(1<<i)
		}
		time.Sleep(sleepDuration)
	}
	return fmt.Errorf("failed after %d retries: %w", httpSink.retryMax, lastErr)
}
func (httpSink *HttpSink) shouldSample() bool {
	if httpSink.sampleRate <= 1 {
		return true
	}
	httpSink.sampleMutex.Lock()
	defer httpSink.sampleMutex.Unlock()
	if httpSink.sampleWindow > 0 && time.Since(httpSink.sampleLastReset) > httpSink.sampleWindow {
		httpSink.sampleCounter = 0
		httpSink.sampleLastReset = time.Now()
	}
	httpSink.sampleCounter++
	return httpSink.sampleCounter%httpSink.sampleRate == 0
}
