// Package iface defines an interface for the slasher database,
package iface

import (
	"context"
	"io"

	"github.com/prysmaticlabs/prysm/shared/backuputil"
)

// Database represents a full access database with the proper DB helper functions.
type Database interface {
	io.Closer
	backuputil.BackupExporter
	DatabasePath() string
	ClearDB() error
	SavePublicKey(ctx context.Context, validatorIdx uint64, pubKey []byte) error
}
