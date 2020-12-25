package main

import (
	"context"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
)

func (s *Slasher) receiveAttestations(ctx context.Context) {
	sub := s.feeder.IndexedAttestationFeed().Subscribe(s.receivedAttsChan)
	defer close(s.receivedAttsChan)
	defer sub.Unsubscribe()
	for {
		select {
		case att := <-s.receivedAttsChan:
			log.Infof("Got attestation with indices %v", att.AttestingIndices)
			s.attQueue.push(att)
		case <-sub.Err():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Slasher) processQueuedAttestations(ctx context.Context) {
	ticker := slotutil.GetEpochTicker(time.Now(), uint64(2) /* seconds per slot */)
	defer ticker.Done()
	for {
		select {
		case currentEpoch := <-ticker.C():
			atts := s.attQueue.dequeue()
			if !validateAttestations(atts) {
				// TODO: Defer is ready at a future time.
				continue
			}
			// TODO: Store indexed attestation into database.

			// Group by validator index and process batches.
			// TODO: Perform concurrently with wait groups...?
			groupedAtts := s.groupByValidatorIndex(atts)
			for subqueueIdx, atts := range groupedAtts {
				s.detectAttestationBatch(subqueueIdx, atts, currentEpoch)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Slasher) groupByValidatorIndex(attestations []*ethpb.IndexedAttestation) map[uint64][]*ethpb.IndexedAttestation {
	groupedAttestations := make(map[uint64][]*ethpb.IndexedAttestation)
	for _, att := range attestations {
		subqueueIndices := make(map[uint64]bool)
		for _, validatorIdx := range att.AttestingIndices {
			chunkIdx := s.config.validatorChunkIndex(validatorIdx)
			subqueueIndices[chunkIdx] = true
		}
		for subqueueIdx := range subqueueIndices {
			groupedAttestations[subqueueIdx] = append(
				groupedAttestations[subqueueIdx],
				att,
			)
		}
	}
	return groupedAttestations
}

func validateAttestations(att []*ethpb.IndexedAttestation) bool {
	return true
}
