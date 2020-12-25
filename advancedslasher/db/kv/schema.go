package kv

const (
	DEFAULT_CHUNK_SIZE uint64 = 256
)

var (
	// Slasher related buckets.
	slasherChunkHashesBucket = []byte("slasher-chunk-hashes")
	slasherChunksBucket      = []byte("slasher-chunks")
)
