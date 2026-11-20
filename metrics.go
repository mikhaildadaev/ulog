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
	"bufio"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Публичные структуры
type Metrics struct {
	noCopy
	asyncDroppedTotal atomic.Int64
	failedTotal       atomic.Int64
	rateLimitedTotal  atomic.Int64
	retriedTotal      atomic.Int64
	sentBytes         atomic.Int64
	sentTotal         atomic.Int64
	startTime         atomic.Int64
	// Batch
	batchBufferSize   atomic.Int64
	batchFlushesTotal atomic.Int64
	batchSizeMax      atomic.Int64
	batchSizeMin      atomic.Int64
	batchSizeTotal    atomic.Int64
	// Circuit breaker
	circuitBlockedTotal atomic.Int64
	circuitClosesTotal  atomic.Int64
	circuitOpensTotal   atomic.Int64
	circuitState        atomic.Int32
	// Dedup
	dedupCacheSize      atomic.Int64
	dedupEvictionsTotal atomic.Int64
	dedupHitsTotal      atomic.Int64
	dedupMissesTotal    atomic.Int64
	// Filtering
	filteredDataTotal atomic.Int64
	// Sampling
	sampledOutTotal atomic.Int64
	sampledTotal    atomic.Int64
}
type metricDescription struct {
	name   string
	help   string
	format string
	value  func(metrics *Metrics) int64
}
type MetricsRegistry struct {
	mu      sync.RWMutex
	metrics map[string]*Metrics
}
type MetricsSnapshot struct {
	AsyncDroppedTotal int64
	FailedTotal       int64
	RateLimitedTotal  int64
	RetriedTotal      int64
	SentBytes         int64
	SentTotal         int64
	// Batch
	BatchBufferSize   int64
	BatchFlushesTotal int64
	BatchSizeMax      int64
	BatchSizeMin      int64
	BatchSizeTotal    int64
	// Circuit breaker
	CircuitBlockedTotal int64
	CircuitClosesTotal  int64
	CircuitOpensTotal   int64
	CircuitState        int32
	// Dedup
	DedupCacheSize      int64
	DedupEvictionsTotal int64
	DedupHitsTotal      int64
	DedupMissesTotal    int64
	// Filtering
	FilteredDataTotal int64
	// Sampling
	SampledOutTotal int64
	SampledTotal    int64
}

// Публичные конструкторы
func NewMetrics() *Metrics {
	metrics := &Metrics{}
	metrics.resetStartTime()
	metrics.batchSizeMin.Store(batchSizeMinInit)
	return metrics
}
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: make(map[string]*Metrics),
	}
}

// Публичные функции
func GlobalRegistry() *MetricsRegistry {
	return globalRegistry
}
func MetricsGet() *Metrics {
	return defaultMetrics
}
func MetricsHandler(metrics *Metrics) http.Handler {
	if metrics == nil {
		metrics = MetricsGet()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		metrics.Write(w)
	})
}

// Публичные функции
func MetricsRegistryHandler(metricsRegistry *MetricsRegistry) http.Handler {
	if metricsRegistry == nil {
		metricsRegistry = globalRegistry
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		metricsRegistry.Write(w)
	})
}

