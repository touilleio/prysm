package main

import (
	"context"
	"testing"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/event"
)

type mockFeeder struct{}

func (m *mockFeeder) IndexedAttestationFeed() *event.Feed {
	return new(event.Feed)
}

func TestSlasher_receiveAttestations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Slasher{
		receivedAttsChan: make(chan *ethpb.IndexedAttestation, 0),
		feeder:           &mockFeeder{},
	}
	exitChan := make(chan struct{})
	go func() {
		s.receiveAttestations(ctx)
		exitChan <- struct{}{}
	}()
	s.receivedAttsChan <- &ethpb.IndexedAttestation{
		AttestingIndices: []uint64{1, 2, 3},
	}
	s.receivedAttsChan <- &ethpb.IndexedAttestation{
		AttestingIndices: []uint64{4, 5, 6},
	}
	cancel()
	<-exitChan
}
