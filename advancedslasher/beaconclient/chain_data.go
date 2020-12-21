package beaconclient

import (
	"context"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/params"
	"go.opencensus.io/trace"
)

var syncStatusPollingInterval = time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second

// ChainHead requests the latest beacon chain head
// from a beacon node via gRPC.
func (bs *Service) ChainHead(
	ctx context.Context,
) (*ethpb.ChainHead, error) {
	ctx, span := trace.StartSpan(ctx, "beaconclient.ChainHead")
	defer span.End()
	res, err := bs.beaconClient.GetChainHead(ctx, &ptypes.Empty{})
	if err != nil || res == nil {
		return nil, errors.Wrap(err, "Could not retrieve chain head or got nil chain head")
	}
	return res, nil
}

// GenesisValidatorsRoot requests or fetch from memory the beacon chain genesis
// validators root via gRPC.
func (bs *Service) GenesisValidatorsRoot(
	ctx context.Context,
) ([]byte, error) {
	ctx, span := trace.StartSpan(ctx, "beaconclient.GenesisValidatorsRoot")
	defer span.End()

	if bs.genesisValidatorRoot == nil {
		res, err := bs.nodeClient.GetGenesis(ctx, &ptypes.Empty{})
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve genesis data")
		}
		if res == nil {
			return nil, errors.Wrap(err, "nil genesis data")
		}
		bs.genesisValidatorRoot = res.GenesisValidatorsRoot
	}
	return bs.genesisValidatorRoot, nil
}
