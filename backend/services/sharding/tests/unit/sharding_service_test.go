package service_test

import (
	"bytes"
	"testing"

	"github.com/vindyang/cs464-project/backend/services/shared/service"
)

func TestChunkFile_SplitsAsExpected(t *testing.T) {
	svc := service.NewShardingService()
	data := []byte("abcdefghijklmnopqrstuvwxyz") // 26 bytes

	chunks, err := svc.ChunkFile(data, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	if len(chunks[0]) != 10 || len(chunks[1]) != 10 || len(chunks[2]) != 6 {
		t.Fatalf("unexpected chunk sizes: %d, %d, %d", len(chunks[0]), len(chunks[1]), len(chunks[2]))
	}
}

func TestChunkFile_Validation(t *testing.T) {
	svc := service.NewShardingService()

	if _, err := svc.ChunkFile([]byte("abc"), 0); err == nil {
		t.Fatalf("expected error for non-positive chunk size")
	}

	if _, err := svc.ChunkFile([]byte{}, 4); err == nil {
		t.Fatalf("expected error for empty data")
	}
}

func TestEncodeDecodeChunk_RoundTripAndReconstruct(t *testing.T) {
	svc := service.NewShardingService()
	data := []byte("abcdefghijkl") // 12 bytes, divisible by K=3

	shards, err := svc.EncodeChunk(data, 3, 5)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decoded, err := svc.DecodeChunk(shards, 3, 5)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !bytes.Equal(decoded, data) {
		t.Fatalf("decoded data mismatch: got %q want %q", decoded, data)
	}

	// Simulate one missing shard and verify reconstruction
	shards[1] = nil
	reconstructed, err := svc.DecodeChunk(shards, 3, 5)
	if err != nil {
		t.Fatalf("decode with missing shard failed: %v", err)
	}

	if !bytes.Equal(reconstructed, data) {
		t.Fatalf("reconstructed data mismatch: got %q want %q", reconstructed, data)
	}
}

func TestEncodeDecodeChunk_Validation(t *testing.T) {
	svc := service.NewShardingService()

	if _, err := svc.EncodeChunk([]byte("abc"), 0, 3); err == nil {
		t.Fatalf("expected error for invalid K")
	}

	if _, err := svc.EncodeChunk([]byte("abc"), 4, 3); err == nil {
		t.Fatalf("expected error when K > N")
	}

	if _, err := svc.EncodeChunk([]byte{}, 2, 3); err == nil {
		t.Fatalf("expected error for empty chunk data")
	}

	if _, err := svc.DecodeChunk(make([][]byte, 4), 3, 5); err == nil {
		t.Fatalf("expected error for shard count mismatch")
	}

	insufficient := make([][]byte, 5)
	insufficient[0] = []byte{1, 2}
	insufficient[2] = []byte{3, 4}
	if _, err := svc.DecodeChunk(insufficient, 3, 5); err == nil {
		t.Fatalf("expected error for insufficient available shards")
	}
}

func TestChecksumFunctions(t *testing.T) {
	svc := service.NewShardingService()
	data := []byte("checksum-data")

	h1 := svc.CalculateChecksum(data)
	h2 := svc.CalculateChecksum(data)
	if h1 != h2 || h1 == "" {
		t.Fatalf("expected stable non-empty checksum, got %q and %q", h1, h2)
	}

	readerHash, err := service.CalculateChecksumFromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("CalculateChecksumFromReader failed: %v", err)
	}

	if readerHash != h1 {
		t.Fatalf("reader checksum mismatch: got %q want %q", readerHash, h1)
	}
}
