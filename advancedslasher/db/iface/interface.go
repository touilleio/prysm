// Package iface defines an interface for the slasher database,
package iface

import (
	"context"
	"io"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/advancedslasher/db/kv"
	"github.com/prysmaticlabs/prysm/shared/backuputil"
)

// Database represents a full access database with the proper DB helper functions.
type Database interface {
	io.Closer
	backuputil.BackupExporter
	DatabasePath() string
	ClearDB() error

	// Epochs written.
	UpdateLatestEpochWrittenForValidators(ctx context.Context, validatorIndices []uint64, epoch uint64) error
	LatestEpochWrittenForValidator(ctx context.Context, validatorIdx uint64) (uint64, bool, error)

	// Attesting records.
	AttestationRecordForValidator(ctx context.Context, validatorIdx, targetEpoch uint64) (*kv.AttesterRecord, error)
	SaveAttestationRecordForValidator(
		ctx context.Context,
		validatorIdx uint64,
		attestation *ethpb.IndexedAttestation,
	) error

	// Chunks.
	LoadChunk(ctx context.Context, kind uint8, key uint64) ([]uint16, bool, error)
	SaveChunks(ctx context.Context, kind uint8, chunkKeys []uint64, chunks [][]uint16) error
}
