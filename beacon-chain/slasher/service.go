package slasher

import (
	"context"

	"github.com/prysmaticlabs/prysm/shared/event"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("prefix", "slasher")

// SlashableDataFeeder --
type SlashableDataFeeder interface {
	IndexedAttestationFeed() *event.Feed
}

type Service struct {
	feeder SlashableDataFeeder
}

type Config struct {
	Feeder SlashableDataFeeder
}

func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	return &Service{
		feeder: cfg.Feeder,
	}, nil
}

// Start --
func (s *Service) Start() {
	go s.detectIncomingAttestations()
}

// Stop --
func (s *Service) Stop() error {
	return nil
}

// Stop --
func (s *Service) Status() error {
	return nil
}
