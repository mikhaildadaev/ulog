// Copyright (c) 2026 Mikhail Dadaev
// All rights reserved.
//
// This source code is licensed under the MIT License found in the
// LICENSE file in the root directory of this source tree.
package ulog

import (
	"testing"
)

// Публичные функции
func BenchmarkInfo(b *testing.B) {
	logger := New()
	for i := 0; i < b.N; i++ {
		logger.Info("test")
	}
}
func BenchmarkInfof(b *testing.B) {
	logger := New()
	for i := 0; i < b.N; i++ {
		logger.Infof("test %d", i)
	}
}
