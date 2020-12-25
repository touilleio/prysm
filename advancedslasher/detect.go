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
}

func (s *Slasher) processBatch() {
}

func (s *Slasher) updateMaxTargetChunk(validatorIdx uint64, chunk []byte) {
	//let mut epoch = start_epoch;
	//while config.chunk_index(epoch) == chunk_index && epoch <= current_epoch {
	//	if new_target_epoch > self.chunk.get_target(validator_index, epoch, config)? {
	//	self.chunk
	//	.set_target(validator_index, epoch, new_target_epoch, config)?;
	//} else {
	//	// We can stop.
	//	return Ok(false);
	//}
	//	epoch += 1;
	//}
	//// If the epoch to update now lies beyond the current chunk and is less than
	//// or equal to the current epoch, then continue to the next chunk to update it.
	//Ok(epoch <= current_epoch)
}
