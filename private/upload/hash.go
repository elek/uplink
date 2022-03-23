package upload

import (
	"context"
	"hash"
	"storj.io/common/pb"
	"storj.io/common/pkcrypto"
	"storj.io/common/storj"
)

type Hasher struct {
	Output HashedPieceLayer
	hasher hash.Hash
}

func (h *Hasher) BeginPieceUpload(ctx context.Context, limit *pb.OrderLimit, address *pb.NodeAddress, privateKey storj.PiecePrivateKey) error {
	h.hasher = pkcrypto.NewHash()
	return h.Output.BeginPieceUpload(ctx, limit, address, privateKey)
}

func (h *Hasher) WritePieceUpload(ctx context.Context, data []byte) error {
	h.hasher.Write(data)
	return h.Output.WritePieceUpload(ctx, data)

}

func (h *Hasher) CommitPieceUpload(ctx context.Context) (*pb.PieceHash, error) {
	//TODO: reuse the hash buffer here
	err := h.Output.HashCalculated(ctx, h.hasher.Sum([]byte{}))
	if err != nil {
		return nil, err
	}
	return h.Output.CommitPieceUpload(ctx)
}

type HashedPieceLayer interface {
	PieceLayer
	HashCalculated(ctx context.Context, hash []byte) error
}

var _ PieceLayer = &Hasher{}
