package kv

import (
	"bytes"
	"errors"

	"go.etcd.io/bbolt"
)

var migrationSourceTargetEpochsBucketKey = []byte("source_target_epochs_bucket_0")

func migrateSourceTargetEpochsBucket(tx *bbolt.Tx) error {
	mb := tx.Bucket(migrationsBucket)
	if v := mb.Get(migrationSourceTargetEpochsBucketKey); bytes.Equal(v, migrationCompleted) {
		return nil
	}

	pksBucket := tx.Bucket(pubKeysBucket)
	err := pksBucket.ForEach(func(pk, _ []byte) error {
		pkb := pksBucket.Bucket(pk)
		if pkb == nil {
			return errors.New("nil bucket entry in pubkeys bucket")
		}

		sourceBucket := pkb.Bucket(attestationSourceEpochsBucket)
		if sourceBucket == nil {
			return nil
		}

		targetBucket, err := pkb.CreateBucketIfNotExists(attestationTargetEpochsBucket)
		if err != nil {
			return err
		}

		return sourceBucket.ForEach(func(sourceEpochBytes, targetEpochsBytes []byte) error {
			for i := 0; i < len(targetEpochsBytes); i += 8 {
				if err := insertTargetSource(targetBucket, targetEpochsBytes[i:i+8], sourceEpochBytes); err != nil {
					return err
				}
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	return mb.Put(migrationSourceTargetEpochsBucketKey, migrationCompleted)
}

func insertTargetSource(bkt *bbolt.Bucket, targetEpochBytes, sourceEpochBytes []byte) error {
	var existingAttestedSourceBytes []byte
	if existing := bkt.Get(targetEpochBytes); existing != nil {
		existingAttestedSourceBytes = append(existing, sourceEpochBytes...)
	} else {
		existingAttestedSourceBytes = sourceEpochBytes
	}
	return bkt.Put(targetEpochBytes, existingAttestedSourceBytes)
}
