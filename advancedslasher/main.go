package main

import (
	"context"
	"os"
	"time"

	"github.com/prysmaticlabs/prysm/advancedslasher/db/kv"
	"github.com/prysmaticlabs/prysm/shared/event"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"runtime"
)

var log = logrus.WithField("prefix", "slasher")

func main() {
	formatter := new(prefixed.TextFormatter)
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slasherDB, err := kv.NewKVStore("/tmp/advancedslasher", &kv.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll("/tmp/advancedslasher"); err != nil {
			panic(err)
		}
	}()

	genesisTime := time.Now()
	slotsPerEpoch := uint64(2)
	secondsPerSlot := uint64(4)
	log.Infof("Genesis time reached %v", genesisTime)

	mockFeeder := &MockFeeder{
		feed:             new(event.Feed),
		validatorIndices: []uint64{1},
		genesisTime:      genesisTime,
		slotsPerEpoch:    slotsPerEpoch,
		secondsPerSlot:   secondsPerSlot,
	}
	go mockFeeder.generateFakeAttestations(ctx)
	slasher, err := NewSlasher(ctx, &ServiceConfig{
		Feeder:         mockFeeder,
		SlasherDB:      slasherDB,
		GenesisTime:    genesisTime,
		SlotsPerEpoch:  slotsPerEpoch,
		SecondsPerSlot: secondsPerSlot,
	})
	if err != nil {
		log.Fatal(err)
	}
	slasher.Start()
}
