// Copyright (c) 2026 Mikhail Dadaev
// All rights reserved.
//
// This source code is licensed under the MIT License found in the
// LICENSE file in the root directory of this source tree.
package ulog

import (
	"testing"
)

// Бенчмарки компонентов
func BenchmarkDebug(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("message debug")
	}
}
func BenchmarkDebugf(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debugf("message debug #%d", i)
	}
}
func BenchmarkError(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("message error")
	}
}
func BenchmarkErrorf(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Errorf("message error #%d", i)
	}
}
func BenchmarkInfo(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("message info")
	}
}
func BenchmarkInfof(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infof("message info #%d", i)
	}
}
func BenchmarkWarn(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Warn("message warning")
	}
}
func BenchmarkWarnf(b *testing.B) {
	logger := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Warnf("message warning #%d", i)
	}
}
func BenchmarkLoggerWriter(b *testing.B) {
	logger := New()
	writer := NewWithWriter(logger, LevelInfo)
	data := []byte("test message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(data)
	}
}
func BenchmarkWithDisabledLevel(b *testing.B) {
	logger := New()
	logger.SetLevel(LevelError)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("this should be skipped")
	}
}