// Публичные методы
func (metrics *Metrics) AsyncDroppedTotal() int64   { return metrics.asyncDroppedTotal.Load() }
func (metrics *Metrics) FailedTotal() int64         { return metrics.failedTotal.Load() }
func (metrics *Metrics) RateLimitedTotal() int64    { return metrics.rateLimitedTotal.Load() }
func (metrics *Metrics) RetriedTotal() int64        { return metrics.retriedTotal.Load() }
func (metrics *Metrics) SentBytes() int64           { return metrics.sentBytes.Load() }
func (metrics *Metrics) SentTotal() int64           { return metrics.sentTotal.Load() }
func (metrics *Metrics) StartTime() time.Time       { return time.Unix(0, metrics.startTime.Load()) }
func (metrics *Metrics) BatchBufferSize() int64     { return metrics.batchBufferSize.Load() }
func (metrics *Metrics) BatchFlushesTotal() int64   { return metrics.batchFlushesTotal.Load() }
func (metrics *Metrics) BatchSizeMax() int64        { return metrics.batchSizeMax.Load() }
func (metrics *Metrics) BatchSizeMin() int64        { return metrics.batchSizeMin.Load() }
func (metrics *Metrics) BatchSizeTotal() int64      { return metrics.batchSizeTotal.Load() }
func (metrics *Metrics) CircuitBlockedTotal() int64 { return metrics.circuitBlockedTotal.Load() }
func (metrics *Metrics) CircuitClosesTotal() int64  { return metrics.circuitClosesTotal.Load() }
func (metrics *Metrics) CircuitOpensTotal() int64   { return metrics.circuitOpensTotal.Load() }
func (metrics *Metrics) CircuitState() int32        { return metrics.circuitState.Load() }
func (metrics *Metrics) DedupCacheSize() int64      { return metrics.dedupCacheSize.Load() }
func (metrics *Metrics) DedupEvictionsTotal() int64 { return metrics.dedupEvictionsTotal.Load() }
func (metrics *Metrics) DedupHitsTotal() int64      { return metrics.dedupHitsTotal.Load() }
func (metrics *Metrics) DedupMissesTotal() int64    { return metrics.dedupMissesTotal.Load() }
func (metrics *Metrics) FilteredDataTotal() int64   { return metrics.filteredDataTotal.Load() }
func (metrics *Metrics) SampledOutTotal() int64     { return metrics.sampledOutTotal.Load() }
func (metrics *Metrics) SampledTotal() int64        { return metrics.sampledTotal.Load() }
func (metrics *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		AsyncDroppedTotal: metrics.asyncDroppedTotal.Load(),
		FailedTotal:       metrics.failedTotal.Load(),
		RateLimitedTotal:  metrics.rateLimitedTotal.Load(),
		RetriedTotal:      metrics.retriedTotal.Load(),
		SentBytes:         metrics.sentBytes.Load(),
		SentTotal:         metrics.sentTotal.Load(),
		// Batch
		BatchBufferSize:   metrics.batchBufferSize.Load(),
		BatchFlushesTotal: metrics.batchFlushesTotal.Load(),
		BatchSizeMax:      metrics.batchSizeMax.Load(),
		BatchSizeMin:      metrics.batchSizeMin.Load(),
		BatchSizeTotal:    metrics.batchSizeTotal.Load(),
		// Circuit breaker
		CircuitBlockedTotal: metrics.circuitBlockedTotal.Load(),
		CircuitClosesTotal:  metrics.circuitClosesTotal.Load(),
		CircuitOpensTotal:   metrics.circuitOpensTotal.Load(),
		CircuitState:        metrics.circuitState.Load(),
		// Dedup
		DedupCacheSize:      metrics.dedupCacheSize.Load(),
		DedupEvictionsTotal: metrics.dedupEvictionsTotal.Load(),
		DedupHitsTotal:      metrics.dedupHitsTotal.Load(),
		DedupMissesTotal:    metrics.dedupMissesTotal.Load(),
		// Filtering
		FilteredDataTotal: metrics.filteredDataTotal.Load(),
		// Sampling
		SampledOutTotal: metrics.sampledOutTotal.Load(),
		SampledTotal:    metrics.sampledTotal.Load(),
	}
}
func (metrics *Metrics) Write(w io.Writer) {
	buf := make([]byte, 0, 8192)
	for i := range metricDescriptions {
		d := &metricDescriptions[i]
		buf = append(buf, "# HELP "...)
		buf = append(buf, d.name...)
		buf = append(buf, ' ')
		buf = append(buf, d.help...)
		buf = append(buf, '\n', '#', ' ', 'T', 'Y', 'P', 'E', ' ')
		buf = append(buf, d.name...)
		buf = append(buf, ' ')
		buf = append(buf, d.format...)
		buf = append(buf, '\n')
		buf = append(buf, d.name...)
		buf = append(buf, ' ')
		buf = strconv.AppendInt(buf, d.value(metrics), 10)
		buf = append(buf, '\n')
	}
	if _, err := w.Write(buf); err != nil {
		fmt.Fprintf(DefaultWriterErr, "ulog: metrics write error: %v\n", err)
	}
}
func (metricsRegistry *MetricsRegistry) Get(name string) *Metrics {
	metricsRegistry.mu.RLock()
	defer metricsRegistry.mu.RUnlock()
	return metricsRegistry.metrics[name]
}
func (metricsRegistry *MetricsRegistry) Len() int {
	metricsRegistry.mu.RLock()
	defer metricsRegistry.mu.RUnlock()
	return len(metricsRegistry.metrics)
}
func (metricsRegistry *MetricsRegistry) Register(name string, metrics *Metrics) {
	if name == "" || metrics == nil {
		return
	}
	metricsRegistry.mu.Lock()
	defer metricsRegistry.mu.Unlock()
	metricsRegistry.metrics[name] = metrics
}
func (metricsRegistry *MetricsRegistry) Unregister(name string) {
	metricsRegistry.mu.Lock()
	defer metricsRegistry.mu.Unlock()
	delete(metricsRegistry.metrics, name)
}
func (metricsRegistry *MetricsRegistry) Write(w io.Writer) {
	metricsRegistry.mu.RLock()
	names := make([]string, 0, len(metricsRegistry.metrics))
	snapshot := make(map[string]*Metrics, len(metricsRegistry.metrics))
	for name, m := range metricsRegistry.metrics {
		names = append(names, name)
		snapshot[name] = m
	}
	metricsRegistry.mu.RUnlock()
	sort.Strings(names)

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for i := range metricDescriptions {
		d := &metricDescriptions[i]
		bw.WriteString("# HELP ")
		bw.WriteString(d.name)
		bw.WriteByte(' ')
		bw.WriteString(d.help)
		bw.WriteString("\n# TYPE ")
		bw.WriteString(d.name)
		bw.WriteByte(' ')
		bw.WriteString(d.format)
		bw.WriteByte('\n')
		for _, name := range names {
			fmt.Fprintf(bw, "%s{sink=%q} %d\n", d.name, name, d.value(snapshot[name]))
		}
	}
}

