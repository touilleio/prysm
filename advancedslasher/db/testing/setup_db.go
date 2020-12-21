// Package testing defines useful helper functions for unit tests with
// the slasher database.
package testing

import (
	"testing"

	slasherDB "github.com/prysmaticlabs/prysm/advancedslasher/db"
	"github.com/prysmaticlabs/prysm/advancedslasher/db/kv"
)

// SetupSlasherDB instantiates and returns a SlasherDB instance.
func SetupSlasherDB(t testing.TB) *kv.Store {
	cfg := &kv.Config{}
	db, err := slasherDB.NewDB(t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("Failed to instantiate DB: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}
	})
	return db
}
