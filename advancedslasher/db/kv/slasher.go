package kv

import (
	"context"

	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

// LatestEpochWrittenForValidator --
func (db *Store) LatestEpochWrittenForValidator(
	ctx context.Context, validatorIdx uint64,
) (uint64, bool, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.LatestEpochWrittenForValidator")
	defer span.End()
	var epoch uint64
	err := db.db.View(func(tx *bolt.Tx) error {
		return nil
	})
	return epoch, true, err
}
