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
				s.applyAttestationForValidatorMaxChunk(
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

// TODO: Handle for min chunks as well.
func (s *Slasher) applyAttestationForValidatorMaxChunk(
	updatedChunks map[uint64][]uint16,
	validatorChunkIdx uint64,
	validatorIdx uint64,
	att *ethpb.IndexedAttestation,
	currentEpoch uint64,
) {
	sourceEpoch := att.Data.Source.Epoch
	chunkIdx := s.config.chunkIndex(sourceEpoch)
	currentChunk := s.chunkForUpdate(updatedChunks, validatorChunkIdx, chunkIdx)

	// Check slashable, if so, return the slashing.
	slashable := s.checkSlashableChunk(currentChunk)
	if slashable {
		// TODO: Handle slashing appropriately.
		return
	}

	// Get the first start epoch for max chunk.
	var startEpoch uint64
	if sourceEpoch < currentEpoch {
		startEpoch = sourceEpoch + 1
	}

	// Update the chunks accordingly.
	for {
		chunkIdx = s.config.chunkIndex(startEpoch)
		currentChunk = s.chunkForUpdate(updatedChunks, validatorChunkIdx, chunkIdx)
		keepGoing := s.updateMaxChunk(
			currentChunk,
			chunkIdx,
			validatorIdx,
			startEpoch,
			att.Data.Target.Epoch,
			currentEpoch,
		)
		if !keepGoing {
			break
		}
		// Get the next start epoch for max chunk.
		// Move to first epoch of next chunk.
		startEpoch = (startEpoch/s.config.chunkSize + 1) * s.config.chunkSize
	}
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

func (s *Slasher) updateMaxChunk(
	chunk []uint16,
	chunkIdx,
	validatorIdx,
	startEpoch,
	newTargetEpoch,
	currentEpoch uint64,
) bool {
	epoch := startEpoch
	for s.config.chunkIndex(epoch) == chunkIdx && epoch <= currentEpoch {
		if newTargetEpoch > s.getChunkTarget(chunk, validatorIdx, epoch) {
			s.setChunkTarget(chunk, validatorIdx, epoch, newTargetEpoch)
		} else {
			return false
		}
		epoch += 1
	}
	// If the epoch to update now lies beyond the current chunk and is less than
	// or equal to the current epoch, then continue to the next chunk to update it.
	return epoch <= currentEpoch
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

func (s *Slasher) getChunkTarget(chunk []uint16, validatorIdx, epoch uint64) uint64 {
	if uint64(len(chunk)) != s.config.chunkSize*s.config.validatorChunkSize {
		panic("Not right length")
	}
	validatorOffset := s.config.validatorOffset(validatorIdx)
	chunkOffset := s.config.chunkOffset(epoch)
	cellIdx := s.config.cellIndex(validatorOffset, chunkOffset)
	if cellIdx >= uint64(len(chunk)) {
		panic("Cell index out of bounds (print cell index)")
	}
	distance := chunk[cellIdx]
	return epoch + uint64(distance)
}

func (s *Slasher) setChunkTarget(chunk []uint16, validatorIdx, epoch, targetEpoch uint64) {
	distance := s.epochDistance(targetEpoch, epoch)
	s.setRawDistance(chunk, validatorIdx, epoch, distance)
}

func (s *Slasher) epochDistance(epoch, baseEpoch uint64) uint16 {
	// TODO: Check safe math.
	distance := epoch - baseEpoch
	// TODO: Check max distance and throw error otherwise.
	return uint16(distance)
}
