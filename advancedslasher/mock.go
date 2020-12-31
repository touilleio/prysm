package main

import (
	"context"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"

	"github.com/prysmaticlabs/prysm/shared/event"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
)

type MockFeeder struct {
	feed             *event.Feed
	genesisTime      time.Time
	validatorIndices []uint64
	secondsPerSlot   uint64
	slotsPerEpoch    uint64
}

func (m *MockFeeder) IndexedAttestationFeed() *event.Feed {
	return m.feed
}

func (m *MockFeeder) generateFakeAttestations(ctx context.Context) {
	ticker := slotutil.GetEpochTicker(m.genesisTime, m.secondsPerSlot*m.slotsPerEpoch)
	defer ticker.Done()
	for {
		select {
		case currentEpoch := <-ticker.C():
			sourceEpoch := currentEpoch
			targetEpoch := currentEpoch
			if sourceEpoch != 0 {
				sourceEpoch -= 1
			}
			indexedAtt := &ethpb.IndexedAttestation{
				AttestingIndices: m.validatorIndices,
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{
						Epoch: sourceEpoch,
					},
					Target: &ethpb.Checkpoint{
						Epoch: targetEpoch,
					},
				},
			}
			if currentEpoch > 0 && currentEpoch%4 == 0 {
				log.Infof("Simulating surround vote at epoch %d", currentEpoch)
				// Create a surround vote.
				indexedAtt.Data.Source.Epoch -= 2
			}
			log.Warnf("Submitting att with source %d, target %d", indexedAtt.Data.Source.Epoch, targetEpoch)
			m.feed.Send(indexedAtt)
		case <-ctx.Done():
			return
		}
	}
}
