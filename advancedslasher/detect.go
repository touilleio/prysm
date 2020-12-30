package main

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

func (s *Slasher) detectAttestationBatch(
	validatorChunkIdx uint64,
	batch []*ethpb.IndexedAttestation,
	currentEpoch uint64,
) {
	// Split the batch up into horizontal segments.
	// Map chunk indexes in the range `0..self.config.chunk_size` to attestations
	// for those chunks.
	attestationsForChunk := make(map[uint64][]*ethpb.IndexedAttestation)
	for _, att := range batch {
		chunkIdx := s.config.chunkIndex(att.Data.Source.Epoch)
		attestationsForChunk[chunkIdx] = append(attestationsForChunk[chunkIdx], att)
	}

	s.updateMaxArrays(validatorChunkIdx, attestationsForChunk, currentEpoch)
	// TODO: s.updateMinArrays(validatorChunkIdx, attestationsForChunk, currentEpoch)

	// Update all relevant validators for current epoch.
	// TODO: Complete.
	//
	//for validator_index in config.validator_indices_in_chunk(validator_chunk_index) {
	//	db.update_current_epoch_for_validator(validator_index, current_epoch, txn)?;
	//}
}
