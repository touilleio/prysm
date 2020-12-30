package main

import (
	"context"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"

	"github.com/prysmaticlabs/prysm/shared/event"
)

type MockFeeder struct {
	feed             *event.Feed
	validatorIndices []uint64
}

func (m *MockFeeder) IndexedAttestationFeed() *event.Feed {
	return m.feed
}

func (m *MockFeeder) generateFakeAttestations(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	currentEpoch := uint64(1)
	for {
		select {
		case <-ticker.C:
			indexedAtt := &ethpb.IndexedAttestation{
				AttestingIndices: m.validatorIndices,
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{
						Epoch: currentEpoch - 1,
					},
					Target: &ethpb.Checkpoint{
						Epoch: currentEpoch,
					},
				},
			}
			if currentEpoch%10 == 0 {
				log.Infof("Simulating surround vote at epoch %d", currentEpoch)
				// Create a surround vote.
				indexedAtt.Data.Source.Epoch -= 1
			}
			m.feed.Send(indexedAtt)
			currentEpoch++
		case <-ctx.Done():
			return
		}
	}
}
