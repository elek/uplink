package upload

import (
	"bytes"
	"context"
	"storj.io/common/pb"
	"storj.io/common/storj"
)

type PieceCache struct {
	buffer *bytes.Buffer
	output HashedPieceLayer
}

func (pc *PieceCache) HashCalculated(ctx context.Context, hash []byte) error {
	return pc.output.HashCalculated(ctx, hash)
}

//TODO: this caching is required to avoid to frequent signature. Can be avoided if we separated the signing and data sending.
func (pc *PieceCache) BeginPieceUpload(ctx context.Context, limit *pb.OrderLimit, address *pb.NodeAddress, privateKey storj.PiecePrivateKey) error {
	return pc.output.BeginPieceUpload(ctx, limit, address, privateKey)
}

func (pc *PieceCache) WritePieceUpload(ctx context.Context, data []byte) error {
	if pc.buffer.Len() < 262144 {
		pc.buffer.Write(data)
		return nil
	}
	err := pc.output.WritePieceUpload(ctx, pc.buffer.Bytes())
	pc.buffer.Reset()
	return err

}

func (pc *PieceCache) CommitPieceUpload(ctx context.Context) (*pb.PieceHash, error) {
	return pc.output.CommitPieceUpload(ctx)
}

var _ HashedPieceLayer = &PieceCache{}
