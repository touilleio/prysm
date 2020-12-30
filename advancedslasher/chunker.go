package main

import (
	"context"
	"math"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"

	"github.com/prysmaticlabs/prysm/advancedslasher/db"
	"github.com/prysmaticlabs/prysm/shared/params"
)

var (
	_ = Chunker(&MinChunk{})
	_ = Chunker(&MaxChunk{})
)

type slashingKind int
type chunkKind int

const (
	minChunk chunkKind = iota
	maxChunk
)

const (
	notSlashable slashingKind = iota
	doubleVote
	surroundingVote
	surroundedVote
)

type Chunker interface {
	NeutralElement() uint16
	Chunk() []uint16
	CheckSlashable(
		slasherDB db.Database,
		validatorIdx uint64,
		attestation *ethpb.IndexedAttestation,
	) (bool, slashingKind, error)
	Update(
		chunkIdx,
		validatorIdx,
		startEpoch,
		newTargetEpoch,
		currentEpoch uint64,
	) (bool, error)
	FirstStartEpoch(sourceEpoch, currentEpoch uint64) uint64
	NextStartEpoch(startEpoch uint64) uint64
}

type MinChunk struct {
	config *Config
	data   []uint16
}

type MaxChunk struct {
	config *Config
	data   []uint16
}

func NewMinChunk(config *Config) *MinChunk {
	m := &MinChunk{
		config: config,
	}
	data := make([]uint16, config.chunkSize*config.validatorChunkSize)
	for i := 0; i < len(data); i++ {
		data[i] = m.NeutralElement()
	}
	m.data = data
	return m
}

func MinChunkFrom(config *Config, chunk []uint16) (*MinChunk, error) {
	if uint64(len(chunk)) != config.chunkSize*config.validatorChunkSize {
		return nil, errors.New("wrong size for min chunk")
	}
	return &MinChunk{
		config: config,
		data:   chunk,
	}, nil
}

func (m *MinChunk) NeutralElement() uint16 {
	return math.MaxUint16
}

func (m *MinChunk) CheckSlashable(
	slasherDB db.Database,
	validatorIdx uint64,
	attestation *ethpb.IndexedAttestation,
) (bool, slashingKind, error) {
	minTarget := getChunkTarget(m.data, validatorIdx, attestation.Data.Source.Epoch, m.config)
	if attestation.Data.Target.Epoch > minTarget {
		existingAttRecord, err := slasherDB.AttestationRecordForValidator(context.Background(), validatorIdx, minTarget)
		if err != nil {
			return false, notSlashable, err
		}
		if attestation.Data.Source.Epoch < existingAttRecord.SourceEpoch {
			return true, surroundingVote, nil
		}
		return true, doubleVote, nil
	}
	return false, notSlashable, nil
}

func (m *MinChunk) Chunk() []uint16 {
	return m.data
}

func (m *MinChunk) Update(chunkIdx, validatorIdx, startEpoch, newTargetEpoch, currentEpoch uint64) (bool, error) {
	// TODO: Make it a saturating sub.
	minEpoch := currentEpoch - (m.config.historyLength - 1)
	epoch := startEpoch
	for m.config.chunkIndex(epoch) == chunkIdx && epoch >= minEpoch {
		if newTargetEpoch < getChunkTarget(m.data, validatorIdx, epoch, m.config) {
			setChunkTarget(m.data, validatorIdx, epoch, newTargetEpoch, m.config)
		} else {
			// We can stop.
			return false, nil
		}
		epoch -= 1
	}
	return epoch >= minEpoch, nil
}

func (m *MinChunk) FirstStartEpoch(sourceEpoch, currentEpoch uint64) uint64 {
	if sourceEpoch > currentEpoch-m.config.historyLength {
		if sourceEpoch == 0 {
			panic("Cannot be 0")
		}
		return sourceEpoch - 1
	} else {
		return params.BeaconConfig().FarFutureEpoch
	}
}

// Move to last epoch of previous chunk
func (m *MinChunk) NextStartEpoch(startEpoch uint64) uint64 {
	return startEpoch/m.config.chunkSize*m.config.chunkSize - 1
}

