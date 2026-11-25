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
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Публичные функции
func Test_Metrics(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics() returned nil")
	}
	if m.StartTime().IsZero() {
		t.Error("StartTime should be set")
	}
	if got := m.batchSizeMin.Load(); got != batchSizeMinInit {
		t.Errorf("BatchSizeMin = %d, want %d (sentinel)", got, batchSizeMinInit)
	}
}
func Test_Metrics_GetMetrics(t *testing.T) {
	t.Parallel()
	m := MetricsGet()
	if m == nil {
		t.Fatal("MetricsGet() returned nil")
	}
	if m != defaultMetrics {
		t.Error("MetricsGet() should return defaultMetrics")
	}
}
func Test_Metrics_Reset(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.asyncDroppedTotal.Store(1)
	m.failedTotal.Store(2)
	m.rateLimitedTotal.Store(3)
	m.retriedTotal.Store(4)
	m.sentBytes.Store(5)
	m.sentTotal.Store(6)
	m.batchBufferSize.Store(7)
	m.batchFlushesTotal.Store(8)
	m.batchMergeFailedTotal.Store(9)
	m.batchSizeMax.Store(10)
	m.batchSizeMin.Store(11)
	m.batchSizeTotal.Store(12)
	m.circuitBlockedTotal.Store(13)
	m.circuitClosesTotal.Store(14)
	m.circuitOpensTotal.Store(15)
	m.circuitState.Store(16)
	m.dedupCacheSize.Store(17)
	m.dedupEvictionsTotal.Store(18)
	m.dedupHitsTotal.Store(19)
	m.dedupMissesTotal.Store(20)
	m.filteredDataTotal.Store(21)
	m.sampledOutTotal.Store(22)
	m.sampledTotal.Store(23)
	m.reset()
	snap := m.Snapshot()
	checks := []struct {
		name string
		got  int64
	}{
		{"AsyncDroppedTotal", snap.AsyncDroppedTotal},
		{"FailedTotal", snap.FailedTotal},
		{"RateLimitedTotal", snap.RateLimitedTotal},
		{"RetriedTotal", snap.RetriedTotal},
		{"SentBytes", snap.SentBytes},
		{"SentTotal", snap.SentTotal},
		{"BatchBufferSize", snap.BatchBufferSize},
		{"BatchFlushesTotal", snap.BatchFlushesTotal},
		{"BatchMergeFailedTotal", snap.BatchMergeFailedTotal},
		{"BatchSizeMax", snap.BatchSizeMax},
		{"BatchSizeTotal", snap.BatchSizeTotal},
		{"CircuitBlockedTotal", snap.CircuitBlockedTotal},
		{"CircuitClosesTotal", snap.CircuitClosesTotal},
		{"CircuitOpensTotal", snap.CircuitOpensTotal},
		{"CircuitState", int64(snap.CircuitState)},
		{"DedupCacheSize", snap.DedupCacheSize},
		{"DedupEvictionsTotal", snap.DedupEvictionsTotal},
		{"DedupHitsTotal", snap.DedupHitsTotal},
		{"DedupMissesTotal", snap.DedupMissesTotal},
		{"FilteredDataTotal", snap.FilteredDataTotal},
		{"SampledOutTotal", snap.SampledOutTotal},
		{"SampledTotal", snap.SampledTotal},
	}
	for _, c := range checks {
		if c.got != 0 {
			t.Errorf("%s = %d, want 0", c.name, c.got)
		}
	}
	if got := m.batchSizeMin.Load(); got != batchSizeMinInit {
		t.Errorf("BatchSizeMin = %d, want sentinel %d", got, batchSizeMinInit)
	}
}
func Test_Metrics_ResetStartTime(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	old := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	m.startTime.Store(old.UnixNano())
	m.resetStartTime()
	got := m.StartTime()
	if got.IsZero() {
		t.Fatal("StartTime is zero after ResetStartTime")
	}
	if !got.After(old) {
		t.Errorf("StartTime was not updated: old=%v new=%v", old, got)
	}
	if got.Year() < 2020 {
		t.Errorf("StartTime looks wrong: %v", got)
	}
}
func Test_Metrics_Snapshot(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.asyncDroppedTotal.Store(1)
	m.failedTotal.Store(2)
	m.rateLimitedTotal.Store(3)
	m.retriedTotal.Store(4)
	m.sentBytes.Store(5)
	m.sentTotal.Store(6)
	m.batchBufferSize.Store(7)
	m.batchFlushesTotal.Store(8)
	m.batchMergeFailedTotal.Store(9)
	m.batchSizeMax.Store(10)
	m.batchSizeMin.Store(11)
	m.batchSizeTotal.Store(12)
	m.circuitBlockedTotal.Store(13)
	m.circuitClosesTotal.Store(14)
	m.circuitOpensTotal.Store(15)
	m.circuitState.Store(16)
	m.dedupCacheSize.Store(17)
	m.dedupEvictionsTotal.Store(18)
	m.dedupHitsTotal.Store(19)
	m.dedupMissesTotal.Store(20)
	m.filteredDataTotal.Store(21)
	m.sampledOutTotal.Store(22)
	m.sampledTotal.Store(23)
	snap := m.Snapshot()
	want := MetricsSnapshot{
		AsyncDroppedTotal:     1,
		FailedTotal:           2,
		RateLimitedTotal:      3,
		RetriedTotal:          4,
		SentBytes:             5,
		SentTotal:             6,
		BatchBufferSize:       7,
		BatchFlushesTotal:     8,
		BatchMergeFailedTotal: 9,
		BatchSizeMax:          10,
		BatchSizeMin:          11,
		BatchSizeTotal:        12,
		CircuitBlockedTotal:   13,
		CircuitClosesTotal:    14,
		CircuitOpensTotal:     15,
		CircuitState:          16,
		DedupCacheSize:        17,
		DedupEvictionsTotal:   18,
		DedupHitsTotal:        19,
		DedupMissesTotal:      20,
		FilteredDataTotal:     21,
		SampledOutTotal:       22,
		SampledTotal:          23,
	}
	if snap != want {
		t.Errorf("Snapshot mismatch:\n got: %+v\nwant: %+v", snap, want)
	}
}
func Test_Metrics_Write_AllNames(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.asyncDroppedTotal.Store(1)
	m.failedTotal.Store(2)
	m.rateLimitedTotal.Store(3)
	m.retriedTotal.Store(4)
	m.sentBytes.Store(5)
	m.sentTotal.Store(6)
	m.batchBufferSize.Store(7)
	m.batchFlushesTotal.Store(8)
	m.batchMergeFailedTotal.Store(9)
	m.batchSizeMax.Store(10)
	m.batchSizeMin.Store(11)
	m.batchSizeTotal.Store(12)
	m.circuitBlockedTotal.Store(13)
	m.circuitClosesTotal.Store(14)
	m.circuitOpensTotal.Store(15)
	m.circuitState.Store(16)
	m.dedupCacheSize.Store(17)
	m.dedupEvictionsTotal.Store(18)
	m.dedupHitsTotal.Store(19)
	m.dedupMissesTotal.Store(20)
	m.filteredDataTotal.Store(21)
	m.sampledOutTotal.Store(22)
	m.sampledTotal.Store(23)
	var sb strings.Builder
	m.Write(&sb)
	out := sb.String()
	checks := []struct {
		name string
		want int64
	}{
		{"ulog_async_dropped_total", 1},
		{"ulog_failed_total", 2},
		{"ulog_rate_limited_total", 3},
		{"ulog_retried_total", 4},
		{"ulog_sent_bytes_total", 5},
		{"ulog_sent_total", 6},
		{"ulog_batch_buffer_size", 7},
		{"ulog_batch_flushes_total", 8},
		{"ulog_batch_merge_failed_total", 9},
		{"ulog_batch_size_max", 10},
		{"ulog_batch_size_min", 11},
		{"ulog_batch_size_total", 12},
		{"ulog_circuit_blocked_total", 13},
		{"ulog_circuit_closes_total", 14},
		{"ulog_circuit_opens_total", 15},
		{"ulog_circuit_state", 16},
		{"ulog_dedup_cache_size", 17},
		{"ulog_dedup_evictions_total", 18},
		{"ulog_dedup_hits_total", 19},
		{"ulog_dedup_misses_total", 20},
		{"ulog_filtered_data_total", 21},
		{"ulog_sampled_out_total", 22},
		{"ulog_sampled_total", 23},
	}
	for _, c := range checks {
		expected := c.name + " " + strconv.FormatInt(c.want, 10) + "\n"
		if !strings.Contains(out, expected) {
			t.Errorf("missing or wrong value: %q\noutput:\n%s", expected, out)
		}
	}
	if !strings.Contains(out, "# HELP ulog_sent_total") {
		t.Error("missing HELP for ulog_sent_total")
	}
	if !strings.Contains(out, "# TYPE ulog_sent_total counter") {
		t.Error("missing TYPE counter for ulog_sent_total")
	}
	if !strings.Contains(out, "# TYPE ulog_uptime_seconds gauge") {
		t.Error("missing TYPE gauge for ulog_uptime_seconds")
	}
	if !strings.Contains(out, "# TYPE ulog_circuit_state gauge") {
		t.Error("missing TYPE gauge for ulog_circuit_state")
	}
}
func Test_Metrics_Write_BatchSizeMinSentinel(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	var sb strings.Builder
	m.Write(&sb)
	out := sb.String()
	sentinel := strconv.FormatInt(batchSizeMinInit, 10)
	if strings.Contains(out, "ulog_batch_size_min "+sentinel) {
		t.Errorf("sentinel value %d leaked into output", batchSizeMinInit)
	}
	if !strings.Contains(out, "ulog_batch_size_min 0\n") {
		t.Errorf("expected ulog_batch_size_min to be 0 when uninitialized; output:\n%s", out)
	}
}
func Test_Metrics_Write_BatchSizeMinRealValue(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.batchSizeMin.Store(42)
	var sb strings.Builder
	m.Write(&sb)
	if !strings.Contains(sb.String(), "ulog_batch_size_min 42\n") {
		t.Errorf("expected ulog_batch_size_min 42; output:\n%s", sb.String())
	}
}
func Test_Metrics_Handler(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.sentTotal.Store(999)
	m.circuitState.Store(1)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	MetricsHandler(m).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want prefix text/plain", ct)
	}
	if !strings.Contains(ct, "version=0.0.4") {
		t.Errorf("Content-Type = %q, want version=0.0.4", ct)
	}
	out := rec.Body.String()
	if !strings.Contains(out, "ulog_sent_total 999\n") {
		t.Errorf("body did not reflect injected metrics: %s", out)
	}
	if !strings.Contains(out, "ulog_circuit_state 1\n") {
		t.Errorf("body missing injected circuit_state: %s", out)
	}
}
func Test_Metrics_Handler_WithGlobal(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	MetricsHandler(MetricsGet()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "ulog_sent_total") {
		t.Error("body missing ulog_sent_total")
	}
}

