package service_test

import (
	"crypto/rand"
	"testing"

	"github.com/vindyang/cs464-project/backend/services/sharding/internal/app"
)

func makeRandomData(size int) []byte {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return data
}

func BenchmarkEncodeChunk(b *testing.B) {
	svc := app.NewShardingService()
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1 * 1024},
		{"1MB", 1 * 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}

	for _, tc := range sizes {
		data := makeRandomData(tc.size)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tc.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := svc.EncodeChunk(data, 4, 6); err != nil {
					b.Fatalf("EncodeChunk failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkDecodeChunk_AllPresent(b *testing.B) {
	svc := app.NewShardingService()
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1 * 1024},
		{"1MB", 1 * 1024 * 1024},
	}

	for _, tc := range sizes {
		data := makeRandomData(tc.size)
		shards, err := svc.EncodeChunk(data, 4, 6)
		if err != nil {
			b.Fatalf("setup EncodeChunk failed: %v", err)
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tc.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Copy shards so each iteration starts with full set
				shardsCopy := make([][]byte, len(shards))
				copy(shardsCopy, shards)
				if _, err := svc.DecodeChunk(shardsCopy, 4, 6); err != nil {
					b.Fatalf("DecodeChunk failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkDecodeChunk_WithReconstruction(b *testing.B) {
	svc := app.NewShardingService()
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1 * 1024},
		{"1MB", 1 * 1024 * 1024},
	}

	for _, tc := range sizes {
		data := makeRandomData(tc.size)
		shards, err := svc.EncodeChunk(data, 4, 6)
		if err != nil {
			b.Fatalf("setup EncodeChunk failed: %v", err)
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tc.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Copy shards and nil out 2 parity shards to force Reed-Solomon reconstruction
				shardsCopy := make([][]byte, len(shards))
				copy(shardsCopy, shards)
				shardsCopy[4] = nil
				shardsCopy[5] = nil
				if _, err := svc.DecodeChunk(shardsCopy, 4, 6); err != nil {
					b.Fatalf("DecodeChunk with reconstruction failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkChecksumSHA256(b *testing.B) {
	svc := app.NewShardingService()
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1 * 1024},
		{"1MB", 1 * 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}

	for _, tc := range sizes {
		data := makeRandomData(tc.size)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tc.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = svc.CalculateChecksum(data)
			}
		})
	}
}