// Приватные константы
const batchSizeMinInit int64 = math.MaxInt64

// Приватные переменные
var defaultMetrics = NewMetrics()
var globalRegistry = NewMetricsRegistry()
var metricDescriptions = []metricDescription{
	{"ulog_async_dropped_total", "Total drops due to full async buffer.", "counter", func(metrics *Metrics) int64 { return metrics.asyncDroppedTotal.Load() }},
	{"ulog_failed_total", "Total send failures.", "counter", func(metrics *Metrics) int64 { return metrics.failedTotal.Load() }},
	{"ulog_rate_limited_total", "Total rate-limit responses.", "counter", func(metrics *Metrics) int64 { return metrics.rateLimitedTotal.Load() }},
	{"ulog_retried_total", "Total retries.", "counter", func(metrics *Metrics) int64 { return metrics.retriedTotal.Load() }},
	{"ulog_sent_bytes_total", "Total bytes sent.", "counter", func(metrics *Metrics) int64 { return metrics.sentBytes.Load() }},
	{"ulog_sent_total", "Total log entries sent.", "counter", func(metrics *Metrics) int64 { return metrics.sentTotal.Load() }},
	{"ulog_uptime_seconds", "Uptime in seconds.", "gauge", func(metrics *Metrics) int64 {
		return int64(time.Since(time.Unix(0, metrics.startTime.Load())).Seconds())
	}},
	// Batch
	{"ulog_batch_buffer_size", "Current batch buffer size.", "gauge", func(metrics *Metrics) int64 { return metrics.batchBufferSize.Load() }},
	{"ulog_batch_flushes_total", "Total batch flushes.", "counter", func(metrics *Metrics) int64 { return metrics.batchFlushesTotal.Load() }},
	{"ulog_batch_size_max", "Max batch size seen.", "gauge", func(metrics *Metrics) int64 { return metrics.batchSizeMax.Load() }},
	{"ulog_batch_size_min", "Min batch size seen.", "gauge",
		func(metrics *Metrics) int64 {
			v := metrics.batchSizeMin.Load()
			if v == batchSizeMinInit {
				return 0
			}
			return v
		}},
	{"ulog_batch_size_total", "Total entries in batches.", "counter", func(metrics *Metrics) int64 { return metrics.batchSizeTotal.Load() }},
	// Circuit breaker
	{"ulog_circuit_blocked_total", "Total blocked by circuit.", "counter", func(metrics *Metrics) int64 { return metrics.circuitBlockedTotal.Load() }},
	{"ulog_circuit_closes_total", "Total circuit closes.", "counter", func(metrics *Metrics) int64 { return metrics.circuitClosesTotal.Load() }},
	{"ulog_circuit_opens_total", "Total circuit opens.", "counter", func(metrics *Metrics) int64 { return metrics.circuitOpensTotal.Load() }},
	{"ulog_circuit_state", "Circuit state (0=closed, 1=open, 2=half-open).", "gauge", func(metrics *Metrics) int64 { return int64(metrics.circuitState.Load()) }},
	// Dedup
	{"ulog_dedup_cache_size", "Current dedup cache size.", "gauge", func(metrics *Metrics) int64 { return metrics.dedupCacheSize.Load() }},
	{"ulog_dedup_evictions_total", "Total dedup evictions.", "counter", func(metrics *Metrics) int64 { return metrics.dedupEvictionsTotal.Load() }},
	{"ulog_dedup_hits_total", "Total dedup hits.", "counter", func(metrics *Metrics) int64 { return metrics.dedupHitsTotal.Load() }},
	{"ulog_dedup_misses_total", "Total dedup misses.", "counter", func(metrics *Metrics) int64 { return metrics.dedupMissesTotal.Load() }},
	// Filtering
	{"ulog_filtered_data_total", "Total filtered by data type.", "counter", func(metrics *Metrics) int64 { return metrics.filteredDataTotal.Load() }},
	// Sampling
	{"ulog_sampled_out_total", "Total entries sampled out.", "counter", func(metrics *Metrics) int64 { return metrics.sampledOutTotal.Load() }},
	{"ulog_sampled_total", "Total entries sampled.", "counter", func(metrics *Metrics) int64 { return metrics.sampledTotal.Load() }},
}

