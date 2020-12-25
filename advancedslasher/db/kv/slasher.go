package kv

import (
	"context"

	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

// SlasherChunkForAttestation --
func (db *Store) SlasherChunkForAttestation(
	ctx context.Context, validatorIdx uint64, chunkIndex uint64,
) ([]byte, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.SlasherChunkForAttestation")
	defer span.End()
	var chunk []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		hashesBucket := tx.Bucket(slasherChunkHashesBucket)
		chunksBucket := tx.Bucket(slasherChunksBucket)
		key := append(
			bytesutil.Uint64ToBytesLittleEndian(validatorIdx),
			bytesutil.Uint64ToBytesLittleEndian(chunkIndex)...,
		)
		chunkHash := hashesBucket.Get(key)
		if chunkHash == nil {
			return nil
		}
		rawChunk := chunksBucket.Get(chunkHash)
		copy(chunk, rawChunk)
		return nil
	})
	return chunk, err
}
