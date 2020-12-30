package kv

import (
	"context"

	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

type AttesterRecord struct {
	AttestationDataHash [32]byte
	SourceEpoch         uint64
	TargetEpoch         uint64
}

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

func (db *Store) AttestationRecordForValidator(
	ctx context.Context,
	validatorIdx,
	targetEpoch uint64,
) (*AttesterRecord, error) {
	return nil, nil
}
