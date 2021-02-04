package kv

import (
	"context"
	"fmt"
	"sync"
	"time"

	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
)

func TestIT(cliCtx *cli.Context) error {
	dataDir := cliCtx.String(cmd.DataDirFlag.Name)
	ctx := context.Background()
	numValidators := 2000
	numEpochs := uint64(20000)
	pubKeys := make([][48]byte, numValidators)
	for i := 0; i < numValidators; i++ {
		var pk [48]byte
		copy(pk[:], fmt.Sprintf("%d", i))
		pubKeys[i] = pk
	}
	validatorDB, err := NewKVStore(ctx, dataDir, pubKeys)
	if err != nil {
		return err
	}
	// Every validator will have attested every (source, target) sequential pair
	// since genesis up to and including the weak subjectivity period epoch (54,000).
	fmt.Println("Writing attestting history for 2000 keys")
	err = validatorDB.update(func(tx *bolt.Tx) error {
		for i, pubKey := range pubKeys {
			bucket := tx.Bucket(pubKeysBucket)
			pkBucket, err := bucket.CreateBucketIfNotExists(pubKey[:])
			if err != nil {
				return err
			}
			sourceEpochsBucket, err := pkBucket.CreateBucketIfNotExists(attestationSourceEpochsBucket)
			if err != nil {
				return err
			}
			fmt.Println("Writing 20k epochs for pubkey", pubKey)
			for epoch := uint64(1); epoch < numEpochs; epoch++ {
				source := epoch - 1
				target := epoch
				sourceEpoch := bytesutil.Uint64ToBytesBigEndian(source)
				targetEpoch := bytesutil.Uint64ToBytesBigEndian(target)
				if err := sourceEpochsBucket.Put(sourceEpoch, targetEpoch); err != nil {
					return err
				}
			}
			fmt.Println("Done writing 20k epochs for pubkey", i)
		}
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Println("Done writing attestting history for 2000 keys")

	start := time.Now()
	fmt.Println("Starting to check slashable attestation for 2000 vals, 20k epochs history")
	var wg sync.WaitGroup
	wg.Add(numValidators)
	for _, pubKey := range pubKeys {
		go func(w *sync.WaitGroup, pk [48]byte) {
			defer w.Done()
			incomingAtt := createAttestation(numEpochs+1, numEpochs+2)
			_, err = validatorDB.CheckSlashableAttestation(ctx, pk, [32]byte{}, incomingAtt)
			if err != nil {
				panic(err)
			}
		}(&wg, pubKey)
	}
	wg.Wait()
	end := time.Now()
	fmt.Printf("Took %v to check\n", end.Sub(start))
	return nil
}

func createAttestation(source, target uint64) *ethpb.IndexedAttestation {
	return &ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: source},
			Target: &ethpb.Checkpoint{Epoch: target},
		},
	}
}
