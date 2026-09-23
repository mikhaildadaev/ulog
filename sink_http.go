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
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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
	sampleCounter                int32
	sampleLastReset              time.Time
	sampleMutex                  sync.Mutex
	sampleRate                   int32
	sampleWindow                 time.Duration
	wg                           sync.WaitGroup
}

// Публичные конструкторы
func NewSinkHttp(endPoint string, params ...httpParams) *SinkHttp {
	sinkHttp := &SinkHttp{
		batchChan:          make(chan struct{}),
		batchFlushChan:     make(chan struct{}, 1),
		batchSize:          100,
		batchTicker:        time.NewTicker(5 * time.Second),
		batchTickerUpdate:  make(chan *time.Ticker, 8),
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
	sinkHttp.wg.Add(1)
	go func() {
		defer sinkHttp.wg.Done()
		sinkHttp.batchLoop()
	}()
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
		sinkHttp.batchMutex.Unlock()
		if oldTicker != nil {
			oldTicker.Stop()
		}
		select {
		case sinkHttp.batchTickerUpdate <- newTicker:
		default:
		}
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
		default:
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
func getField(field Field) any {
	if extractor, ok := fieldExtractor[field.typeValue]; ok {
		return extractor(field)
	}
	return nil
}
func getLogData(fields []Field) string {
	for _, field := range fields {
		if field.nameKey == "message" {
			return field.valueString
		}
	}
	return ""
}
func getMetricData(fields []Field) (name string, value float64) {
	for _, field := range fields {
		switch field.nameKey {
		case "name":
			name = field.valueString
		case "value":
			switch field.typeValue {
			case FieldFloat64:
				value = field.valueFloat64
			case FieldInt64:
				value = float64(field.valueInt64)
			case FieldInt:
				value = float64(field.valueInt)
			}
		}
	}
	if name == "" {
		name = "unnamed-metric"
	}
	return name, value
}
func getKafkaAttributes(fields []Field) map[string]any {
	valueData := make(map[string]any, len(fields))
	for _, field := range fields {
		v := getField(field)
		if field.typeValue == FieldString {
			switch field.nameKey {
			case "trace_id":
				if normalized, err := normalizeTraceID(field.valueString); err == nil {
					v = normalized
				}
			case "span_id":
				if normalized, err := normalizeSpanID(field.valueString); err == nil {
					v = normalized
				}
			}
		}
		valueData[field.nameKey] = v
	}
	return valueData
}
func getKafkaKey(fields []Field) string {
	priorities := []string{"trace_id", "node_id", "user_id", "request_id"}
	for _, k := range priorities {
		for _, field := range fields {
			if field.nameKey != k || field.typeValue != FieldString {
				continue
			}
			if k == "trace_id" {
				if normalized, err := normalizeTraceID(field.valueString); err == nil {
					return normalized
				}
				continue
			}
			return field.valueString
		}
	}
	return ""
}
func getOpenTelemetryAttributes(fields []Field, skipKeys ...string) []OTLPAttribute {
	attrs := make([]OTLPAttribute, 0, len(fields))
	for _, f := range fields {
		skip := false
		for _, k := range skipKeys {
			if f.nameKey == k {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		switch f.typeValue {
		case FieldString:
			v := f.valueString
			switch f.nameKey {
			case "trace_id":
				normalized, err := normalizeTraceID(v)
				if err != nil {
					fmt.Fprintf(DefaultWriterErr, "ulog: skipping invalid trace_id: %v\n", err)
					continue
				}
				v = normalized
			case "span_id":
				normalized, err := normalizeSpanID(v)
				if err != nil {
					fmt.Fprintf(DefaultWriterErr, "ulog: skipping invalid span_id: %v\n", err)
					continue
				}
				v = normalized
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{StringValue: v},
			})
		case FieldInt:
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{IntValue: fmt.Sprintf("%d", f.valueInt)},
			})
		case FieldInt64:
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{IntValue: fmt.Sprintf("%d", f.valueInt64)},
			})
		case FieldFloat64:
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{DoubleValue: f.valueFloat64},
			})
		case FieldBool:
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{BoolValue: f.valueBool},
			})
		case FieldDuration:
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{StringValue: f.valueDuration.String()},
			})
		case FieldTime:
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{StringValue: f.valueTime.Format(time.RFC3339Nano)},
			})
		case FieldStrings:
			arr := make([]OTLPAttrValue, len(f.valueStrings))
			for i, v := range f.valueStrings {
				arr[i] = OTLPAttrValue{StringValue: v}
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{ArrayValue: arr},
			})
		case FieldInts:
			arr := make([]OTLPAttrValue, len(f.valueInts))
			for i, v := range f.valueInts {
				arr[i] = OTLPAttrValue{IntValue: fmt.Sprintf("%d", v)}
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{ArrayValue: arr},
			})
		case FieldInts64:
			arr := make([]OTLPAttrValue, len(f.valueInts64))
			for i, v := range f.valueInts64 {
				arr[i] = OTLPAttrValue{IntValue: fmt.Sprintf("%d", v)}
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{ArrayValue: arr},
			})
		case FieldFloats64:
			arr := make([]OTLPAttrValue, len(f.valueFloats64))
			for i, v := range f.valueFloats64 {
				arr[i] = OTLPAttrValue{DoubleValue: v}
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{ArrayValue: arr},
			})
		case FieldBools:
			arr := make([]OTLPAttrValue, len(f.valueBools))
			for i, v := range f.valueBools {
				arr[i] = OTLPAttrValue{BoolValue: v}
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{ArrayValue: arr},
			})
		case FieldDurations:
			arr := make([]OTLPAttrValue, len(f.valueDurations))
			for i, v := range f.valueDurations {
				arr[i] = OTLPAttrValue{StringValue: v.String()}
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{ArrayValue: arr},
			})
		case FieldTimes:
			arr := make([]OTLPAttrValue, len(f.valueTimes))
			for i, v := range f.valueTimes {
				arr[i] = OTLPAttrValue{StringValue: v.Format(time.RFC3339Nano)}
			}
			attrs = append(attrs, OTLPAttribute{
				Key:   f.nameKey,
				Value: OTLPAttrValue{ArrayValue: arr},
			})
		}
	}
	return attrs
}
func getTraceData(fields []Field) (name, traceID, spanID string, duration int64, err error) {
	var (
		rawTraceID string
		rawSpanID  string
		rawName    string
		rawDur     int64
		hasDur     bool
	)
	for _, f := range fields {
		switch f.nameKey {
		case "trace_id":
			if f.typeValue == FieldString {
				rawTraceID = f.valueString
			}
		case "span_id":
			if f.typeValue == FieldString {
				rawSpanID = f.valueString
			}
		case "name":
			if f.typeValue == FieldString {
				rawName = f.valueString
			}
		case "duration":
			var ms int64
			switch f.typeValue {
			case FieldInt:
				ms = int64(f.valueInt)
			case FieldInt64:
				ms = f.valueInt64
			case FieldDuration:
				ms = f.valueDuration.Milliseconds()
			case FieldString:
				d, parseErr := time.ParseDuration(f.valueString)
				if parseErr != nil {
					return "", "", "", 0, fmt.Errorf("invalid duration string: %w", parseErr)
				}
				ms = d.Milliseconds()
			}
			rawDur = ms
			hasDur = true
		}
	}
	if rawTraceID == "" {
		return "", "", "", 0, fmt.Errorf("trace_id is required")
	}
	if traceID, err = normalizeTraceID(rawTraceID); err != nil {
		return "", "", "", 0, err
	}
	if rawSpanID == "" {
		return "", "", "", 0, fmt.Errorf("span_id is required")
	}
	if spanID, err = normalizeSpanID(rawSpanID); err != nil {
		return "", "", "", 0, err
	}
	name = rawName
	if name == "" {
		name = "unnamed-trace"
	}
	switch {
	case hasDur && rawDur > 0:
		duration = rawDur
	case hasDur && rawDur <= 0:
		return "", "", "", 0, fmt.Errorf("duration must be positive, got %d", rawDur)
	default:
		duration = 1
	}
	return name, traceID, spanID, duration, nil
}
func normalizeTraceID(value string) (string, error) {
	v := strings.ToLower(strings.ReplaceAll(value, "-", ""))
	if len(v) != 32 {
		return "", fmt.Errorf("invalid trace_id length: got %d, want 32 (input: %q)", len(v), value)
	}
	if _, err := hex.DecodeString(v); err != nil {
		return "", fmt.Errorf("invalid trace_id hex: %w (input: %q)", err, value)
	}
	return v, nil
}
func normalizeSpanID(value string) (string, error) {
	v := strings.ToLower(strings.ReplaceAll(value, "-", ""))
	if len(v) != 16 {
		return "", fmt.Errorf("invalid span_id length: got %d, want 16 (input: %q)", len(v), value)
	}
	if _, err := hex.DecodeString(v); err != nil {
		return "", fmt.Errorf("invalid span_id hex: %w (input: %q)", err, value)
	}
	return v, nil
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
		allowed := sinkHttp.circuitHalfOpenProbeInFlight.CompareAndSwap(false, true)
		sinkHttp.circuitMutex.Unlock()
		return allowed
	case circuitStateHalfOpen:
		sinkHttp.circuitMutex.Lock()
		if sinkHttp.circuitState.Load() != circuitStateHalfOpen {
			sinkHttp.circuitMutex.Unlock()
			return false
		}
		allowed := sinkHttp.circuitHalfOpenProbeInFlight.CompareAndSwap(false, true)
		sinkHttp.circuitMutex.Unlock()
		return allowed
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
		if sinkHttp.circuitState.Load() != circuitStateHalfOpen {
			return
		}
		sinkHttp.circuitHalfOpenProbeInFlight.Store(false)
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
	interval := sinkHttp.dedupWindow / 10
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sinkHttp.evictDedupCache()
		case <-sinkHttp.dedupStopChan:
			return
		}
	}
}
func (sinkHttp *SinkHttp) evictDedupCache() {
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
	if sinkHttp.dedupCacheMaxSize > 0 && sinkHttp.dedupCacheCount.Load() >= sinkHttp.dedupCacheMaxSize {
		sinkHttp.evictDedupCache()
	}
	if _, loaded := sinkHttp.dedupCache.LoadOrStore(hash, time.Now()); !loaded {
		sinkHttp.dedupCacheCount.Add(1)
	}
	return false
}
func (sinkHttp *SinkHttp) hashFields(fields []Field) uint64 {
	hash := fnv.New64a()
	for _, f := range fields {
		hash.Write([]byte(f.nameKey))
		hash.Write([]byte{0})
		fmt.Fprintf(hash, "%v", getField(f))
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
	sinkHttp.sampleMutex.Lock()
	defer sinkHttp.sampleMutex.Unlock()
	if sinkHttp.sampleWindow > 0 && time.Since(sinkHttp.sampleLastReset) > sinkHttp.sampleWindow {
		sinkHttp.sampleCounter = 0
		sinkHttp.sampleLastReset = time.Now()
	}
	sinkHttp.sampleCounter++
	if sinkHttp.sampleCounter <= 0 {
		sinkHttp.sampleCounter = 1
	}
	return sinkHttp.sampleCounter%sinkHttp.sampleRate == 0
}
func (rateLimitError *rateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", rateLimitError.retryAfter)
}