// Max chunker methods.
func NewMaxChunker(config *Config) *MaxChunk {
	m := &MaxChunk{
		config: config,
	}
	data := make([]uint16, config.chunkSize*config.validatorChunkSize)
	for i := 0; i < len(data); i++ {
		data[i] = m.NeutralElement()
	}
	m.data = data
	return m
}

func MaxChunkFrom(config *Config, chunk []uint16) (*MaxChunk, error) {
	if uint64(len(chunk)) != config.chunkSize*config.validatorChunkSize {
		return nil, errors.New("wrong size for max chunk")
	}
	return &MaxChunk{
		config: config,
		data:   chunk,
	}, nil
}

func (m *MaxChunk) NeutralElement() uint16 {
	return 0
}

func (m *MaxChunk) CheckSlashable(
	slasherDB db.Database,
	validatorIdx uint64,
	attestation *ethpb.IndexedAttestation,
) (bool, slashingKind, error) {
	// TODO: Use right context.
	maxTarget := getChunkTarget(m.data, validatorIdx, attestation.Data.Source.Epoch, m.config)
	if attestation.Data.Target.Epoch < maxTarget {
		existingAttRecord, err := slasherDB.AttestationRecordForValidator(context.Background(), validatorIdx, maxTarget)
		if err != nil {
			return false, notSlashable, err
		}
		if existingAttRecord.SourceEpoch < attestation.Data.Source.Epoch {
			return true, surroundedVote, nil
		}
		return true, doubleVote, nil
	}
	return false, notSlashable, nil
}

func (m *MaxChunk) Chunk() []uint16 {
	return m.data
}

func (m *MaxChunk) Update(chunkIdx, validatorIdx, startEpoch, newTargetEpoch, currentEpoch uint64) (bool, error) {
	epoch := startEpoch
	for m.config.chunkIndex(epoch) == chunkIdx && epoch <= currentEpoch {
		if newTargetEpoch > getChunkTarget(m.data, validatorIdx, epoch, m.config) {
			setChunkTarget(m.data, validatorIdx, epoch, newTargetEpoch, m.config)
		} else {
			return false, nil
		}
		epoch += 1
	}
	// If the epoch to update now lies beyond the current chunk and is less than
	// or equal to the current epoch, then continue to the next chunk to update it.
	return epoch <= currentEpoch, nil
}

func (m *MaxChunk) FirstStartEpoch(sourceEpoch, currentEpoch uint64) uint64 {
	if sourceEpoch < currentEpoch {
		return sourceEpoch + 1
	} else {
		return params.BeaconConfig().FarFutureEpoch
	}
}

// Move to first epoch of next chunk.
func (m *MaxChunk) NextStartEpoch(startEpoch uint64) uint64 {
	return (startEpoch/m.config.chunkSize + 1) * m.config.chunkSize
}

func setChunkRawDistance(
	chunk []uint16,
	validatorIdx uint64,
	epoch uint64,
	targetDistance uint16,
	config *Config,
) {
	validatorOffset := config.validatorOffset(validatorIdx)
	chunkOffset := config.chunkOffset(epoch)
	cellIdx := config.cellIndex(validatorOffset, chunkOffset)
	if cellIdx >= uint64(len(chunk)) {
		panic("Cell index out of bounds (print cell index)")
	}
	chunk[cellIdx] = targetDistance
}

func getChunkTarget(chunk []uint16, validatorIdx, epoch uint64, config *Config) uint64 {
	if uint64(len(chunk)) != config.chunkSize*config.validatorChunkSize {
		panic("Not right length")
	}
	validatorOffset := config.validatorOffset(validatorIdx)
	chunkOffset := config.chunkOffset(epoch)
	cellIdx := config.cellIndex(validatorOffset, chunkOffset)
	if cellIdx >= uint64(len(chunk)) {
		panic("Cell index out of bounds (print cell index)")
	}
	distance := chunk[cellIdx]
	return epoch + uint64(distance)
}

func setChunkTarget(chunk []uint16, validatorIdx, epoch, targetEpoch uint64, config *Config) {
	distance := epochDistance(targetEpoch, epoch)
	setChunkRawDistance(chunk, validatorIdx, epoch, distance, config)
}

func epochDistance(epoch, baseEpoch uint64) uint16 {
	// TODO: Check safe math.
	distance := epoch - baseEpoch
	// TODO: Check max distance and throw error otherwise.
	return uint16(distance)
}
