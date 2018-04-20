package main

import (
	"testing"
)

func Benchmark_RandStringId_12Symbols(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandStringId(12)
	}
}

func Benchmark_createSecret_12_lower_upper_digits_symbols(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createSecret(12, false, false, false)
	}
}

func Benchmark_createSecret_12_lower_upper_digits(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createSecret(12, false, false, true)
	}
}

func Benchmark_createSecret_12_lower_upper(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createSecret(12, false, true, true)
	}
}

func Benchmark_createSecret_12_lower(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createSecret(12, true, true, true)
	}
}
