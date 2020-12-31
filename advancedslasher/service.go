package main

import (
	"context"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/advancedslasher/db"
	"github.com/prysmaticlabs/prysm/shared/event"
)

// SlashableDataFeeder --
type SlashableDataFeeder interface {
	IndexedAttestationFeed() *event.Feed
}

// Config --
type ServiceConfig struct {
	Feeder         SlashableDataFeeder
	SlasherDB      db.Database
	GenesisTime    time.Time
	SecondsPerSlot uint64
	SlotsPerEpoch  uint64
}

// Service --
type Slasher struct {
	ctx              context.Context
	config           *Config
	feeder           SlashableDataFeeder
	slasherDB        db.Database
	receivedAttsChan chan *ethpb.IndexedAttestation
	attQueue         *attestationQueue
	genesisTime      time.Time
	secondsPerSlot   uint64
	slotsPerEpoch    uint64
}

// NewService --
func NewSlasher(ctx context.Context, cfg *ServiceConfig) (*Slasher, error) {
	return &Slasher{
		ctx: ctx,
		config: &Config{
			chunkSize:          DEFAULT_CHUNK_SIZE,
			validatorChunkSize: DEFAULT_VALIDATOR_CHUNK_SIZE,
			historyLength:      DEFAULT_HISTORY_LENGTH,
			updatePeriod:       DEFAULT_UPDATE_PERIOD,
		},
		feeder:           cfg.Feeder,
		slasherDB:        cfg.SlasherDB,
		receivedAttsChan: make(chan *ethpb.IndexedAttestation, 1),
		attQueue:         newAttestationQueue(),
		genesisTime:      cfg.GenesisTime,
		secondsPerSlot:   cfg.SecondsPerSlot,
		slotsPerEpoch:    cfg.SlotsPerEpoch,
	}, nil
}

// Start --
func (s *Slasher) Start() {
	go s.processQueuedAttestations(s.ctx)
	s.receiveAttestations(s.ctx)
}
