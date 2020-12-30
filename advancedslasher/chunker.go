package main

var (
	_ = Chunker(&MinChunker{})
	_ = Chunker(&MaxChunker{})
)

type Chunker interface {
	NeutralElement() uint16
	Update(
		chunkIdx,
		validatorIdx,
		startEpoch,
		newTargetEpoch,
		currentEpoch uint64,
	) (bool, error)
	FirstStartEpoch(sourceEpoch, currentEpoch uint64) uint64
	NextStartEpoch(startEpoch uint64) uint64
	Load(validatorChunkIdx, chunkIdx uint64) error
	Store(validatorChunkIdx, chunkIdx uint64) error
}

type MinChunker struct {
}

type MaxChunker struct {
}

func (m *MinChunker) NeutralElement() uint16 {
	panic("implement me")
}

func (ch *MinChunker) Update(chunkIdx, validatorIdx, startEpoch, newTargetEpoch, currentEpoch uint64) (bool, error) {
	panic("implement me")
}

func (ch *MinChunker) FirstStartEpoch(sourceEpoch, currentEpoch uint64) uint64 {
	panic("implement me")
}

func (ch *MinChunker) NextStartEpoch(startEpoch uint64) uint64 {
	panic("implement me")
}

func (ch *MinChunker) Load(validatorChunkIdx, chunkIdx uint64) error {
	panic("implement me")
}

func (ch *MinChunker) Store(validatorChunkIdx, chunkIdx uint64) error {
	panic("implement me")
}

func (m *MaxChunker) NeutralElement() uint16 {
	panic("implement me")
}

func (m *MaxChunker) Update(chunkIdx, validatorIdx, startEpoch, newTargetEpoch, currentEpoch uint64) (bool, error) {
	panic("implement me")
}

func (m *MaxChunker) FirstStartEpoch(sourceEpoch, currentEpoch uint64) uint64 {
	panic("implement me")
}

func (m *MaxChunker) NextStartEpoch(startEpoch uint64) uint64 {
	panic("implement me")
}

func (m *MaxChunker) Load(validatorChunkIdx, chunkIdx uint64) error {
	panic("implement me")
}

func (m *MaxChunker) Store(validatorChunkIdx, chunkIdx uint64) error {
	panic("implement me")
}
