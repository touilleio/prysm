package kv

import (
	"context"

	ssz "github.com/ferranbt/fastssz"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
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
	ctx, span := trace.StartSpan(ctx, "AdvancedSlasherDB.LatestEpochWrittenForValidator")
	defer span.End()
	var epoch uint64
	var exists bool
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(epochByValidatorBucket)
		enc := ssz.MarshalUint64(make([]byte, 0), validatorIdx)
		epochBytes := bkt.Get(enc)
		if epochBytes == nil {
			return nil
		}
		epoch = ssz.UnmarshallUint64(epochBytes)
		exists = true
		return nil
	})
	return epoch, exists, err
}

// UpdateLatestEpochWrittenForValidator --
func (db *Store) UpdateLatestEpochWrittenForValidator(ctx context.Context, validatorIdx, epoch uint64) error {
	ctx, span := trace.StartSpan(ctx, "AdvancedSlasherDB.UpdateLatestEpochWrittenForValidator")
	defer span.End()
	return db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(epochByValidatorBucket)
		key := ssz.MarshalUint64(make([]byte, 0), validatorIdx)
		val := ssz.MarshalUint64(make([]byte, 0), epoch)
		return bkt.Put(key, val)
	})
}

func (db *Store) AttestationRecordForValidator(
	ctx context.Context,
	validatorIdx,
	targetEpoch uint64,
) (*AttesterRecord, error) {
	ctx, span := trace.StartSpan(ctx, "AdvancedSlasherDB.AttestationRecordForValidator")
	defer span.End()
	var record *AttesterRecord
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(attestationRecordsBucket)
		encIdx := ssz.MarshalUint64(make([]byte, 0), validatorIdx)
		encEpoch := ssz.MarshalUint64(make([]byte, 0), targetEpoch)
		key := append(encIdx, encEpoch...)
		value := bkt.Get(key)
		if value == nil {
			return nil
		}
		record = &AttesterRecord{
			SourceEpoch: ssz.UnmarshallUint64(value[0:8]),
			TargetEpoch: ssz.UnmarshallUint64(value[8:]),
		}
		return nil
	})
	return record, err
}

func (db *Store) SaveAttestationRecordForValidator(
	ctx context.Context,
	validatorIdx uint64,
	attestation *ethpb.IndexedAttestation,
) error {
	ctx, span := trace.StartSpan(ctx, "AdvancedSlasherDB.SaveAttestationRecordForValidator")
	defer span.End()
	return db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(attestationRecordsBucket)
		encIdx := ssz.MarshalUint64(make([]byte, 0), validatorIdx)
		encEpoch := ssz.MarshalUint64(make([]byte, 0), attestation.Data.Target.Epoch)
		key := append(encIdx, encEpoch...)
		value := make([]byte, 16)
		copy(value[0:8], ssz.MarshalUint64(make([]byte, 0), attestation.Data.Source.Epoch))
		copy(value[8:], ssz.MarshalUint64(make([]byte, 0), attestation.Data.Target.Epoch))
		return bkt.Put(key, value)
	})
}

func (db *Store) LoadChunk(ctx context.Context, key uint64) ([]uint16, bool, error) {
	ctx, span := trace.StartSpan(ctx, "AdvancedSlasherDB.LoadChunk")
	defer span.End()
	var chunk []uint16
	var exists bool
	err := db.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(slasherChunksBucket)
		keyBytes := ssz.MarshalUint64(make([]byte, 0), key)
		chunkBytes := bkt.Get(keyBytes)
		if chunkBytes == nil {
			return nil
		}
		chunk = make([]uint16, 0)
		for i := 0; i < len(chunkBytes); i += 2 {
			distance := ssz.UnmarshallUint16(chunkBytes[i : i+2])
			chunk = append(chunk, distance)
		}
		exists = true
		return nil
	})
	return chunk, exists, err
}

func (db *Store) SaveChunk(ctx context.Context, key uint64, chunk []uint16) error {
	ctx, span := trace.StartSpan(ctx, "AdvancedSlasherDB.SaveChunk")
	defer span.End()
	return db.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(slasherChunksBucket)
		keyBytes := ssz.MarshalUint64(make([]byte, 0), key)
		val := make([]byte, 0)
		for i := 0; i < len(chunk); i++ {
			val = append(val, ssz.MarshalUint16(make([]byte, 0), chunk[i])...)
		}
		return bkt.Put(keyBytes, val)
	})
}