// Приватные структуры
type noCopy struct{}

// Приватные функции
func atomicMaxInt64(target *atomic.Int64, value int64) int64 {
	for {
		cur := target.Load()
		if value <= cur {
			return cur
		}
		if target.CompareAndSwap(cur, value) {
			return value
		}
	}
}
func atomicMinInt64(target *atomic.Int64, value int64) int64 {
	for {
		cur := target.Load()
		if value >= cur {
			return cur
		}
		if target.CompareAndSwap(cur, value) {
			return value
		}
	}
}

// Приватные методы
func (*noCopy) Lock()                               {}
func (*noCopy) Unlock()                             {}
func (metrics *Metrics) incAsyncDropped()           { metrics.asyncDroppedTotal.Add(1) }
func (metrics *Metrics) incFailed()                 { metrics.failedTotal.Add(1) }
func (metrics *Metrics) incRateLimited()            { metrics.rateLimitedTotal.Add(1) }
func (metrics *Metrics) incRetried()                { metrics.retriedTotal.Add(1) }
func (metrics *Metrics) addSentBytes(n int64)       { metrics.sentBytes.Add(n) }
func (metrics *Metrics) incSent()                   { metrics.sentTotal.Add(1) }
func (metrics *Metrics) incSentN(n int64)           { metrics.sentTotal.Add(n) }
func (metrics *Metrics) setBatchBufferSize(n int64) { metrics.batchBufferSize.Store(n) }
func (metrics *Metrics) incBatchFlushes()           { metrics.batchFlushesTotal.Add(1) }
func (metrics *Metrics) addBatchSizeTotal(n int64)  { metrics.batchSizeTotal.Add(n) }
func (metrics *Metrics) incCircuitBlocked()         { metrics.circuitBlockedTotal.Add(1) }
func (metrics *Metrics) incCircuitCloses()          { metrics.circuitClosesTotal.Add(1) }
func (metrics *Metrics) incCircuitOpens()           { metrics.circuitOpensTotal.Add(1) }
func (metrics *Metrics) setCircuitState(s int32)    { metrics.circuitState.Store(s) }
func (metrics *Metrics) setDedupCacheSize(n int64)  { metrics.dedupCacheSize.Store(n) }
func (metrics *Metrics) incDedupEvictions()         { metrics.dedupEvictionsTotal.Add(1) }
func (metrics *Metrics) incDedupHits()              { metrics.dedupHitsTotal.Add(1) }
func (metrics *Metrics) incDedupMisses()            { metrics.dedupMissesTotal.Add(1) }
func (metrics *Metrics) incFilteredData()           { metrics.filteredDataTotal.Add(1) }
func (metrics *Metrics) incSampledOut()             { metrics.sampledOutTotal.Add(1) }
func (metrics *Metrics) incSampled()                { metrics.sampledTotal.Add(1) }
func (metrics *Metrics) reset() {
	metrics.asyncDroppedTotal.Store(0)
	metrics.failedTotal.Store(0)
	metrics.rateLimitedTotal.Store(0)
	metrics.retriedTotal.Store(0)
	metrics.sentBytes.Store(0)
	metrics.sentTotal.Store(0)
	// Batch
	metrics.batchBufferSize.Store(0)
	metrics.batchFlushesTotal.Store(0)
	metrics.batchSizeMax.Store(0)
	metrics.batchSizeMin.Store(batchSizeMinInit)
	metrics.batchSizeTotal.Store(0)
	// Circuit breaker
	metrics.circuitBlockedTotal.Store(0)
	metrics.circuitClosesTotal.Store(0)
	metrics.circuitOpensTotal.Store(0)
	metrics.circuitState.Store(0)
	// Dedup
	metrics.dedupCacheSize.Store(0)
	metrics.dedupEvictionsTotal.Store(0)
	metrics.dedupHitsTotal.Store(0)
	metrics.dedupMissesTotal.Store(0)
	// Filtering
	metrics.filteredDataTotal.Store(0)
	// Sampling
	metrics.sampledOutTotal.Store(0)
	metrics.sampledTotal.Store(0)
}
func (metrics *Metrics) resetStartTime() {
	metrics.startTime.Store(time.Now().UnixNano())
}
