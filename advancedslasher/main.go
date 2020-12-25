package main

import (
	"context"
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
	mockFeeder := &MockFeeder{feed: new(event.Feed)}
	go mockFeeder.generateFakeAttestations(ctx)
	slasher, err := NewSlasher(ctx, &ServiceConfig{
		Feeder:    mockFeeder,
		SlasherDB: slasherDB,
	})
	if err != nil {
		log.Fatal(err)
	}
	slasher.Start()
}
