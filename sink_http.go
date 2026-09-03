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
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Публичные структуры
type SinkHttp struct {
	batchBuffer        [][]byte
	batchChan          chan struct{}
	batchMutex         sync.Mutex
	batchSize          int
	batchTicker        *time.Ticker
	circuitEnabled     bool
	circuitFailures    atomic.Int32
	circuitMaxFailures int
	circuitLastFailure atomic.Int64
	circuitMutex       sync.Mutex
	circuitState       atomic.Int32
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
	once               sync.Once
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
func NewSinkHttp(endPoint string, params ...httpParams) *SinkHttp {
	sinkHttp := &SinkHttp{
		batchChan:          make(chan struct{}),
		batchSize:          100,
		batchTicker:        time.NewTicker(5 * time.Second),
		circuitEnabled:     true,
		circuitMaxFailures: 10,
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
	sinkHttp.circuitState.Store(circuitStateClosed)
	sinkHttp.circuitFailures.Store(0)
	for _, param := range params {
		param(sinkHttp)
	}
	if sinkHttp.dedupWindow > 0 {
		go sinkHttp.cleanupDedupCache()
	}
	return sinkHttp
}

// Публичные функции
func WithHttpBatch(size int, flushInterval time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.batchSize = size
		sinkHttp.batchTicker = time.NewTicker(flushInterval)
		go sinkHttp.batchLoop()
	}
}
func WithHttpCircuitBreaker(maxFailures int, timeout time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.circuitEnabled = true
		sinkHttp.circuitMaxFailures = maxFailures
		sinkHttp.circuitState.Store(circuitStateClosed)
		sinkHttp.circuitTimeout = timeout
	}
}
func WithHttpDedupWindow(window time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.dedupWindow = window
	}
}
func WithHttpDisabledBatch() httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.batchSize = 0
		sinkHttp.batchTicker = nil
	}
}
func WithHttpDisabledCircuit() httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.circuitEnabled = false
	}
}
func WithHttpDisableKeepAlive() httpParams {
	return func(sinkHttp *SinkHttp) {
		if transport, ok := sinkHttp.client.Transport.(*http.Transport); ok {
			transport.DisableKeepAlives = true
		}
	}
}
func WithHttpFilterData(typeData TypeData) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.filterData = typeData
	}
}
func WithHttpFilterLevel(level TypeLevel) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.filterLevel = level
	}
}
func WithHttpFormatter(formatter func(attributes writeAttributes, fields []Field) ([]byte, error)) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.formatter = formatter
	}
}
func WithHttpHeader(key, value string) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.headers[key] = value
	}
}
func WithHttpMethod(method string) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.method = method
	}
}
func WithHttpRetry(maxRetries int, backoff time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.retryMax = maxRetries
		sinkHttp.retryBackoff = backoff
	}
}
func WithHttpSampleRate(rate int32) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.sampleRate = rate
	}
}
func WithHttpSampleWindow(window time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.sampleWindow = window
	}
}
func WithHttpTimeout(timeout time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.client.Timeout = timeout
	}
}

// Публичные методы
func (sinkHttp *SinkHttp) Close() error {
	var err error
	sinkHttp.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("ulog: panic during SinkHttp.Close(): %v", r)
			}
		}()
		sinkHttp.mutex.Lock()
		if sinkHttp.closed {
			sinkHttp.mutex.Unlock()
			return
		}
		sinkHttp.closed = true
		sinkHttp.mutex.Unlock()
		if sinkHttp.batchSize > 0 {
			sinkHttp.wg.Add(1)
			go func() {
				defer sinkHttp.wg.Done()
				done := make(chan struct{})
				go func() {
					sinkHttp.flush()
					close(done)
				}()
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					fmt.Fprintf(DefaultWriterErr, "ulog: SinkHttp.flush() timeout\n")
				}
			}()
		}
		if sinkHttp.dedupStopChan != nil {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(DefaultWriterErr, "ulog: panic closing dedupStopChan: %v\n", r)
				}
			}()
			close(sinkHttp.dedupStopChan)
		}
		sinkHttp.batchMutex.Lock()
		ticker := sinkHttp.batchTicker
		sinkHttp.batchMutex.Unlock()
		if sinkHttp.batchSize > 0 {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(DefaultWriterErr, "ulog: panic closing batchChan: %v\n", r)
				}
			}()
			close(sinkHttp.batchChan)
			if ticker != nil {
				ticker.Stop()
			}
		}
		done := make(chan struct{})
		go func() {
			sinkHttp.wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(maxWaitHttp):
			fmt.Fprintf(DefaultWriterErr, "ulog: SinkHttp.Close() timeout waiting for wg\n")
		}
		sinkHttp.client.CloseIdleConnections()
	})
	return err
}
func (sinkHttp *SinkHttp) Write(p []byte) (n int, err error) {
	return sinkHttp.sendWithRetry(p)
}
func (sinkHttp *SinkHttp) WriteWithAttributes(attributes writeAttributes, fields []Field) (n int, err error) {
	if attributes.typeLevel < sinkHttp.filterLevel {
		return 0, nil
	}
	if attributes.typeData != sinkHttp.filterData && sinkHttp.filterData > TypeData(defaultType) {
		return 0, nil
	}
	if attributes.typeLevel != LevelError && attributes.typeLevel != LevelFatal {
		if !sinkHttp.shouldSample() {
			return 0, nil
		}
		if sinkHttp.isDuplicate(fields) {
			return 0, nil
		}
	}
	body, err := sinkHttp.formatter(attributes, fields)
	if err != nil {
		return 0, fmt.Errorf("formatter error: %w", err)
	}
	if sinkHttp.batchSize > 0 {
		sinkHttp.batchMutex.Lock()
		sinkHttp.batchBuffer = append(sinkHttp.batchBuffer, body)
		needFlush := len(sinkHttp.batchBuffer) >= sinkHttp.batchSize
		sinkHttp.batchMutex.Unlock()
		if needFlush {
			go sinkHttp.flush()
		}
		return len(body), nil
	}
	return sinkHttp.sendWithRetry(body)
}

