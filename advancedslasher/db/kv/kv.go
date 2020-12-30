// Package kv defines a bolt-db, key-value store implementation of
// the slasher database interface.
package kv

import (
	"context"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/shared/fileutil"
	"github.com/prysmaticlabs/prysm/shared/params"
	bolt "go.etcd.io/bbolt"
)

const (
	// SlasherDbDirName is the name of the directory containing the slasher database.
	SlasherDbDirName = "slasherdata"
	// DatabaseFileName is the name of the slasher database.
	DatabaseFileName = "slasher.db"
)

// Store defines an implementation of the slasher Database interface
// using BoltDB as the underlying persistent kv-store for eth2.
type Store struct {
	db           *bolt.DB
	databasePath string
}

// Config options for the slasher db.
type Config struct {
}

// Close closes the underlying boltdb database.
func (db *Store) Close() error {
	return db.db.Close()
}

// ClearDB removes any previously stored data at the configured data directory.
func (db *Store) ClearDB() error {
	if _, err := os.Stat(db.databasePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filepath.Join(db.databasePath, DatabaseFileName))
}

// DatabasePath at which this database writes files.
func (db *Store) DatabasePath() string {
	return db.databasePath
}

func createBuckets(tx *bolt.Tx, buckets ...[]byte) error {
	for _, bucket := range buckets {
		if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
			return err
		}
	}
	return nil
}

// NewKVStore initializes a new boltDB key-value store at the directory
// path specified, creates the kv-buckets based on the schema, and stores
// an open connection db object as a property of the Store struct.
func NewKVStore(dirPath string, cfg *Config) (*Store, error) {
	hasDir, err := fileutil.HasDir(dirPath)
	if err != nil {
		return nil, err
	}
	if !hasDir {
		if err := fileutil.MkdirAll(dirPath); err != nil {
			return nil, err
		}
	}

	datafile := path.Join(dirPath, DatabaseFileName)
	boltDB, err := bolt.Open(datafile, params.BeaconIoConfig().ReadWritePermissions, &bolt.Options{Timeout: params.BeaconIoConfig().BoltTimeout})
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return nil, errors.New("cannot obtain database lock, database may be in use by another process")
		}
		return nil, err
	}
	kv := &Store{db: boltDB, databasePath: dirPath}

	if err := kv.db.Update(func(tx *bolt.Tx) error {
		return createBuckets(
			tx,
			slasherChunkHashesBucket,
			slasherChunksBucket,
		)
	}); err != nil {
		return nil, err
	}

	return kv, err
}

// Size returns the db size in bytes.
func (db *Store) Size() (int64, error) {
	var size int64
	err := db.db.View(func(tx *bolt.Tx) error {
		size = tx.Size()
		return nil
	})
	return size, err
}

func (db *Store) SavePublicKey(ctx context.Context, validatorIdx uint64, pubKey []byte) error {
	return nil
}
