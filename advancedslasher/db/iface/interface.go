// Package iface defines an interface for the slasher database,
package iface

import (
	"context"
	"io"

	"github.com/prysmaticlabs/prysm/advancedslasher/db/kv"
	"github.com/prysmaticlabs/prysm/shared/backuputil"
)

// Database represents a full access database with the proper DB helper functions.
type Database interface {
	io.Closer
	backuputil.BackupExporter
	DatabasePath() string
	ClearDB() error

	UpdateLatestEpochWrittenForValidator(ctx context.Context, validatorIdx, epoch uint64) error
	LatestEpochWrittenForValidator(ctx context.Context, validatorIdx uint64) (uint64, bool, error)

	AttestationRecordForValidator(ctx context.Context, validatorIdx, targetEpoch uint64) (*kv.AttesterRecord, error)
	LoadChunk(ctx context.Context, key uint64) ([]uint16, bool, error)
}