// Приватные константы
const (
	circuitStateClosed = iota
	circuitStateOpen
	circuitStateHalfOpen
)

// Приватные переменные
var fieldExtractor = map[TypeField]func(Field) any{
	FieldString:   func(field Field) any { return field.valueString },
	FieldInt:      func(field Field) any { return field.valueInt },
	FieldInt64:    func(field Field) any { return field.valueInt64 },
	FieldFloat64:  func(field Field) any { return field.valueFloat64 },
	FieldBool:     func(field Field) any { return field.valueBool },
	FieldDuration: func(field Field) any { return field.valueDuration.String() },
	FieldTime:     func(field Field) any { return field.valueTime.Format(time.RFC3339Nano) },
	FieldStrings:  func(field Field) any { return field.valueStrings },
	FieldInts:     func(field Field) any { return field.valueInts },
	FieldInts64:   func(field Field) any { return field.valueInts64 },
	FieldFloats64: func(field Field) any { return field.valueFloats64 },
	FieldBools:    func(field Field) any { return field.valueBools },
	FieldDurations: func(field Field) any {
		result := make([]string, len(field.valueDurations))
		for i, d := range field.valueDurations {
			result[i] = d.String()
		}
		return result
	},
	FieldTimes: func(field Field) any {
		result := make([]string, len(field.valueTimes))
		for i, t := range field.valueTimes {
			result[i] = t.Format(time.RFC3339Nano)
		}
		return result
	},
}

// Приватные структуры
type rateLimitError struct {
	retryAfter time.Duration
}
type httpParams func(*SinkHttp)

