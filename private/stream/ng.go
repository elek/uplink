package stream

import (
	"context"
	"fmt"
	"github.com/vivint/infectious"
	"github.com/zeebo/errs"
	"io"
	"storj.io/common/context2"
	"storj.io/common/encryption"
	"storj.io/common/paths"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"storj.io/uplink/private/eestream"
	"storj.io/uplink/private/metaclient"
	"storj.io/uplink/private/piecestore"
	"time"
)

type ObjectWriter struct {
	encryptionStore      *encryption.Store
	bucket               string
	key                  string
	expiration           time.Time
	encryptionParameters storj.EncryptionParameters
}

func NewObjectWriter() (*ObjectWriter, error) {
	return &ObjectWriter{}, nil
}

func (o *ObjectWriter) Write(data []byte) error {
	derivedKey, err := encryption.DeriveContentKey(o.bucket, paths.NewUnencrypted(o.key), o.encryptionStore)
	if err != nil {
		return errs.Wrap(err)
	}
	encPath, err := encryption.EncryptPathWithStoreCipher(o.bucket, paths.NewUnencrypted(o.key), o.encryptionStore)
	if err != nil {
		return errs.Wrap(err)
	}

	beginObjectReq := &metaclient.BeginObjectParams{
		Bucket:               []byte(o.bucket),
		EncryptedObjectKey:   []byte(encPath.Raw()),
		ExpiresAt:            o.expiration,
		EncryptionParameters: o.encryptionParameters,
	}

	var streamID storj.StreamID
	defer func() {
		if err != nil && !streamID.IsZero() {
			s.deleteCancelledObject(context2.WithoutCancellation(ctx), bucket, encPath.Raw(), streamID)
			return
		}
	}()

	var (
		currentSegment    uint32
		contentKey        storj.Key
		streamSize        int64
		lastSegmentSize   int64
		encryptedKey      []byte
		encryptedKeyNonce storj.Nonce
		segmentRS         eestream.RedundancyStrategy

		requestsToBatch = make([]metaclient.BatchItem, 0, 2)
	)

}

type SegmentWriter struct {
	chunkSize       int
	segmentSize     int
	segmentPosition int
}

func NewSegmentWriter() (*SegmentWriter, error) {
	return &SegmentWriter{}, nil
}

type ChunkedWriter struct {
	size   int
	output io.Writer
	buffer []byte
	pos    int
}

func NewChunkedWriter(output io.Writer, size int) (*ChunkedWriter, error) {
	return &ChunkedWriter{
		size:   size,
		output: output,
		buffer: make([]byte, size),
	}, nil
}

func (c ChunkedWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if c.pos > 0 {
		panic("not implemented")
		//remaining := c.size - c.pos
		//if remaining > len(p) {
		//	copy(c.buffer[c.pos:],)
		//}
	}

	if len(p) >= c.size {
		n, err := c.output.Write(p[0:c.size])
		if err != nil {
			return n, err
		}
		return c.Write(p[c.size:])
	}
	panic("not implemented")
}

var _ io.Writer = &ChunkedWriter{}

type Buffer struct {
	Buffer []byte
}

type BufferWriter interface {
	Write(buffer Buffer) error
}

type StreamWriter struct {
}

type EncryptedWriter struct {
	transformer encryption.Transformer
	blockNo     int64
	buffer      []byte
	output      EncryptedOutput
}

type EncryptedOutput func([]byte) error

func NewEncryptedWriter(output EncryptedOutput, cipher storj.CipherSuite, key *storj.Key, startingNonce *storj.Nonce, encryptedBlockSize int) (*EncryptedWriter, error) {
	transformer, err := encryption.NewEncrypter(cipher, key, startingNonce, encryptedBlockSize)
	if err != nil {
		return nil, err
	}
	return &EncryptedWriter{
		transformer: transformer,
		buffer:      make([]byte, encryptedBlockSize),
		output:      output,
	}, nil
}

func (e EncryptedWriter) Write(data []byte) (int, error) {
	if len(data) != e.transformer.InBlockSize() {
		return 0, errs.New("EncryptedOutput requires exactly %d bytes to encrypt (but %d)", e.transformer.InBlockSize(), len(data))
	}
	out, err := e.transformer.Transform(e.buffer[:0], data, e.blockNo)
	if err != nil {
		return 0, err
	}
	return e.transformer.OutBlockSize(), e.output(out)
}

type ECWriter struct {
	fec    *infectious.FEC
	output ECOutput
}

type ECOutput interface {
	WriteShare(number int, data []byte)
	Wait() error
}

func NewECWRiter(output *PieceGroupWriter, scheme *pb.RedundancyScheme) (*ECWriter, error) {
	fec, err := infectious.NewFEC(int(scheme.GetMinReq()), int(scheme.GetTotal()))
	if err != nil {
		return nil, err
	}
	return &ECWriter{
		fec:    fec,
		output: output,
	}, nil
}

func (e *ECWriter) Write(buffer []byte) error {
	err := e.fec.Encode(buffer, func(share infectious.Share) {
		e.output.WriteShare(share.Number, share.Data)
	})
	return errs.Combine(err, e.output.Wait())
}

type PieceGroupWriter struct {
}

func NewPieceGroupWriter() (*PieceGroupWriter, error) {
	return &PieceGroupWriter{}, nil
}

func (p PieceGroupWriter) WriteShare(number int, data []byte) {
	fmt.Printf("write share %d %d\n", number, len(data))
}

func (p PieceGroupWriter) Wait() error {
	return nil
}

var _ ECOutput = &PieceGroupWriter{}

type PieceWriter struct {
	client *piecestore.Client
}

func NewPieceWriter(ctx context.Context, limit *pb.AddressedOrderLimit, privateKey storj.PiecePrivateKey) (*PieceWriter, error) {
	//defer mon.Task()(&ctx, "node: "+nodeName)(&err)
	//defer func() { err = errs.Combine(err, data.Close()) }()
	//
	//storageNodeID := limit.GetLimit().StorageNodeId
	//
	//client, err := piecestore.Dial(ctx, ec.dialer, storj.NodeURL{
	//	ID:      storageNodeID,
	//	Address: limit.GetStorageNodeAddress().Address,
	//}, piecestore.DefaultConfig)

	return &PieceWriter{
		client: nil,
	}, nil

}

func (p *PieceWriter) Write() error {
	return nil
}

type Discard struct {
}

func (d Discard) Write(buffer Buffer) error {
	fmt.Println("write")
	return nil
}

var _ BufferWriter = Discard{}
