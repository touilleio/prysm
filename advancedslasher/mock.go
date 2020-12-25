package main

import (
	"context"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"

	"github.com/prysmaticlabs/prysm/shared/event"
	"github.com/prysmaticlabs/prysm/shared/rand"
)

type MockFeeder struct {
	feed *event.Feed
}

func (m *MockFeeder) IndexedAttestationFeed() *event.Feed {
	return m.feed
}

func (m *MockFeeder) generateFakeAttestations(ctx context.Context) {
	gen := rand.NewGenerator()
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mockIndices := make([]uint64, 10)
			for i := 0; i < len(mockIndices); i++ {
				mockIndices[i] = uint64(gen.Intn(16384))
			}
			indexedAtt := &ethpb.IndexedAttestation{
				AttestingIndices: mockIndices,
				Data: &ethpb.AttestationData{
					Source: &ethpb.Checkpoint{
						Epoch: 0,
					},
					Target: &ethpb.Checkpoint{
						Epoch: 1,
					},
				},
			}
			m.feed.Send(indexedAtt)
		case <-ctx.Done():
			return
		}
	}
}