func Test_Metrics_Handler_Nil(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	MetricsHandler(nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want prefix text/plain", ct)
	}
	if rec.Body.Len() == 0 {
		t.Error("body is empty")
	}
}
func Test_Metrics_Handler_Isolation(t *testing.T) {
	t.Parallel()
	m1 := NewMetrics()
	m2 := NewMetrics()
	m1.sentTotal.Store(111)
	m2.sentTotal.Store(222)
	req1 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec1 := httptest.NewRecorder()
	MetricsHandler(m1).ServeHTTP(rec1, req1)
	req2 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec2 := httptest.NewRecorder()
	MetricsHandler(m2).ServeHTTP(rec2, req2)
	if !strings.Contains(rec1.Body.String(), "ulog_sent_total 111\n") {
		t.Errorf("m1 output wrong: %s", rec1.Body.String())
	}
	if !strings.Contains(rec2.Body.String(), "ulog_sent_total 222\n") {
		t.Errorf("m2 output wrong: %s", rec2.Body.String())
	}
}
func Test_Metrics_Handler_Concurrent(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	m.sentTotal.Store(42)
	var wg sync.WaitGroup
	var failures atomic.Int32
	const workers = 16
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
				rec := httptest.NewRecorder()
				MetricsHandler(m).ServeHTTP(rec, req)
				if !strings.Contains(rec.Body.String(), "ulog_sent_total") {
					failures.Add(1)
					return
				}
			}
		}()
	}
	wg.Wait()
	if n := failures.Load(); n != 0 {
		t.Errorf("%d goroutines failed to find ulog_sent_total", n)
	}
}
func Test_Metrics_Concurrent(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	var wg sync.WaitGroup
	const workers = 32
	const iterations = 1000
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				m.sentTotal.Add(1)
				m.failedTotal.Add(1)
				m.asyncDroppedTotal.Add(1)
				_ = m.Snapshot()
			}
		}()
	}
	wg.Wait()
	snap := m.Snapshot()
	want := int64(workers * iterations)
	if snap.SentTotal != want {
		t.Errorf("SentTotal = %d, want %d", snap.SentTotal, want)
	}
	if snap.FailedTotal != want {
		t.Errorf("FailedTotal = %d, want %d", snap.FailedTotal, want)
	}
	if snap.AsyncDroppedTotal != want {
		t.Errorf("AsyncDroppedTotal = %d, want %d", snap.AsyncDroppedTotal, want)
	}
}
func Test_Metrics_AtomicMaxInt64(t *testing.T) {
	t.Parallel()
	var target atomic.Int64
	target.Store(10)
	if got := atomicMaxInt64(&target, 5); got != 10 {
		t.Errorf("atomicMaxInt64(5) = %d, want 10", got)
	}
	if got := atomicMaxInt64(&target, 20); got != 20 {
		t.Errorf("atomicMaxInt64(20) = %d, want 20", got)
	}
	if target.Load() != 20 {
		t.Errorf("target = %d, want 20", target.Load())
	}
}
func Test_Metrics_AtomicMinInt64(t *testing.T) {
	t.Parallel()
	var target atomic.Int64
	target.Store(batchSizeMinInit)
	if got := atomicMinInt64(&target, 100); got != 100 {
		t.Errorf("atomicMinInt64(100) = %d, want 100", got)
	}
	if got := atomicMinInt64(&target, 50); got != 50 {
		t.Errorf("atomicMinInt64(50) = %d, want 50", got)
	}
	if got := atomicMinInt64(&target, 200); got != 50 {
		t.Errorf("atomicMinInt64(200) = %d, want 50", got)
	}
	if target.Load() != 50 {
		t.Errorf("target = %d, want 50", target.Load())
	}
}
func Test_Metrics_AtomicMaxInt64_MaxValue(t *testing.T) {
	t.Parallel()
	var target atomic.Int64
	target.Store(math.MaxInt64)
	if got := atomicMaxInt64(&target, math.MaxInt64); got != math.MaxInt64 {
		t.Errorf("got %d, want MaxInt64", got)
	}
}
func Test_Metrics_Isolation(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	requests1 := 0
	requests2 := 0
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests1++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests2++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()
	m1 := NewMetrics()
	m2 := NewMetrics()
	sink1 := NewSinkHttp(server1.URL,
		WithHttpMetrics(m1),
		WithHttpDisabledBatch(),
		WithHttpDisabledCircuit(),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sink1.Close()
	sink2 := NewSinkHttp(server2.URL,
		WithHttpMetrics(m2),
		WithHttpDisabledBatch(),
		WithHttpDisabledCircuit(),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sink2.Close()
	attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
	fields := []Field{String("message", "test")}
	for i := 0; i < 3; i++ {
		sink1.WriteWithAttributes(attrs, fields)
	}
	sink2.WriteWithAttributes(attrs, fields)
	if got := m1.sentTotal.Load(); got != 3 {
		t.Errorf("m1.SentTotal = %d, want 3", got)
	}
	if got := m2.sentTotal.Load(); got != 1 {
		t.Errorf("m2.SentTotal = %d, want 1", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if requests1 != 3 {
		t.Errorf("server1 got %d requests, want 3", requests1)
	}
	if requests2 != 1 {
		t.Errorf("server2 got %d requests, want 1", requests2)
	}
}
func Test_MetricsRegistry_Concurrent(t *testing.T) {
	t.Parallel()
	r := NewMetricsRegistry()
	var wg sync.WaitGroup
	const workers = 16
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				name := "sink" + strconv.Itoa(id)
				m := NewMetrics()
				m.sentTotal.Store(int64(j))
				r.Register(name, m)
				_ = r.Len()
				_ = r.Get(name)
				r.Unregister(name)
			}
		}(i)
	}
	wg.Wait()
	if r.Len() != 0 {
		t.Errorf("Len = %d, want 0 after all unregister", r.Len())
	}
}
func Test_MetricsRegistry_Handler(t *testing.T) {
	t.Parallel()
	r := NewMetricsRegistry()
	m := NewMetrics()
	m.sentTotal.Store(42)
	r.Register("test", m)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	MetricsRegistryHandler(r).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `ulog_sent_total{sink="test"} 42`) {
		t.Errorf("missing labeled series: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "# HELP ulog_sent_total") {
		t.Error("missing HELP header")
	}
	if !strings.Contains(rec.Body.String(), "# TYPE ulog_sent_total counter") {
		t.Error("missing TYPE header")
	}
}
func Test_MetricsRegistry_HandlerNil(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	MetricsRegistryHandler(nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
func Test_MetricsRegistry_Register(t *testing.T) {
	t.Parallel()
	r := NewMetricsRegistry()
	m1 := NewMetrics()
	m2 := NewMetrics()
	r.Register("a", m1)
	r.Register("b", m2)
	if r.Len() != 2 {
		t.Errorf("Len = %d, want 2", r.Len())
	}
	if r.Get("a") != m1 {
		t.Error("Get(a) != m1")
	}
	r.Unregister("a")
	if r.Len() != 1 {
		t.Errorf("Len = %d, want 1", r.Len())
	}
	if r.Get("a") != nil {
		t.Error("Get(a) should be nil")
	}
}
func Test_MetricsRegistry_SinkFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	m := NewMetrics()
	reg := NewMetricsRegistry()
	sink, err := NewSinkFile(
		filepath.Join(dir, "test.log"),
		WithFileMetrics(m),
		WithFileMetricsName(reg, "file"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sink.Close()
	if reg.Len() != 1 {
		t.Errorf("registry.Len = %d, want 1", reg.Len())
	}
	if reg.Get("file") != m {
		t.Error("registry.Get(file) != m")
	}
	attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
	fields := []Field{String("message", "test")}
	sink.WriteWithAttributes(attrs, fields)
	if m.SentTotal() != 1 {
		t.Errorf("SentTotal = %d, want 1", m.SentTotal())
	}
}
func Test_MetricsRegistry_SinkHttp(t *testing.T) {
	t.Parallel()
	m := NewMetrics()
	reg := NewMetricsRegistry()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	sink := NewSinkHttp(srv.URL,
		WithHttpMetrics(m),
		WithHttpMetricsName(reg, "test"),
		WithHttpDisabledBatch(),
		WithHttpDisabledCircuit(),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sink.Close()
	if reg.Len() != 1 {
		t.Errorf("registry.Len = %d, want 1", reg.Len())
	}
	if reg.Get("test") != m {
		t.Error("registry.Get(test) != m")
	}
	attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
	fields := []Field{String("message", "test")}
	sink.WriteWithAttributes(attrs, fields)
	if m.SentTotal() != 1 {
		t.Errorf("SentTotal = %d, want 1", m.SentTotal())
	}
}
func Test_MetricsRegistry_Write(t *testing.T) {
	t.Parallel()
	r := NewMetricsRegistry()
	m1 := NewMetrics()
	m1.sentTotal.Store(100)
	m2 := NewMetrics()
	m2.sentTotal.Store(200)
	r.Register("discord", m1)
	r.Register("slack", m2)
	var sb strings.Builder
	r.Write(&sb)
	out := sb.String()
	if !strings.Contains(out, `ulog_sent_total{sink="discord"} 100`) {
		t.Errorf("missing discord series: %s", out)
	}
	if !strings.Contains(out, `ulog_sent_total{sink="slack"} 200`) {
		t.Errorf("missing slack series: %s", out)
	}
	if strings.Count(out, "# HELP ulog_sent_total ") != 1 {
		t.Errorf("HELP for ulog_sent_total written %d times, want 1", strings.Count(out, "# HELP ulog_sent_total "))
	}
}
