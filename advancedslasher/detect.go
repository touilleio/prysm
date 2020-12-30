package main

import (
	"context"

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

	for _, validatorIdx := range s.config.validatorIndicesInChunk(validatorChunkIdx) {
		for _, att := range batch {
			err := s.slasherDB.SaveAttestationRecordForValidator(context.Background(), validatorIdx, att)
			if err != nil {
				panic(err)
			}
		}
	}

	slashings := s.updateChunks(validatorChunkIdx, attestationsForChunk, currentEpoch, minChunk)
	moreSlashings := s.updateChunks(validatorChunkIdx, attestationsForChunk, currentEpoch, maxChunk)

	totalSlashings := append(slashings, moreSlashings...)
	log.Infof("Total slashings found in batch: %d", len(totalSlashings))

	// Update all relevant validators for current epoch.
	for _, validatorIdx := range s.config.validatorIndicesInChunk(validatorChunkIdx) {
		err := s.slasherDB.UpdateLatestEpochWrittenForValidator(context.Background(), validatorIdx, currentEpoch)
		if err != nil {
			panic(err)
		}
	}
}

func (s *Slasher) updateChunks(
	validatorChunkIdx uint64,
	chunkAttestations map[uint64][]*ethpb.IndexedAttestation,
	currentEpoch uint64,
	kind chunkKind,
) []slashingKind {
	updatedChunks := make(map[uint64]Chunker)
	// Update the arrays for the change in epoch.
	for _, validatorIdx := range s.config.validatorIndicesInChunk(validatorChunkIdx) {
		s.epochUpdateForValidator(updatedChunks, validatorChunkIdx, validatorIdx, currentEpoch, kind)
	}

	slashings := make([]slashingKind, 0)
	for _, attestations := range chunkAttestations {
		for _, att := range attestations {
			for _, validatorIdx := range att.AttestingIndices {
				if validatorChunkIdx != s.config.validatorChunkIndex(validatorIdx) {
					continue
				}
				// TODO: Return slashing and handle.
				slashKind := s.applyAttestationForValidator(
					updatedChunks,
					validatorChunkIdx,
					validatorIdx,
					currentEpoch,
					att,
					kind,
				)
				slashings = append(slashings, slashKind)
			}
		}
	}

	// Store chunks on disk.
	for chunkIdx, chunk := range updatedChunks {
		key := s.config.diskKey(validatorChunkIdx, chunkIdx)
		err := s.slasherDB.SaveChunk(context.Background(), key, chunk.Chunk())
		if err != nil {
			panic(err)
		}
	}
	return slashings
}

func (s *Slasher) epochUpdateForValidator(
	updatedChunks map[uint64]Chunker,
	validatorChunkIdx,
	validatorIdx,
	currentEpoch uint64,
	kind chunkKind,
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
		currentChunk := s.chunkForUpdate(updatedChunks, validatorChunkIdx, chunkIdx, kind)
		for s.config.chunkIndex(epoch) == chunkIdx && epoch <= currentEpoch {
			setChunkRawDistance(
				currentChunk.Chunk(),
				validatorIdx,
				epoch,
				currentChunk.NeutralElement(),
				s.config,
			)
			epoch++
		}
		updatedChunks[chunkIdx] = currentChunk
	}
}

func (s *Slasher) applyAttestationForValidator(
	updatedChunks map[uint64]Chunker,
	validatorChunkIdx,
	validatorIdx,
	currentEpoch uint64,
	att *ethpb.IndexedAttestation,
	kind chunkKind,
) slashingKind {
	sourceEpoch := att.Data.Source.Epoch
	chunkIdx := s.config.chunkIndex(sourceEpoch)
	currentChunk := s.chunkForUpdate(updatedChunks, validatorChunkIdx, chunkIdx, kind)

	// Check slashable, if so, return the slashing.
	slashable, slashKind, err := currentChunk.CheckSlashable(s.slasherDB, validatorIdx, att)
	if err != nil {
		panic(err)
	}
	if slashable {
		// TODO: Handle slashing appropriately.
		log.Infof("Slashable with kind %v", slashKind)
		return slashKind
	}

	// Get the first start epoch for max chunk.
	// TODO: Check slashing status.
	startEpoch := currentChunk.FirstStartEpoch(sourceEpoch, currentEpoch)

	// Update the chunks accordingly.
	for {
		chunkIdx = s.config.chunkIndex(startEpoch)
		currentChunk = s.chunkForUpdate(updatedChunks, validatorChunkIdx, chunkIdx, kind)
		keepGoing, _ := currentChunk.Update(
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
		startEpoch = currentChunk.NextStartEpoch(startEpoch)
	}
	return notSlashable
}

func (s *Slasher) chunkForUpdate(
	updatedChunks map[uint64]Chunker,
	validatorChunkIndex,
	chunkIndex uint64,
	kind chunkKind,
) Chunker {
	if chunk, ok := updatedChunks[chunkIndex]; ok {
		return chunk
	}
	// Load from DB, if it does not exist, then create an empty chunk.
	key := s.config.diskKey(validatorChunkIndex, chunkIndex)
	data, exists, err := s.slasherDB.LoadChunk(context.Background(), key)
	if err != nil {
		panic(err)
	}
	var existingChunk Chunker
	if !exists {
		switch kind {
		case minChunk:
			existingChunk = NewMinChunk(s.config)
		case maxChunk:
			existingChunk = NewMaxChunker(s.config)
		}
	}
	switch kind {
	case minChunk:
		existingChunk, err = MinChunkFrom(s.config, data)
	case maxChunk:
		existingChunk, err = MaxChunkFrom(s.config, data)
	}
	if err != nil {
		panic(err)
	}
	updatedChunks[chunkIndex] = existingChunk
	return existingChunk
}
