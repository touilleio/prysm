package main

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

func (s *Slasher) updateMaxArrays(
	validatorChunkIdx uint64,
	chunkAttestations map[uint64][]*ethpb.IndexedAttestation,
	currentEpoch uint64,
) {
	updatedChunks := make(map[uint64][]uint16)
	// Update the arrays for the change in epoch.
	for _, validatorIdx := range s.config.validatorIndicesInChunk(validatorChunkIdx) {
		s.epochUpdateForValidator(updatedChunks, validatorChunkIdx, validatorIdx, currentEpoch)
	}

	for _, attestations := range chunkAttestations {
		for _, att := range attestations {
			for _, validatorIdx := range att.AttestingIndices {
				if validatorChunkIdx != s.config.validatorChunkIndex(validatorIdx) {
					continue
				}
				s.applyAttestationForValidator(
					updatedChunks,
					validatorChunkIdx,
					validatorIdx,
					att,
					currentEpoch,
				)
			}
		}
	}

	// Store chunks on disk.
	//metrics::inc_counter_vec_by(
	//	&SLASHER_NUM_CHUNKS_UPDATED,
	//	&[T::name()],
	//updated_chunks.len() as u64,
	//);

	//for (chunk_index, chunk) in updated_chunks {
	//	//chunk.store(db, txn, validator_chunk_index, chunk_index, config)?;
	//}
	// return slashings
}

func (s *Slasher) epochUpdateForValidator(
	updatedChunks map[uint64][]uint16,
	validatorChunkIdx uint64,
	validatorIdx uint64,
	currentEpoch uint64,
) {
	prevEpochWritten, exists, err := s.slasherDB.LatestEpochWrittenForValidator(s.ctx, validatorIdx)
	if err != nil {
		panic(err)
	}
	epoch := prevEpochWritten
	if !exists {
		epoch = 0
	}
	for epoch <= currentEpoch {
		chunkIdx := s.config.chunkIndex(epoch)
		currentChunk := s.chunkForUpdate(updatedChunks, validatorChunkIdx, chunkIdx)
		for s.config.chunkIndex(epoch) == chunkIdx && epoch <= currentEpoch {
			currentChunk = s.setRawDistance(
				currentChunk,
				validatorIdx,
				epoch,
				0, /* natural max element */
			)
			epoch++
		}
		updatedChunks[chunkIdx] = currentChunk
	}
}

func (s *Slasher) applyAttestationForValidator(
	updatedChunks map[uint64][]uint16,
	validatorChunkIdx uint64,
	validatorIdx uint64,
	att *ethpb.IndexedAttestation,
	currentEpoch uint64,
) {
	chunkIdx := s.config.chunkIndex(att.Data.Source.Epoch)
	currentChunk := s.chunkForUpdate(updatedChunks, validatorChunkIdx, chunkIdx)

	// Check slashable, if so, return the slashing.
	slashable := s.checkSlashableChunk(currentChunk)
	if slashable {
		// TODO: Handle slashing appropriately.
		return
	}

	// Update the chunks accordingly.
}

func (s *Slasher) chunkForUpdate(
	updatedChunks map[uint64][]uint16,
	validatorChunkIndex uint64,
	chunkIndex uint64,
) []uint16 {
	chunk, ok := updatedChunks[chunkIndex]
	if ok {
		return chunk
	}
	// Load from DB, if it does not exist, then create an empty chunk.
	return []uint16{}
}

func (s *Slasher) checkSlashableChunk(chunk []uint16) bool {
	return false
}

func (s *Slasher) setRawDistance(
	chunk []uint16,
	validatorIdx uint64,
	epoch uint64,
	targetDistance uint16,
) []uint16 {
	validatorOffset := s.config.validatorOffset(validatorIdx)
	chunkOffset := s.config.chunkOffset(epoch)
	cellIdx := s.config.cellIndex(validatorOffset, chunkOffset)
	chunk[cellIdx] = targetDistance
	return chunk
}
