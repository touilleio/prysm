package main

const (
	DEFAULT_CHUNK_SIZE           uint64 = 16
	DEFAULT_VALIDATOR_CHUNK_SIZE uint64 = 256
	DEFAULT_HISTORY_LENGTH       uint64 = 4096
	DEFAULT_UPDATE_PERIOD        uint64 = 12
	DEFAULT_MAX_DB_SIZE          uint64 = 256 * 1024 // 256 GiB
	DEFAULT_BROADCAST            bool   = false
)

type Config struct {
	chunkSize          uint64
	validatorChunkSize uint64
	historyLength      uint64
	updatePeriod       uint64
}

func (c *Config) chunkIndex(epoch uint64) uint64 {
	return (epoch % c.historyLength) / c.chunkSize
}

func (c *Config) validatorChunkIndex(validatorIndex uint64) uint64 {
	return validatorIndex / c.validatorChunkSize
}

func (c *Config) chunkOffset(epoch uint64) uint64 {
	return epoch % c.chunkSize
}

func (c *Config) validatorOffset(validatorIndex uint64) uint64 {
	return validatorIndex % c.validatorChunkSize
}

/// Map the validator and epoch chunk indexes into a single value for use as a database key.
func (c *Config) diskKey(validatorChunkIndex uint64, chunkIndex uint64) uint64 {
	width := c.historyLength / c.chunkSize
	return validatorChunkIndex*width + chunkIndex
}

/// Map the validator and epoch offsets into an index for chunk data.
func (c *Config) cellIndex(validatorOffset uint64, chunkOffset uint64) uint64 {
	return validatorOffset*c.chunkSize + chunkOffset
}

func (c *Config) validatorIndicesInChunk(validatorChunkIdx uint64) []uint64 {
	validatorIndices := make([]uint64, 0)
	low := validatorChunkIdx * c.validatorChunkSize
	high := (validatorChunkIdx + 1) * c.validatorChunkSize
	for i := low; i < high; i++ {
		validatorIndices = append(validatorIndices, i)
	}
	return validatorIndices
}
