package main

import ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"

func (s *Slasher) updateArrays(
	validatorChunkIdx uint64,
	chunkAttestations map[uint64][]*ethpb.IndexedAttestation,
	currentEpoch uint64,
) {
	updatedChunks := make(map[uint64][]byte)
	// Update the arrays for the change in epoch.
	//for _, validatorIdx := range s.config.validatorIndicesInChunk(validatorChunkIdx) {
	//	epoch_update_for_validator(
	//		db,
	//		txn,
	//		&mut updated_chunks,
	//		validator_chunk_index,
	//		validator_index,
	//		current_epoch,
	//		config,
	//	)
	//}

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

func (s *Slasher) applyAttestationForValidator(
	updateChunks map[uint64][]byte,
	validatorChunkIdx uint64,
	validatorIdx uint64,
	att *ethpb.IndexedAttestation,
	currentEpoch uint64,
) {

}
