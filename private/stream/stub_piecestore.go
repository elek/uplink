package stream

import (
	"context"
	"fmt"
	"storj.io/common/pb"
	"storj.io/common/signing"
	"storj.io/common/storj"
	"storj.io/drpc"
	"sync"
	"time"
)

type pieceStoreStub struct {
	node   *nodeStub
	closed chan struct{}
	once   sync.Once
	key    storj.PiecePrivateKey
}

func NewPieceStoreStub(node *nodeStub) *pieceStoreStub {
	_, key, err := storj.NewPieceKey()
	if err != nil {
		panic(err)
	}

	return &pieceStoreStub{
		key:    key,
		node:   node,
		closed: make(chan struct{}),
	}
}

func (p pieceStoreStub) Close() error {
	p.once.Do(func() {
		close(p.closed)
	})
	return nil
}

func (p pieceStoreStub) Closed() <-chan struct{} {
	return p.closed
}

func (p pieceStoreStub) Invoke(ctx context.Context, rpc string, enc drpc.Encoding, in, out drpc.Message) error {
	//TODO implement me
	panic("implement me")
}

func (p pieceStoreStub) NewStream(ctx context.Context, rpc string, enc drpc.Encoding) (drpc.Stream, error) {
	return &pieceStoreStream{
		key:      p.key,
		rpc:      rpc,
		ctx:      ctx,
		node:     p.node,
		requests: make(chan drpc.Message, 1000),
	}, nil
}

type pieceStoreStream struct {
	ctx      context.Context
	node     *nodeStub
	rpc      string
	requests chan drpc.Message
	key      storj.PiecePrivateKey
}

func (p pieceStoreStream) Context() context.Context {
	return p.ctx
}

func (p pieceStoreStream) MsgSend(msg drpc.Message, enc drpc.Encoding) error {
	switch m := msg.(type) {
	case *pb.PieceUploadRequest:
		if m.Done == nil {
			return nil
		}
	default:
		panic(fmt.Sprintf("%T is not supported", m))
	}
	p.requests <- msg
	return nil
}

func (p pieceStoreStream) MsgRecv(msg drpc.Message, enc drpc.Encoding) error {
	request := <-p.requests

	switch m := request.(type) {
	case *pb.PieceUploadRequest:
		response := msg.(*pb.PieceUploadResponse)
		if m.Done != nil {
			signer := signing.SignerFromFullIdentity(p.node.Identity)
			m.Done.Timestamp = time.Now()
			hash, err := signing.SignPieceHash(p.ctx, signer, m.Done)
			if err != nil {
				return err
			}
			response.Done = hash
		}
	default:
		panic(fmt.Sprintf("%T is not supported", m))
	}
	return nil
}

func (p pieceStoreStream) CloseSend() error {
	fmt.Println("Close Send " + p.node.Address)
	close(p.requests)
	return nil
}

func (p pieceStoreStream) Close() error {
	fmt.Println("Close")
	close(p.requests)
	return nil
}

var _ drpc.Conn = &pieceStoreStub{}
