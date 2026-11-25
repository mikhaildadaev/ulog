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
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Публичные структуры
type SinkHttp struct {
	batchBuffer                  [][]byte
	batchChan                    chan struct{}
	batchFlushChan               chan struct{}
	batchMerger                  func([][]byte) ([]byte, error)
	batchMutex                   sync.Mutex
	batchSize                    int
	batchTicker                  *time.Ticker
	batchTickerUpdate            chan *time.Ticker
	circuitEnabled               bool
	circuitFailures              atomic.Int32
	circuitHalfOpenProbeInFlight atomic.Bool
	circuitHalfOpenProbeStart    atomic.Int64
	circuitMaxFailures           int
	circuitLastFailure           atomic.Int64
	circuitMutex                 sync.Mutex
	circuitState                 atomic.Int32
	circuitTimeout               time.Duration
	client                       *http.Client
	closed                       bool
	dedupCache                   sync.Map
	dedupCacheCount              atomic.Int64
	dedupCacheMaxSize            int64
	dedupEvictMutex              sync.Mutex
	dedupStopChan                chan struct{}
	dedupWindow                  time.Duration
	endPoint                     string
	filterData                   TypeData
	filterLevel                  TypeLevel
	formatter                    func(attributes writeAttributes, fields []Field) ([]byte, error)
	headers                      map[string]string
	method                       string
	metrics                      *Metrics
	metricsName                  string
	metricsRegistry              *MetricsRegistry
	mutex                        sync.Mutex
	once                         sync.Once
	retryBackoff                 time.Duration
	retryMax                     int
	sampleCounter                atomic.Int64
	sampleRate                   atomic.Int32
	sampleWindow                 time.Duration
	sampleWindowStart            atomic.Int64
	wg                           sync.WaitGroup
}

// Публичные конструкторы
func NewSinkHttp(endPoint string, params ...httpParams) *SinkHttp {
	sinkHttp := &SinkHttp{
		batchChan:          make(chan struct{}),
		batchFlushChan:     make(chan struct{}, 1),
		batchSize:          100,
		batchTicker:        time.NewTicker(5 * time.Second),
		batchTickerUpdate:  make(chan *time.Ticker),
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
		dedupCacheMaxSize: 100000,
		dedupStopChan:     make(chan struct{}),
		endPoint:          endPoint,
		filterData:        defaultType,
		filterLevel:       defaultLevel,
		formatter:         defaultFormatter,
		headers:           make(map[string]string),
		method:            "POST",
		metrics:           defaultMetrics,
		retryBackoff:      time.Second,
		retryMax:          0,
	}
	sinkHttp.circuitState.Store(circuitStateClosed)
	sinkHttp.circuitFailures.Store(0)
	sinkHttp.sampleRate.Store(1)
	sinkHttp.wg.Add(1)
	go func() {
		defer sinkHttp.wg.Done()
		sinkHttp.batchLoop()
	}()
	for _, param := range params {
		param(sinkHttp)
	}
	if sinkHttp.metricsRegistry != nil && sinkHttp.metricsName != "" {
		sinkHttp.metricsRegistry.Register(sinkHttp.metricsName, sinkHttp.metrics)
	}
	if sinkHttp.dedupWindow > 0 {
		sinkHttp.wg.Add(1)
		stopChan := sinkHttp.dedupStopChan
		go func() {
			defer sinkHttp.wg.Done()
			sinkHttp.cleanupDedupCache(stopChan)
		}()
	}
	if sinkHttp.circuitEnabled {
		sinkHttp.wg.Add(1)
		go func() {
			defer sinkHttp.wg.Done()
			sinkHttp.circuitProbeWatchdog()
		}()
	}
	return sinkHttp
}

