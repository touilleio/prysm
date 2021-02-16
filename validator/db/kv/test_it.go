package kv

import (
	"context"
	"sync"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	"github.com/prysmaticlabs/prysm/shared/timeutils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
)

const (
	mmapSize = 536870912 // 512Mb.
)

var pyrmontGenesisTime = time.Unix(1605700807, 0)

func TestIT(cliCtx *cli.Context) error {
	logrus.SetLevel(logrus.DebugLevel)
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
			_ = cancel
			log.WithField(
				"deadline", deadline,
			).Debug("Set deadline for attestations")

			var wg sync.WaitGroup
			for _, pubKey := range assignedKeys {
				wg.Add(1)
				go func(pk [48]byte) {
					defer wg.Done()
					attest(slotCtx, validatorDB, pk, slot)
				}(pubKey)
			}
			go func() {
				wg.Wait()
				log.WithFields(logrus.Fields{
					"slot":         slot,
					"slotInEpoch":  inEpoch,
					"epoch":        helpers.SlotToEpoch(slot),
					"numAttesters": len(assignedKeys),
				}).Info("Completed attest function for all assigned validators")
			}()
		}
	}
}

func attest(
	ctx context.Context,
	validatorDB *Store,
	pubKey [48]byte,
	slot types.Slot,
) {
	waitOneThird(ctx, slot)
	currentEpoch := helpers.SlotToEpoch(slot)
	source := currentEpoch - 1
	target := currentEpoch
	incomingAtt := createAttestation(slot, source, target)
	signingRoot, err := incomingAtt.Data.HashTreeRoot()
	if err != nil {
		log.WithError(err).Error("Could not compute signing root")
		return
	}
	slashable, err := validatorDB.CheckSlashableAttestation(ctx, pubKey, signingRoot, incomingAtt)
	if err != nil {
		log.WithError(err).Error("Could not check slashable attestation for public key")
		return
	}
	if slashable != NotSlashable {
		log.Warn("Attempted to produce slashable attestation")
		return
	}
	if err := validatorDB.SaveAttestationForPubKey(ctx, pubKey, signingRoot, incomingAtt); err != nil {
		log.WithError(err).Error("Could not save attestation for public key")
	}
	return
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

func waitOneThird(ctx context.Context, slot types.Slot) {
	delay := slotutil.DivideSlotBy(3 /* a third of the slot duration */)
	startTime := slotutil.SlotStartTime(uint64(pyrmontGenesisTime.Unix()), slot)
	finalTime := startTime.Add(delay)
	wait := timeutils.Until(finalTime)
	if wait <= 0 {
		return
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			return
		}
	}
}

func createAttestation(slot types.Slot, source, target types.Epoch) *ethpb.IndexedAttestation {
	return &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Slot:            slot,
			CommitteeIndex:  1,
			BeaconBlockRoot: make([]byte, 32),
			Source:          &ethpb.Checkpoint{Epoch: source, Root: make([]byte, 32)},
			Target:          &ethpb.Checkpoint{Epoch: target, Root: make([]byte, 32)},
		},
	}
}
