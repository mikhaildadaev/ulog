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

// Публичные константы
const (
	TraceFlagsSampled       uint32 = 1 << 0 // W3C: sampled
	TraceFlagsRandomTraceID uint32 = 1 << 1 // W3C: random-trace-id
)

// Публичные структуры
type SinkHttp struct {
	batchBuffer                  [][]byte
	batchChan                    chan struct{}
	batchFlushChan               chan struct{}
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
	mutex                        sync.Mutex
	once                         sync.Once
	retryBackoff                 time.Duration
	retryMax                     int
	sampleCounter                atomic.Int64
	sampleRate                   int32
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
		filterData:        TypeData(defaultType),
		filterLevel:       LevelError,
		formatter:         defaultformatter,
		headers:           make(map[string]string),
		method:            "POST",
		retryBackoff:      time.Second,
		retryMax:          0,
	}
	sinkHttp.circuitState.Store(circuitStateClosed)
	sinkHttp.circuitFailures.Store(0)
	sinkHttp.wg.Add(1)
	go func() {
		defer sinkHttp.wg.Done()
		sinkHttp.batchLoop()
	}()
	for _, param := range params {
		param(sinkHttp)
	}
	if sinkHttp.dedupWindow > 0 {
		sinkHttp.wg.Add(1)
		go func() {
			defer sinkHttp.wg.Done()
			sinkHttp.cleanupDedupCache()
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
func WithHttpCircuitBreaker(maxFailures int, timeout time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.circuitEnabled = true
		sinkHttp.circuitMaxFailures = maxFailures
		sinkHttp.circuitState.Store(circuitStateClosed)
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
func WithHttpRetry(maxRetries int, backoff time.Duration) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.retryMax = maxRetries
		sinkHttp.retryBackoff = backoff
	}
}
func WithHttpSampleRate(rate int32) httpParams {
	return func(sinkHttp *SinkHttp) {
		sinkHttp.sampleRate = rate
		sinkHttp.sampleCounter.Store(0)
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
	})
	return err
}
func (sinkHttp *SinkHttp) Write(p []byte) (n int, err error) {
	return sinkHttp.sendWithRetry(p)
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
		const maxBatchBufferSize = 10000
		if len(sinkHttp.batchBuffer) >= maxBatchBufferSize {
			sinkHttp.batchBuffer = sinkHttp.batchBuffer[1:]
		}
		sinkHttp.batchBuffer = append(sinkHttp.batchBuffer, body)
		needFlush := len(sinkHttp.batchBuffer) >= sinkHttp.batchSize
		sinkHttp.batchMutex.Unlock()
		if needFlush {
			select {
			case sinkHttp.batchFlushChan <- struct{}{}:
			default:
			}
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
type otlpAttributes struct {
	record   []OTLPAttribute
	resource []OTLPAttribute
}
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

// Приватные методы
func (sinkHttp *SinkHttp) batchLoop() {
	sinkHttp.batchMutex.Lock()
	ticker := sinkHttp.batchTicker
	sinkHttp.batchMutex.Unlock()
	for {
		select {
		case <-sinkHttp.batchChan:
			sinkHttp.flush()
			return
		default:
		}
		var tickerC <-chan time.Time
		if ticker != nil {
			tickerC = ticker.C
		}
		select {
		case <-tickerC:
			sinkHttp.flush()
		case <-sinkHttp.batchFlushChan:
			sinkHttp.flush()
		case newTicker := <-sinkHttp.batchTickerUpdate:
			ticker = newTicker
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
		if time.Now().UnixNano()-lastFailure <= sinkHttp.circuitTimeout.Nanoseconds() {
			return false
		}
		sinkHttp.circuitMutex.Lock()
		if sinkHttp.circuitState.Load() != circuitStateOpen {
			sinkHttp.circuitMutex.Unlock()
			return false
		}
		sinkHttp.circuitState.Store(circuitStateHalfOpen)
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
		} else {
			sinkHttp.circuitState.Store(circuitStateOpen)
			sinkHttp.circuitLastFailure.Store(time.Now().UnixNano())
		}
	case circuitStateOpen:
	}
}
func (sinkHttp *SinkHttp) cleanupDedupCache() {
	sinkHttp.mutex.Lock()
	dedupStopChan := sinkHttp.dedupStopChan
	sinkHttp.mutex.Unlock()
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
		case <-dedupStopChan:
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
		body = bytes.TrimRight(batch[0], "\n")
	} else {
		parts := make([][]byte, len(batch))
		for i, b := range batch {
			parts[i] = bytes.TrimRight(b, "\n")
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
	if _, loaded := sinkHttp.dedupCache.LoadOrStore(hash, time.Now()); !loaded {
		sinkHttp.dedupCacheCount.Add(1)
	}
	return false
}
func (sinkHttp *SinkHttp) hashFields(fields []Field) uint64 {
	hash := fnv.New64a()
	var buf [32]byte
	for _, f := range fields {
		hash.Write([]byte(f.nameKey))
		hash.Write([]byte{0})
		switch f.typeValue {
		case FieldString:
			hash.Write([]byte(f.valueString))
		case FieldInt:
			hash.Write(strconv.AppendInt(buf[:0], int64(f.valueInt), 10))
		case FieldInt64:
			hash.Write(strconv.AppendInt(buf[:0], f.valueInt64, 10))
		case FieldFloat64:
			hash.Write(strconv.AppendFloat(buf[:0], f.valueFloat64, 'f', -1, 64))
		case FieldBool:
			if f.valueBool {
				hash.Write([]byte("1"))
			} else {
				hash.Write([]byte("0"))
			}
		case FieldDuration:
			hash.Write(strconv.AppendInt(buf[:0], int64(f.valueDuration), 10))
		case FieldTime:
			hash.Write(strconv.AppendInt(buf[:0], f.valueTime.UnixNano(), 10))
		case FieldStrings:
			for _, s := range f.valueStrings {
				hash.Write([]byte(s))
				hash.Write([]byte{0})
			}
		case FieldInts:
			for _, n := range f.valueInts {
				hash.Write(strconv.AppendInt(buf[:0], int64(n), 10))
				hash.Write([]byte{0})
			}
		case FieldInts64:
			for _, n := range f.valueInts64 {
				hash.Write(strconv.AppendInt(buf[:0], n, 10))
				hash.Write([]byte{0})
			}
		case FieldFloats64:
			for _, v := range f.valueFloats64 {
				hash.Write(strconv.AppendFloat(buf[:0], v, 'f', -1, 64))
				hash.Write([]byte{0})
			}
		case FieldBools:
			for _, v := range f.valueBools {
				if v {
					hash.Write([]byte("1"))
				} else {
					hash.Write([]byte("0"))
				}
				hash.Write([]byte{0})
			}
		case FieldDurations:
			for _, d := range f.valueDurations {
				hash.Write(strconv.AppendInt(buf[:0], int64(d), 10))
				hash.Write([]byte{0})
			}
		case FieldTimes:
			for _, t := range f.valueTimes {
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
		if i == sinkHttp.retryMax {
			break
		}
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
		time.Sleep(sleepDuration)
	}
	return 0, fmt.Errorf("failed after %d retries: %w", sinkHttp.retryMax, lastErr)
}
func (sinkHttp *SinkHttp) shouldSample() bool {
	if sinkHttp.sampleRate <= 1 {
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
	return counter%int64(sinkHttp.sampleRate) == 0
}
func (rateLimitError *rateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", rateLimitError.retryAfter)
}
