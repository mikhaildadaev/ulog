package ulog

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Публичные структуры
type HttpSink struct {
	batchBuffer        [][]byte
	batchChan          chan struct{}
	batchMutex         sync.Mutex
	batchSize          int
	batchTicker        *time.Ticker
	circuitEnabled     bool
	circuitFailures    int32
	circuitMaxFailures int
	circuitLastFailure atomic.Int64
	circuitMutex       sync.Mutex
	circuitState       int32
	circuitTimeout     time.Duration
	client             *http.Client
	closed             bool
	dedupCache         sync.Map
	dedupStopChan      chan struct{}
	dedupWindow        time.Duration
	endPoint           string
	filterData         TypeData
	filterLevel        TypeLevel
	formatter          func(attributes writeAttributes, fields []Field) ([]byte, error)
	headers            map[string]string
	method             string
	mutex              sync.Mutex
	retryBackoff       time.Duration
	retryMax           int
	sampleCounter      int32
	sampleLastReset    time.Time
	sampleMutex        sync.Mutex
	sampleRate         int32
	sampleWindow       time.Duration
	wg                 sync.WaitGroup
}

// Публичные конструкторы
func NewHttpSink(endPoint string, params ...httpParams) *HttpSink {
	httpSink := &HttpSink{
		batchChan:          make(chan struct{}),
		batchSize:          100,
		batchTicker:        time.NewTicker(5 * time.Second),
		circuitEnabled:     true,
		circuitMaxFailures: 10,
		circuitState:       circuitStateClosed,
		circuitTimeout:     10 * time.Second,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
		dedupStopChan: make(chan struct{}),
		endPoint:      endPoint,
		filterData:    TypeData(defaultType),
		filterLevel:   LevelError,
		formatter:     defaultformatter,
		headers:       make(map[string]string),
		method:        "POST",
		retryBackoff:  time.Second,
		retryMax:      0,
	}
	for _, param := range params {
		param(httpSink)
	}
	if httpSink.dedupWindow > 0 {
		go httpSink.cleanupDedupCache()
	}
	return httpSink
}

// Публичные функции
func WithHttpBatch(size int, flushInterval time.Duration) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.batchSize = size
		httpSink.batchTicker = time.NewTicker(flushInterval)
		go httpSink.batchLoop()
	}
}
func WithHttpCircuitBreaker(maxFailures int, timeout time.Duration) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.circuitEnabled = true
		httpSink.circuitMaxFailures = maxFailures
		httpSink.circuitState = circuitStateClosed
		httpSink.circuitTimeout = timeout
	}
}
func WithHttpDedupWindow(window time.Duration) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.dedupWindow = window
	}
}
func WithHttpDisabledBatch() httpParams {
	return func(httpSink *HttpSink) {
		httpSink.batchSize = 0
		httpSink.batchTicker = nil
	}
}
func WithHttpDisabledCircuit() httpParams {
	return func(httpSink *HttpSink) {
		httpSink.circuitEnabled = false
	}
}
func WithHttpDisableKeepAlive() httpParams {
	return func(httpSink *HttpSink) {
		if transport, ok := httpSink.client.Transport.(*http.Transport); ok {
			transport.DisableKeepAlives = true
		}
	}
}
func WithHttpFilterData(typeData TypeData) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.filterData = typeData
	}
}
func WithHttpFilterLevel(level TypeLevel) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.filterLevel = level
	}
}
func WithHttpFormatter(formatter func(attributes writeAttributes, fields []Field) ([]byte, error)) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.formatter = formatter
	}
}
func WithHttpHeader(key, value string) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.headers[key] = value
	}
}
func WithHttpMethod(method string) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.method = method
	}
}
func WithHttpRetry(maxRetries int, backoff time.Duration) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.retryMax = maxRetries
		httpSink.retryBackoff = backoff
	}
}
func WithHttpSampleRate(rate int32) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.sampleRate = rate
	}
}
func WithHttpSampleWindow(window time.Duration) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.sampleWindow = window
	}
}
func WithHttpTimeout(timeout time.Duration) httpParams {
	return func(httpSink *HttpSink) {
		httpSink.client.Timeout = timeout
	}
}

