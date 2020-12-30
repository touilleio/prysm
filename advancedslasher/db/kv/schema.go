package kv

var (
	// Slasher related buckets.
	epochByValidatorBucket   = []byte("epochs-by-validator")
	attestationRecordsBucket = []byte("att-records")
	indexedAttsBucket        = []byte("indexed-atts")
	slasherChunksBucket      = []byte("chunks")
)
