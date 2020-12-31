package kv

import (
	"context"
	"math"
	"testing"

	"github.com/prysmaticlabs/prysm/shared/testutil/require"
)

func TestSaveLoadChunk(t *testing.T) {
	slasherDB := setupDB(t)
	chunk := make([]uint16, 256)
	for i := 0; i < len(chunk); i++ {
		chunk[i] = math.MaxUint16
	}
	keys := []uint64{1}
	ctx := context.Background()
	err := slasherDB.SaveChunks(ctx, 0 /* kind */, keys, [][]uint16{chunk})
	require.NoError(t, err)
	received, _, err := slasherDB.LoadChunk(ctx, 0 /* kind */, keys[0])
	require.NoError(t, err)
	t.Log(chunk)
	t.Log(received)
}
