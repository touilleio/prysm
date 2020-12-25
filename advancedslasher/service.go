package main

import (
	"context"

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
	Feeder    SlashableDataFeeder
	SlasherDB db.Database
}

// Service --
type Slasher struct {
	ctx              context.Context
	config           *Config
	feeder           SlashableDataFeeder
	slasherDB        db.Database
	receivedAttsChan chan *ethpb.IndexedAttestation
	attQueue         *attestationQueue
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
	}, nil
}

// Start --
func (s *Slasher) Start() {
	go s.processAttestations()
	s.receiveAttestations(s.ctx)
}
