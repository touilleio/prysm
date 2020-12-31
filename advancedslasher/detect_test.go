package main

import (
	"context"
	"testing"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	dbTest "github.com/prysmaticlabs/prysm/advancedslasher/db/testing"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func TestSlasher_detectAttestationBatch(t *testing.T) {
	hook := logTest.NewGlobal()

	t.Run("single validator, single chunk, surrounding", func(tt *testing.T) {
		slasherDB := dbTest.SetupSlasherDB(t)
		slasher, err := NewSlasher(context.Background(), &ServiceConfig{
			SlasherDB: slasherDB,
		})
		require.NoError(t, err)
		currentEpoch := uint64(5)
		validatorIndices := []uint64{1}
		validatorChunkIdx := uint64(0)
		atts := []*ethpb.IndexedAttestation{
			createAtt(validatorIndices, 3 /* source */, 4 /* target */),
		}
		slasher.detectAttestationBatch(validatorChunkIdx, atts, currentEpoch)
		atts = []*ethpb.IndexedAttestation{
			createAtt(validatorIndices, 2 /* source */, 5 /* target */),
		}
		slasher.detectAttestationBatch(validatorChunkIdx, atts, currentEpoch)
		assert.LogsContain(t, hook, "Slashing found: SURROUNDING_VOTE")
	})
	t.Run("single validator, single chunk, surrounded", func(tt *testing.T) {
		slasherDB := dbTest.SetupSlasherDB(t)
		slasher, err := NewSlasher(context.Background(), &ServiceConfig{
			SlasherDB: slasherDB,
		})
		require.NoError(t, err)
		currentEpoch := uint64(5)
		validatorIndices := []uint64{1}
		validatorChunkIdx := uint64(0)
		atts := []*ethpb.IndexedAttestation{
			createAtt(validatorIndices, 2 /* source */, 5 /* target */),
		}
		slasher.detectAttestationBatch(validatorChunkIdx, atts, currentEpoch)
		atts = []*ethpb.IndexedAttestation{
			createAtt(validatorIndices, 3 /* source */, 4 /* target */),
		}
		slasher.detectAttestationBatch(validatorChunkIdx, atts, currentEpoch)
		assert.LogsContain(t, hook, "Slashing found: SURROUNDED_VOTE")
	})
}

func createAtt(validatorIndices []uint64, source, target uint64) *ethpb.IndexedAttestation {
	return &ethpb.IndexedAttestation{
		AttestingIndices: validatorIndices,
		Data: &ethpb.AttestationData{
			Target: &ethpb.Checkpoint{
				Epoch: target,
			},
			Source: &ethpb.Checkpoint{
				Epoch: source,
			},
		},
	}
}