// Публичные методы
func (httpSink *HttpSink) Close() error {
	httpSink.mutex.Lock()
	if httpSink.closed {
		httpSink.mutex.Unlock()
		return nil
	}
	httpSink.closed = true
	httpSink.mutex.Unlock()
	if httpSink.batchSize > 0 {
		httpSink.wg.Add(1)
		go func() {
			defer httpSink.wg.Done()
			httpSink.flush()
		}()
	}
	if httpSink.dedupStopChan != nil {
		close(httpSink.dedupStopChan)
	}
	httpSink.batchMutex.Lock()
	ticker := httpSink.batchTicker
	httpSink.batchMutex.Unlock()
	if httpSink.batchSize > 0 {
		close(httpSink.batchChan)
		if ticker != nil {
			ticker.Stop()
		}
	}
	done := make(chan struct{})
	go func() {
		httpSink.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
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
	return httpSink.sendWithRetry(p)
}
func (httpSink *HttpSink) WriteWithAttributes(attributes writeAttributes, fields []Field) (n int, err error) {
	if attributes.typeLevel < httpSink.filterLevel {
		return 0, nil
	}
	if attributes.typeData != httpSink.filterData && httpSink.filterData > TypeData(defaultType) {
		return 0, nil
	}
	if attributes.typeLevel != LevelError && attributes.typeLevel != LevelFatal {
		if !httpSink.shouldSample() {
			return 0, nil
		}
		if httpSink.isDuplicate(fields) {
			return 0, nil
		}
	}
	body, err := httpSink.formatter(attributes, fields)
	if err != nil {
		return 0, fmt.Errorf("formatter error: %w", err)
	}
	if httpSink.batchSize > 0 {
		httpSink.batchMutex.Lock()
		httpSink.batchBuffer = append(httpSink.batchBuffer, body)
		needFlush := len(httpSink.batchBuffer) >= httpSink.batchSize
		httpSink.batchMutex.Unlock()
		if needFlush {
			go httpSink.flush()
		}
		return len(body), nil
	}
	return httpSink.sendWithRetry(body)
}

// Приватные константы
const (
	circuitStateClosed = iota
	circuitStateOpen
	circuitStateHalfOpen
)

// Приватные структуры
type rateLimitError struct {
	retryAfter time.Duration
}
type httpParams func(*HttpSink)

// Приватные функции
func defaultformatter(attributes writeAttributes, fields []Field) ([]byte, error) {
	buf := &bytes.Buffer{}
	formatJson(buf, attributes, fields)
	return buf.Bytes(), nil
}
func getLogMessage(fields []Field) string {
	for _, field := range fields {
		if field.nameKey == "message" {
			return field.valueString
		}
	}
	return ""
}
func getMetricData(fields []Field) (name string, value float64, labels map[string]string) {
	name = ""
	value = 0
	labels = make(map[string]string)
	for _, field := range fields {
		switch field.nameKey {
		case "name":
			name = field.valueString
		case "value":
			value = field.valueFloat64
		default:
			labels[field.nameKey] = field.valueString
		}
	}
	return name, value, labels
}
func getTraceID(fields []Field) string {
	for _, f := range fields {
		if f.nameKey == "trace_id" {
			return f.valueString
		}
	}
	return ""
}
func getTraceDuration(fields []Field) int64 {
	for _, f := range fields {
		if f.nameKey == "duration" {
			return f.valueInt64
		}
	}
	return 0
}
func getTraceName(fields []Field) string {
	for _, f := range fields {
		if f.nameKey == "name" {
			return f.valueString
		}
	}
	return ""
}
func getTraceSpanID(fields []Field) string {
	for _, f := range fields {
		if f.nameKey == "span_id" {
			return f.valueString
		}
	}
	return ""
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
func (httpSink *HttpSink) circuitAllow() bool {
	if !httpSink.circuitEnabled {
		return true
	}
	state := atomic.LoadInt32(&httpSink.circuitState)
	switch state {
	case circuitStateClosed:
		return true
	case circuitStateOpen:
		lastFailure := httpSink.circuitLastFailure.Load()
		if time.Now().UnixNano()-lastFailure > httpSink.circuitTimeout.Nanoseconds() {
			httpSink.circuitMutex.Lock()
			if atomic.LoadInt32(&httpSink.circuitState) == circuitStateOpen {
				atomic.StoreInt32(&httpSink.circuitState, circuitStateHalfOpen)
			}
			httpSink.circuitMutex.Unlock()
			return true
		}
		return false
	case circuitStateHalfOpen:
		return false
	default:
		return true
	}
}
func (httpSink *HttpSink) circuitRecord(success bool) {
	if !httpSink.circuitEnabled {
		return
	}
	state := atomic.LoadInt32(&httpSink.circuitState)
	switch state {
	case circuitStateClosed:
		if !success {
			failures := atomic.AddInt32(&httpSink.circuitFailures, 1)
			httpSink.circuitLastFailure.Store(time.Now().UnixNano())
			if int(failures) >= httpSink.circuitMaxFailures {
				httpSink.circuitMutex.Lock()
				if atomic.LoadInt32(&httpSink.circuitState) == circuitStateClosed {
					atomic.StoreInt32(&httpSink.circuitState, circuitStateOpen)
					atomic.StoreInt32(&httpSink.circuitFailures, 0)
				}
				httpSink.circuitMutex.Unlock()
			}
		} else {
			httpSink.circuitMutex.Lock()
			if atomic.LoadInt32(&httpSink.circuitState) == circuitStateClosed {
				atomic.StoreInt32(&httpSink.circuitFailures, 0)
			}
			httpSink.circuitMutex.Unlock()
		}
	case circuitStateHalfOpen:
		httpSink.circuitMutex.Lock()
		defer httpSink.circuitMutex.Unlock()
		if success {
			atomic.StoreInt32(&httpSink.circuitState, circuitStateClosed)
			atomic.StoreInt32(&httpSink.circuitFailures, 0)
		} else {
			atomic.StoreInt32(&httpSink.circuitState, circuitStateOpen)
			httpSink.circuitLastFailure.Store(time.Now().UnixNano())
		}
	case circuitStateOpen:
	}
}
func (httpSink *HttpSink) cleanupDedupCache() {
	ticker := time.NewTicker(httpSink.dedupWindow)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			httpSink.dedupCache.Range(func(key, value any) bool {
				if now.Sub(value.(time.Time)) > httpSink.dedupWindow {
					httpSink.dedupCache.Delete(key)
				}
				return true
			})
		case <-httpSink.dedupStopChan:
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
	_, err := httpSink.sendWithRetry(body)
	return err
}
func (httpSink *HttpSink) isDuplicate(fields []Field) bool {
	if httpSink.dedupWindow <= 0 {
		return false
	}
	hash := httpSink.hashFields(fields)
	if lastSeen, ok := httpSink.dedupCache.Load(hash); ok {
		if time.Since(lastSeen.(time.Time)) < httpSink.dedupWindow {
			return true
		}
	}
	httpSink.dedupCache.Store(hash, time.Now())
	return false
}
func (httpSink *HttpSink) hashFields(fields []Field) uint64 {
	hash := fnv.New64a()
	for _, f := range fields {
		hash.Write([]byte(f.nameKey))
		hash.Write([]byte{0})
		hash.Write([]byte(fmt.Sprintf("%v", f.valueString)))
		hash.Write([]byte{0})
	}
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
func (httpSink *HttpSink) sendWithRetry(body []byte) (int, error) {
	var lastErr error
	for i := 0; i <= httpSink.retryMax; i++ {
		if !httpSink.circuitAllow() {
			return 0, fmt.Errorf("circuit breaker is open")
		}
		err := httpSink.send(body)
		httpSink.circuitRecord(err == nil)
		if err == nil {
			return len(body), nil
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
	return 0, fmt.Errorf("failed after %d retries: %w", httpSink.retryMax, lastErr)
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
func (rateLimitError *rateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", rateLimitError.retryAfter)
}
