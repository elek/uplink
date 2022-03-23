package upload

import (
	"context"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"storj.io/uplink/private/eestream"
)

type PieceLayer interface {
	BeginPieceUpload(ctx context.Context, limit *pb.OrderLimit, address *pb.NodeAddress, privateKey storj.PiecePrivateKey) error
	WritePieceUpload(ctx context.Context, data []byte) error
	CommitPieceUpload(ctx context.Context) (*pb.PieceHash, error)
}

type LongTailRouter struct {
	CreateOutput func() PieceLayer
	outputs      []PieceLayer
}

func (l *LongTailRouter) BeginSegment(ctx context.Context, piecePrivateKey storj.PiecePrivateKey, limits []*pb.AddressedOrderLimit, redundancyStrategy eestream.RedundancyStrategy) error {
	//at first time we create the outputs for each node
	if l.outputs == nil || len(l.outputs) == 0 {
		for i := 0; i < len(limits); i++ {
			l.outputs = append(l.outputs, l.CreateOutput())
		}
	}
	for ix, limit := range limits {
		//TODO this supposed to be done in async
		err := l.outputs[ix].BeginPieceUpload(ctx, limit.Limit, limit.StorageNodeAddress, piecePrivateKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LongTailRouter) CommitSegment(ctx context.Context) error {
	hashes := []*pb.PieceHash{}
	for _, output := range l.outputs {
		//TODO this supposed to be done in async

		hash, err := output.CommitPieceUpload(ctx)
		if err != nil {
			return err
		}
		//todo return the hashes to upper layers
		hashes = append(hashes, hash)
	}
	return nil
}

func (l *LongTailRouter) StartPieceUpload(ctx context.Context, ecShareIndex int, data []byte) error {
	//TODO: do this in async way
	//TODO: what if we hove not enough worker?
	return l.outputs[ecShareIndex].WritePieceUpload(ctx, data)
}

var _ ErasureEncodedLayer = &LongTailRouter{}
