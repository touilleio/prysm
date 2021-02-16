package kv

import (
	"context"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/featureconfig"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	"github.com/prysmaticlabs/prysm/shared/timeutils"
	"github.com/prysmaticlabs/prysm/shared/traceutil"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

const (
	mmapSize = 536870912 // 512Mb.
)

var pyrmontGenesisTime = time.Unix(1605700807, 0)

func TestIT(cliCtx *cli.Context) error {
	dataDir := cliCtx.String(cmd.DataDirFlag.Name)
	ctx := context.Background()
	validatorDB, err := NewKVStore(ctx, dataDir, &Config{
		InitialMMapSize: mmapSize,
	})
	if err != nil {
		return err
	}
	publicKeys, err := fetchAttestedPublicKeys(validatorDB)
	if err != nil {
		return err
	}
	log.WithField(
		"numKeys", len(publicKeys),
	).Info("Starting pyrmont simulation for public keys")

	duties := assignDuties(publicKeys)
	slotTicker := slotutil.NewSlotTicker(pyrmontGenesisTime, params.BeaconConfig().SecondsPerSlot)
	for {
		select {
		case slot := <-slotTicker.C():
			inEpoch := slotInEpoch(slot)
			assignedKeys, ok := duties[inEpoch]
			if !ok {
				log.Warn("No assigned keys")
				continue
			}
			log.WithFields(logrus.Fields{
				"slot":         slot,
				"slotInEpoch":  inEpoch,
				"epoch":        helpers.SlotToEpoch(slot),
				"numAttesters": len(assignedKeys),
			}).Info("Slot reached")

			deadline := slotDeadline(slot)
			slotCtx, cancel := context.WithDeadline(ctx, deadline)
			log.WithField(
				"deadline", deadline,
			).Debug("Set deadline for attestations")
			go attest()
		}
	}

	//for _, pubKey := range publicKeys {
	//	log.Infof("Public key %#x", pubKey)
	//}

	//start := time.Now()
	//fmt.Println("Starting to check slashable attestation for 2000 vals, 20k epochs history")
	//var wg sync.WaitGroup
	//wg.Add(numValidators)
	//for _, pubKey := range pubKeys {
	//	go func(w *sync.WaitGroup, pk [48]byte) {
	//		defer w.Done()
	//		incomingAtt := createAttestation(numEpochs+1, numEpochs+2)
	//		_, err = validatorDB.CheckSlashableAttestation(ctx, pk, [32]byte{}, incomingAtt)
	//		if err != nil {
	//			panic(err)
	//		}
	//	}(&wg, pubKey)
	//}
	//wg.Wait()
	//end := time.Now()
	//fmt.Printf("Took %v to check\n", end.Sub(start))
	//return nil
}

func attest(validatorDB *Store) {
	incomingAtt := createAttestation(numEpochs+1, numEpochs+2)
	_, err = validatorDB.CheckSlashableAttestation(ctx, pk, [32]byte{}, incomingAtt)
	if err != nil {
		panic(err)
	}
}

func fetchAttestedPublicKeys(validatorDB *Store) ([][48]byte, error) {
	pubKeys := make([][48]byte, 0)
	err := validatorDB.view(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(pubKeysBucket)
		return bucket.ForEach(func(k, _ []byte) error {
			var pk [48]byte
			copy(pk[:], k)
			pubKeys = append(pubKeys, pk)
			return nil
		})
	})
	return pubKeys, err
}

func assignDuties(publicKeys [][48]byte) map[types.Slot][][48]byte {
	validatorsPerSlot := uint64(len(publicKeys)) / uint64(params.BeaconConfig().SlotsPerEpoch)
	duties := make(map[types.Slot][][48]byte)
	for i := types.Slot(0); i < params.BeaconConfig().SlotsPerEpoch; i++ {
		lowOffset := uint64(i) * validatorsPerSlot
		highOffset := lowOffset + validatorsPerSlot
		if i+1 == params.BeaconConfig().SlotsPerEpoch {
			duties[i] = publicKeys[lowOffset:]
		} else {
			duties[i] = publicKeys[lowOffset:highOffset]
		}
	}
	return duties
}

func slotInEpoch(slot types.Slot) types.Slot {
	return slot.ModSlot(params.BeaconConfig().SlotsPerEpoch)
}

func slotDeadline(slot types.Slot) time.Time {
	secs := time.Duration((slot + 1).Mul(params.BeaconConfig().SecondsPerSlot))
	return pyrmontGenesisTime.Add(secs * time.Second)
}

func (v *validator) waitOneThirdOrValidBlock(ctx context.Context, slot types.Slot) {
	ctx, span := trace.StartSpan(ctx, "validator.waitOneThirdOrValidBlock")
	defer span.End()

	// Don't need to wait if requested slot is the same as highest valid slot.
	if slot <= v.highestValidSlot {
		return
	}

	delay := slotutil.DivideSlotBy(3 /* a third of the slot duration */)
	startTime := slotutil.SlotStartTime(v.genesisTime, slot)
	finalTime := startTime.Add(delay)
	wait := timeutils.Until(finalTime)
	if wait <= 0 {
		return
	}
	t := time.NewTimer(wait)
	defer t.Stop()

	bChannel := make(chan *ethpb.SignedBeaconBlock, 1)
	sub := v.blockFeed.Subscribe(bChannel)
	defer sub.Unsubscribe()

	for {
		select {
		case b := <-bChannel:
			if featureconfig.Get().AttestTimely {
				if slot <= b.Block.Slot {
					return
				}
			}
		case <-ctx.Done():
			traceutil.AnnotateError(span, ctx.Err())
			return
		case <-sub.Err():
			log.Error("Subscriber closed, exiting goroutine")
			return
		case <-t.C:
			return
		}
	}
}

func createAttestation(source, target types.Epoch) *ethpb.IndexedAttestation {
	return &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: source},
			Target: &ethpb.Checkpoint{Epoch: target},
		},
	}
}
