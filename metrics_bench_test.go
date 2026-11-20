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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Benchmark_Metrics_Handler(b *testing.B) {
	m := NewMetrics()
	m.sentTotal.Store(12345)
	handler := MetricsHandler(m)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	b.ReportAllocs()
	for b.Loop() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
func Benchmark_Metrics_Snapshot(b *testing.B) {
	m := NewMetrics()
	m.sentTotal.Store(12345)
	b.ReportAllocs()
	for b.Loop() {
		_ = m.Snapshot()
	}
}
func Benchmark_Metrics_Snapshot_Parallel(b *testing.B) {
	m := NewMetrics()
	m.sentTotal.Store(12345)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Snapshot()
		}
	})
}
func Benchmark_Metrics_Write(b *testing.B) {
	m := NewMetrics()
	m.sentTotal.Store(12345)
	m.failedTotal.Store(67)
	m.circuitState.Store(1)
	b.ReportAllocs()
	for b.Loop() {
		var sb strings.Builder
		m.Write(&sb)
	}
}
func Benchmark_Metrics_Write_Discard(b *testing.B) {
	m := NewMetrics()
	m.sentTotal.Store(12345)
	b.ReportAllocs()
	for b.Loop() {
		m.Write(io.Discard)
	}
}