// Публичные функции
func WithHttpBatch(size int, flushInterval time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.batchMutex.Lock()
		oldTicker := sinkHttp.batchTicker
		newTicker := time.NewTicker(flushInterval)
		sinkHttp.batchSize = size
		sinkHttp.batchTicker = newTicker
		hasBuffered := len(sinkHttp.batchBuffer) > 0
		sinkHttp.batchMutex.Unlock()
		if oldTicker != nil {
			oldTicker.Stop()
		}
		if hasBuffered {
			select {
			case sinkHttp.batchFlushChan <- struct{}{}:
			default:
			}
		}
		select {
		case sinkHttp.batchTickerUpdate <- newTicker:
		case <-sinkHttp.batchChan:
			newTicker.Stop()
		}
	}
}
func WithHttpBatchMerger(merger func(batch [][]byte) ([]byte, error)) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.batchMerger = merger
	}
}
func WithHttpCircuitBreaker(maxFailures int, timeout time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.circuitEnabled = true
		sinkHttp.circuitMaxFailures = maxFailures
		sinkHttp.circuitState.Store(circuitStateClosed)
		sinkHttp.metrics.setCircuitState(circuitStateClosed)
		sinkHttp.circuitTimeout = timeout
		sinkHttp.circuitHalfOpenProbeInFlight.Store(false)
		sinkHttp.circuitHalfOpenProbeStart.Store(0)
	}
}
func WithHttpDedupWindow(window time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.dedupWindow = window
	}
}
func WithHttpDedupMaxSize(size int64) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.dedupCacheMaxSize = size
	}
}
func WithHttpDisabledBatch() httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.batchMutex.Lock()
		oldTicker := sinkHttp.batchTicker
		sinkHttp.batchTicker = nil
		sinkHttp.batchSize = 0
		sinkHttp.batchMutex.Unlock()
		if oldTicker != nil {
			oldTicker.Stop()
		}
		select {
		case sinkHttp.batchTickerUpdate <- nil:
		case <-sinkHttp.batchChan:
		}
	}
}
func WithHttpDisabledCircuit() httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.circuitMutex.Lock()
		defer sinkHttp.circuitMutex.Unlock()
		sinkHttp.circuitEnabled = false
		sinkHttp.circuitState.Store(circuitStateClosed)
		sinkHttp.metrics.setCircuitState(circuitStateClosed)
		sinkHttp.circuitFailures.Store(0)
		sinkHttp.circuitHalfOpenProbeInFlight.Store(false)
		sinkHttp.circuitHalfOpenProbeStart.Store(0)
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
func WithHttpMetrics(metrics *Metrics) httpParams {
	return func(sinkHttp *SinkHttp) {
		if metrics != nil {
			sinkHttp.metrics = metrics
		}
	}
}
func WithHttpMetricsName(metricsRegistry *MetricsRegistry, name string) httpParams {
	return func(sinkHttp *SinkHttp) {
		if name == "" {
			return
		}
		if metricsRegistry == nil {
			metricsRegistry = globalRegistry
		}
		sinkHttp.metricsName = name
		sinkHttp.metricsRegistry = metricsRegistry
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
		sinkHttp.sampleCounter.Store(0)
		sinkHttp.sampleRate.Store(rate)
		if sinkHttp.sampleWindow > 0 {
			sinkHttp.sampleWindowStart.Store(time.Now().UnixNano())
		}
	}
}
func WithHttpSampleWindow(window time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.sampleWindow = window
		if window > 0 {
			sinkHttp.sampleWindowStart.Store(time.Now().UnixNano())
			sinkHttp.sampleCounter.Store(0)
		}
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
		dedupStopChan := sinkHttp.dedupStopChan
		sinkHttp.dedupStopChan = nil
		sinkHttp.mutex.Unlock()
		if dedupStopChan != nil {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(DefaultWriterErr, "ulog: panic closing dedupStopChan: %v\n", r)
				}
			}()
			close(dedupStopChan)
		}
		sinkHttp.batchMutex.Lock()
		ticker := sinkHttp.batchTicker
		sinkHttp.batchTicker = nil
		sinkHttp.batchMutex.Unlock()
		if ticker != nil {
			ticker.Stop()
		}
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(DefaultWriterErr, "ulog: panic closing batchChan: %v\n", r)
			}
		}()
		close(sinkHttp.batchChan)
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
		if sinkHttp.metricsRegistry != nil && sinkHttp.metricsName != "" {
			sinkHttp.metricsRegistry.Unregister(sinkHttp.metricsName)
		}
	})
	return err
}
func (sinkHttp *SinkHttp) Write(p []byte) (n int, err error) {
	n, err = sinkHttp.sendWithRetry(p)
	if err == nil {
		sinkHttp.metrics.incSent()
		sinkHttp.metrics.addSentBytes(int64(len(p)))
	}
	return n, err
}
func (sinkHttp *SinkHttp) WriteWithAttributes(attributes writeAttributes, fields []Field) (n int, err error) {
	sinkHttp.mutex.Lock()
	if sinkHttp.closed {
		sinkHttp.mutex.Unlock()
		return 0, fmt.Errorf("ulog: sink is closed")
	}
	sinkHttp.mutex.Unlock()
	if attributes.typeLevel < sinkHttp.filterLevel {
		return 0, nil
	}
	if attributes.typeData != sinkHttp.filterData {
		sinkHttp.metrics.incFilteredData()
		return 0, nil
	}
	if attributes.typeLevel != LevelError && attributes.typeLevel != LevelFatal {
		if !sinkHttp.shouldSample() {
			sinkHttp.metrics.incSampledOut()
			return 0, nil
		}
		sinkHttp.metrics.incSampled()
		if sinkHttp.isDuplicate(attributes.typeData, attributes.typeLevel, fields) {
			sinkHttp.metrics.incDedupHits()
			return 0, nil
		}
		sinkHttp.metrics.incDedupMisses()
	}
	body, err := sinkHttp.formatter(attributes, fields)
	if err != nil {
		return 0, fmt.Errorf("formatter error: %w", err)
	}
	if sinkHttp.batchSize > 0 {
		sinkHttp.batchMutex.Lock()
		const maxBatchBufferSize = 10000
		if len(sinkHttp.batchBuffer) >= maxBatchBufferSize {
			fmt.Fprintf(DefaultWriterErr, "ulog: batch buffer full, dropping oldest entry\n")
			sinkHttp.batchBuffer = sinkHttp.batchBuffer[1:]
		}
		sinkHttp.batchBuffer = append(sinkHttp.batchBuffer, body)
		needFlush := len(sinkHttp.batchBuffer) >= sinkHttp.batchSize
		sinkHttp.metrics.setBatchBufferSize(int64(len(sinkHttp.batchBuffer)))
		sinkHttp.batchMutex.Unlock()
		if needFlush {
			select {
			case sinkHttp.batchFlushChan <- struct{}{}:
			default:
			}
		}
		return len(body), nil
	}
	n, err = sinkHttp.sendWithRetry(body)
	if err == nil {
		sinkHttp.metrics.incSent()
		sinkHttp.metrics.addSentBytes(int64(len(body)))
	}
	return n, err

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
func defaultFormatter(attributes writeAttributes, fields []Field) ([]byte, error) {
	buf := &bytes.Buffer{}
	formatJson(buf, attributes, fields)
	return buf.Bytes(), nil
}

// Приватные методы
func (sinkHttp *SinkHttp) batchLoop() {
	sinkHttp.batchMutex.Lock()
	ticker := sinkHttp.batchTicker
	sinkHttp.batchMutex.Unlock()
	for {
		select {
		case <-sinkHttp.batchChan:
			sinkHttp.flushLogged()
			return
		default:
		}
		var tickerC <-chan time.Time
		if ticker != nil {
			tickerC = ticker.C
		}
		select {
		case <-tickerC:
			sinkHttp.flushLogged()
		case <-sinkHttp.batchFlushChan:
			sinkHttp.flushLogged()
		case newTicker := <-sinkHttp.batchTickerUpdate:
			ticker = newTicker
		case <-sinkHttp.batchChan:
			sinkHttp.flushLogged()
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
		if time.Now().UnixNano()-lastFailure <= sinkHttp.circuitTimeout.Nanoseconds() {
			return false
		}
		sinkHttp.circuitMutex.Lock()
		if sinkHttp.circuitState.Load() != circuitStateOpen {
			sinkHttp.circuitMutex.Unlock()
			return false
		}
		sinkHttp.circuitState.Store(circuitStateHalfOpen)
		sinkHttp.metrics.setCircuitState(circuitStateHalfOpen)
		sinkHttp.circuitHalfOpenProbeInFlight.Store(false)
		sinkHttp.circuitHalfOpenProbeStart.Store(0)
		allowed := sinkHttp.circuitHalfOpenProbeInFlight.CompareAndSwap(false, true)
		if allowed {
			sinkHttp.circuitHalfOpenProbeStart.Store(time.Now().UnixNano())
		}
		sinkHttp.circuitMutex.Unlock()
		return allowed
	case circuitStateHalfOpen:
		sinkHttp.circuitMutex.Lock()
		defer sinkHttp.circuitMutex.Unlock()
		if sinkHttp.circuitState.Load() != circuitStateHalfOpen {
			return false
		}
		if sinkHttp.circuitHalfOpenProbeInFlight.Load() {
			probeStart := sinkHttp.circuitHalfOpenProbeStart.Load()
			if probeStart > 0 && time.Now().UnixNano()-probeStart > sinkHttp.circuitTimeout.Nanoseconds() {
				sinkHttp.circuitHalfOpenProbeInFlight.Store(false)
				fmt.Fprintf(DefaultWriterErr, "ulog: circuit half-open probe timed out, resetting\n")
			} else {
				return false
			}
		}
		if sinkHttp.circuitHalfOpenProbeInFlight.CompareAndSwap(false, true) {
			sinkHttp.circuitHalfOpenProbeStart.Store(time.Now().UnixNano())
			return true
		}
		return false
	default:
		return true
	}
}
func (sinkHttp *SinkHttp) circuitProbeWatchdog() {
	interval := sinkHttp.circuitTimeout / 2
	if interval < 10*time.Millisecond {
		interval = 10 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !sinkHttp.circuitEnabled {
				continue
			}
			if sinkHttp.circuitState.Load() != circuitStateHalfOpen {
				continue
			}
			if !sinkHttp.circuitHalfOpenProbeInFlight.Load() {
				continue
			}
			probeStart := sinkHttp.circuitHalfOpenProbeStart.Load()
			if probeStart == 0 {
				continue
			}
			if time.Now().UnixNano()-probeStart <= sinkHttp.circuitTimeout.Nanoseconds() {
				continue
			}
			sinkHttp.circuitMutex.Lock()
			if sinkHttp.circuitState.Load() == circuitStateHalfOpen {
				sinkHttp.circuitHalfOpenProbeInFlight.Store(false)
				fmt.Fprintf(DefaultWriterErr, "ulog: circuit half-open probe timed out, resetting\n")
			}
			sinkHttp.circuitMutex.Unlock()
		case <-sinkHttp.batchChan:
			return
		}
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
					sinkHttp.metrics.incCircuitOpens()
					sinkHttp.metrics.setCircuitState(circuitStateOpen)
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
		if sinkHttp.circuitState.Load() != circuitStateHalfOpen {
			return
		}
		sinkHttp.circuitHalfOpenProbeInFlight.Store(false)
		sinkHttp.circuitHalfOpenProbeStart.Store(0)
		if success {
			sinkHttp.circuitState.Store(circuitStateClosed)
			sinkHttp.circuitFailures.Store(0)
			sinkHttp.metrics.incCircuitCloses()
			sinkHttp.metrics.setCircuitState(circuitStateClosed)
		} else {
			sinkHttp.circuitState.Store(circuitStateOpen)
			sinkHttp.circuitLastFailure.Store(time.Now().UnixNano())
			sinkHttp.metrics.incCircuitOpens()
			sinkHttp.metrics.setCircuitState(circuitStateOpen)
		}
	case circuitStateOpen:
	}
}
func (sinkHttp *SinkHttp) cleanupDedupCache(stopChan chan struct{}) {
	ttlInterval := sinkHttp.dedupWindow / 10
	if ttlInterval < 100*time.Millisecond {
		ttlInterval = 100 * time.Millisecond
	}
	sizeInterval := 1 * time.Second
	if sizeInterval > ttlInterval {
		sizeInterval = ttlInterval
	}
	ttlTicker := time.NewTicker(ttlInterval)
	defer ttlTicker.Stop()
	sizeTicker := time.NewTicker(sizeInterval)
	defer sizeTicker.Stop()
	for {
		select {
		case <-ttlTicker.C:
			sinkHttp.evictDedupTTL()
		case <-sizeTicker.C:
			sinkHttp.evictDedupSize()
		case <-stopChan:
			return
		}
	}
}
func (sinkHttp *SinkHttp) evictDedupTTL() {
	sinkHttp.dedupEvictMutex.Lock()
	defer sinkHttp.dedupEvictMutex.Unlock()
	now := time.Now()
	sinkHttp.dedupCache.Range(func(key, value any) bool {
		if now.Sub(value.(time.Time)) > sinkHttp.dedupWindow {
			if _, loaded := sinkHttp.dedupCache.LoadAndDelete(key); loaded {
				sinkHttp.dedupCacheCount.Add(-1)
				sinkHttp.metrics.incDedupEvictions()
				sinkHttp.metrics.setDedupCacheSize(sinkHttp.dedupCacheCount.Load())
			}
		}
		return true
	})
}
func (sinkHttp *SinkHttp) evictDedupSize() {
	sinkHttp.dedupEvictMutex.Lock()
	defer sinkHttp.dedupEvictMutex.Unlock()
	if sinkHttp.dedupCacheMaxSize <= 0 {
		return
	}
	count := sinkHttp.dedupCacheCount.Load()
	if count <= sinkHttp.dedupCacheMaxSize {
		return
	}
	type entry struct {
		key      any
		lastSeen time.Time
	}
	entries := make([]entry, 0, count)
	sinkHttp.dedupCache.Range(func(key, value any) bool {
		entries = append(entries, entry{key: key, lastSeen: value.(time.Time)})
		return true
	})
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].lastSeen.Before(entries[j].lastSeen)
	})
	remove := int64(len(entries)) - sinkHttp.dedupCacheMaxSize
	if remove <= 0 {
		return
	}
	for i := int64(0); i < remove; i++ {
		if _, loaded := sinkHttp.dedupCache.LoadAndDelete(entries[i].key); loaded {
			sinkHttp.dedupCacheCount.Add(-1)
			sinkHttp.metrics.incDedupEvictions()
			sinkHttp.metrics.setDedupCacheSize(sinkHttp.dedupCacheCount.Load())
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
	sinkHttp.metrics.setBatchBufferSize(0)
	sinkHttp.batchMutex.Unlock()
	sinkHttp.metrics.incBatchFlushes()
	sinkHttp.metrics.addBatchSizeTotal(int64(len(batch)))
	atomicMaxInt64(&sinkHttp.metrics.batchSizeMax, int64(len(batch)))
	atomicMinInt64(&sinkHttp.metrics.batchSizeMin, int64(len(batch)))
	var body []byte
	if sinkHttp.batchMerger != nil {
		merged, err := sinkHttp.batchMerger(batch)
		if err != nil {
			sinkHttp.metrics.incBatchMergeFailed()
			return err
		}
		body = merged
	} else if len(batch) == 1 {
		body = bytes.TrimRight(batch[0], "\n")
	} else {
		parts := make([][]byte, len(batch))
		for i, b := range batch {
			parts[i] = bytes.TrimRight(b, "\n")
		}
		body = bytes.Join(parts, []byte{'\n'})
	}
	if _, err := sinkHttp.sendWithRetry(body); err != nil {
		return err
	}
	sinkHttp.metrics.incSentN(int64(len(batch)))
	sinkHttp.metrics.addSentBytes(int64(len(body)))
	return nil
}
func (sinkHttp *SinkHttp) flushLogged() {
	if err := sinkHttp.flush(); err != nil {
		fmt.Fprintf(DefaultWriterErr, "ulog: SinkHttp flush failed: %v\n", err)
	}
}
func (sinkHttp *SinkHttp) isDuplicate(typeData TypeData, typeLevel TypeLevel, fields []Field) bool {
	if sinkHttp.dedupWindow <= 0 || len(fields) == 0 {
		return false
	}
	hash := sinkHttp.hashFields(typeData, typeLevel, fields)
	if lastSeen, ok := sinkHttp.dedupCache.Load(hash); ok {
		if time.Since(lastSeen.(time.Time)) < sinkHttp.dedupWindow {
			return true
		}
	}
	if _, loaded := sinkHttp.dedupCache.LoadOrStore(hash, time.Now()); !loaded {
		sinkHttp.dedupCacheCount.Add(1)
		sinkHttp.metrics.setDedupCacheSize(sinkHttp.dedupCacheCount.Load())
	}
	return false
}
func (sinkHttp *SinkHttp) hashFields(typeData TypeData, typeLevel TypeLevel, fields []Field) uint64 {
	sorted := make([]Field, len(fields))
	copy(sorted, fields)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].nameKey != sorted[j].nameKey {
			return sorted[i].nameKey < sorted[j].nameKey
		}
		return sorted[i].typeValue < sorted[j].typeValue
	})
	hash := fnv.New64a()
	hash.Write([]byte{byte(typeData), byte(typeLevel)})
	var buf [32]byte
	for _, v := range sorted {
		hash.Write([]byte(v.nameKey))
		hash.Write([]byte{0})
		switch v.typeValue {
		case FieldString:
			hash.Write([]byte(v.valueString))
		case FieldInt:
			hash.Write(strconv.AppendInt(buf[:0], int64(v.valueInt), 10))
		case FieldInt64:
			hash.Write(strconv.AppendInt(buf[:0], v.valueInt64, 10))
		case FieldFloat64:
			hash.Write(strconv.AppendFloat(buf[:0], v.valueFloat64, 'f', -1, 64))
		case FieldBool:
			if v.valueBool {
				hash.Write([]byte("1"))
			} else {
				hash.Write([]byte("0"))
			}
		case FieldDuration:
			hash.Write(strconv.AppendInt(buf[:0], int64(v.valueDuration), 10))
		case FieldTime:
			hash.Write(strconv.AppendInt(buf[:0], v.valueTime.UnixNano(), 10))
		case FieldStrings:
			for _, s := range v.valueStrings {
				hash.Write([]byte(s))
				hash.Write([]byte{0})
			}
		case FieldInts:
			for _, i := range v.valueInts {
				hash.Write(strconv.AppendInt(buf[:0], int64(i), 10))
				hash.Write([]byte{0})
			}
		case FieldInts64:
			for _, i := range v.valueInts64 {
				hash.Write(strconv.AppendInt(buf[:0], i, 10))
				hash.Write([]byte{0})
			}
		case FieldFloats64:
			for _, f := range v.valueFloats64 {
				hash.Write(strconv.AppendFloat(buf[:0], f, 'f', -1, 64))
				hash.Write([]byte{0})
			}
		case FieldBools:
			for _, b := range v.valueBools {
				if b {
					hash.Write([]byte("1"))
				} else {
					hash.Write([]byte("0"))
				}
				hash.Write([]byte{0})
			}
		case FieldDurations:
			for _, d := range v.valueDurations {
				hash.Write(strconv.AppendInt(buf[:0], int64(d), 10))
				hash.Write([]byte{0})
			}
		case FieldTimes:
			for _, t := range v.valueTimes {
				hash.Write(strconv.AppendInt(buf[:0], t.UnixNano(), 10))
				hash.Write([]byte{0})
			}
		}
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
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
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
func (sinkHttp *SinkHttp) sendWithRetry(body []byte) (n int, err error) {
	var lastErr error
	for i := 0; i <= sinkHttp.retryMax; i++ {
		if !sinkHttp.circuitAllow() {
			sinkHttp.metrics.incCircuitBlocked()
			return 0, fmt.Errorf("circuit breaker is open")
		}
		var sendErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					sendErr = fmt.Errorf("panic in send: %v", r)
				}
			}()
			sendErr = sinkHttp.send(body)
		}()
		sinkHttp.circuitRecord(sendErr == nil)
		if sendErr == nil {
			return len(body), nil
		}
		lastErr = sendErr
		sinkHttp.metrics.incFailed()
		if _, ok := sendErr.(*rateLimitError); ok {
			sinkHttp.metrics.incRateLimited()
		}
		if i == sinkHttp.retryMax {
			break
		}

		sinkHttp.metrics.incRetried()
		var sleepDuration time.Duration
		if rateErr, ok := sendErr.(*rateLimitError); ok {
			sleepDuration = rateErr.retryAfter
		} else {
			shift := i
			if shift > 30 {
				shift = 30
			}
			sleepDuration = sinkHttp.retryBackoff * time.Duration(1<<shift)
		}
		const maxBackoffSleep = 30 * time.Second
		if sleepDuration > maxBackoffSleep {
			sleepDuration = maxBackoffSleep
		}
		time.Sleep(sleepDuration)
	}
	return 0, fmt.Errorf("failed after %d retries: %w", sinkHttp.retryMax, lastErr)
}
func (sinkHttp *SinkHttp) shouldSample() bool {
	rate := sinkHttp.sampleRate.Load()
	if rate <= 1 {
		return true
	}
	if sinkHttp.sampleWindow > 0 {
		now := time.Now().UnixNano()
		start := sinkHttp.sampleWindowStart.Load()
		if start == 0 || now-start >= sinkHttp.sampleWindow.Nanoseconds() {
			if sinkHttp.sampleWindowStart.CompareAndSwap(start, now) {
				sinkHttp.sampleCounter.Store(0)
			}
		}
	}
	counter := sinkHttp.sampleCounter.Add(1)
	return counter%int64(rate) == 0
}
func (rateLimitError *rateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", rateLimitError.retryAfter)
}