// Приватные функции
func defaultformatter(attributes writeAttributes, fields []Field) ([]byte, error) {
	buf := &bytes.Buffer{}
	formatJson(buf, attributes, fields)
	return buf.Bytes(), nil
}
func getLogField(field Field) any {
	if extractor, ok := fieldExtractor[field.typeValue]; ok {
		return extractor(field)
	}
	return nil
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
func (sinkHttp *SinkHttp) batchLoop() {
	for {
		select {
		case <-sinkHttp.batchTicker.C:
			sinkHttp.flush()
		case <-sinkHttp.batchChan:
			sinkHttp.flush()
			return
		}
	}
}
func (sinkHttp *SinkHttp) circuitAllow() bool {
	if !sinkHttp.circuitEnabled {
		return true
	}
	state := sinkHttp.circuitState.Load()
	switch state {
	case circuitStateClosed:
		return true
	case circuitStateOpen:
		lastFailure := sinkHttp.circuitLastFailure.Load()
		if time.Now().UnixNano()-lastFailure > sinkHttp.circuitTimeout.Nanoseconds() {
			sinkHttp.circuitMutex.Lock()
			if sinkHttp.circuitState.Load() == circuitStateOpen {
				sinkHttp.circuitState.Store(circuitStateHalfOpen)
			}
			sinkHttp.circuitMutex.Unlock()
			return true
		}
		return false
	case circuitStateHalfOpen:
		return false
	default:
		return true
	}
}
func (sinkHttp *SinkHttp) circuitRecord(success bool) {
	if !sinkHttp.circuitEnabled {
		return
	}
	state := sinkHttp.circuitState.Load()
	switch state {
	case circuitStateClosed:
		if !success {
			failures := sinkHttp.circuitFailures.Add(1)
			sinkHttp.circuitLastFailure.Store(time.Now().UnixNano())
			if int(failures) >= sinkHttp.circuitMaxFailures {
				sinkHttp.circuitMutex.Lock()
				if sinkHttp.circuitState.Load() == circuitStateClosed {
					sinkHttp.circuitState.Store(circuitStateOpen)
					sinkHttp.circuitFailures.Store(0)
				}
				sinkHttp.circuitMutex.Unlock()
			}
		} else {
			sinkHttp.circuitMutex.Lock()
			if sinkHttp.circuitState.Load() == circuitStateClosed {
				sinkHttp.circuitFailures.Store(0)
			}
			sinkHttp.circuitMutex.Unlock()
		}
	case circuitStateHalfOpen:
		sinkHttp.circuitMutex.Lock()
		defer sinkHttp.circuitMutex.Unlock()
		if success {
			sinkHttp.circuitState.Store(circuitStateClosed)
			sinkHttp.circuitFailures.Store(0)
		} else {
			sinkHttp.circuitState.Store(circuitStateOpen)
			sinkHttp.circuitLastFailure.Store(time.Now().UnixNano())
		}
	case circuitStateOpen:
	}
}
func (sinkHttp *SinkHttp) cleanupDedupCache() {
	ticker := time.NewTicker(sinkHttp.dedupWindow)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			sinkHttp.dedupCache.Range(func(key, value any) bool {
				if now.Sub(value.(time.Time)) > sinkHttp.dedupWindow {
					sinkHttp.dedupCache.Delete(key)
				}
				return true
			})
		case <-sinkHttp.dedupStopChan:
			return
		}
	}
}
func (sinkHttp *SinkHttp) flush() error {
	sinkHttp.batchMutex.Lock()
	if len(sinkHttp.batchBuffer) == 0 {
		sinkHttp.batchMutex.Unlock()
		return nil
	}
	batch := make([][]byte, len(sinkHttp.batchBuffer))
	copy(batch, sinkHttp.batchBuffer)
	sinkHttp.batchBuffer = sinkHttp.batchBuffer[:0]
	sinkHttp.batchMutex.Unlock()
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
	_, err := sinkHttp.sendWithRetry(body)
	return err
}
func (sinkHttp *SinkHttp) isDuplicate(fields []Field) bool {
	if sinkHttp.dedupWindow <= 0 {
		return false
	}
	hash := sinkHttp.hashFields(fields)
	if lastSeen, ok := sinkHttp.dedupCache.Load(hash); ok {
		if time.Since(lastSeen.(time.Time)) < sinkHttp.dedupWindow {
			return true
		}
	}
	sinkHttp.dedupCache.Store(hash, time.Now())
	return false
}
func (sinkHttp *SinkHttp) hashFields(fields []Field) uint64 {
	hash := fnv.New64a()
	for _, f := range fields {
		hash.Write([]byte(f.nameKey))
		hash.Write([]byte{0})
		hash.Write([]byte(fmt.Sprintf("%v", f.valueString)))
		hash.Write([]byte{0})
	}
	return hash.Sum64()
}
func (sinkHttp *SinkHttp) send(body []byte) error {
	req, err := http.NewRequest(sinkHttp.method, sinkHttp.endPoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range sinkHttp.headers {
		req.Header.Set(k, v)
	}
	resp, err := sinkHttp.client.Do(req)
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
func (sinkHttp *SinkHttp) sendWithRetry(body []byte) (int, error) {
	var lastErr error
	for i := 0; i <= sinkHttp.retryMax; i++ {
		if !sinkHttp.circuitAllow() {
			return 0, fmt.Errorf("circuit breaker is open")
		}
		err := sinkHttp.send(body)
		sinkHttp.circuitRecord(err == nil)
		if err == nil {
			return len(body), nil
		}
		lastErr = err
		if i == sinkHttp.retryMax {
			break
		}
		var sleepDuration time.Duration
		if rateErr, ok := err.(*rateLimitError); ok {
			sleepDuration = rateErr.retryAfter
		} else {
			sleepDuration = sinkHttp.retryBackoff * time.Duration(1<<i)
		}
		time.Sleep(sleepDuration)
	}
	return 0, fmt.Errorf("failed after %d retries: %w", sinkHttp.retryMax, lastErr)
}
func (sinkHttp *SinkHttp) shouldSample() bool {
	if sinkHttp.sampleRate <= 1 {
		return true
	}
	sinkHttp.sampleMutex.Lock()
	defer sinkHttp.sampleMutex.Unlock()
	if sinkHttp.sampleWindow > 0 && time.Since(sinkHttp.sampleLastReset) > sinkHttp.sampleWindow {
		sinkHttp.sampleCounter = 0
		sinkHttp.sampleLastReset = time.Now()
	}
	sinkHttp.sampleCounter++
	return sinkHttp.sampleCounter%sinkHttp.sampleRate == 0
}
func (rateLimitError *rateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", rateLimitError.retryAfter)
}
