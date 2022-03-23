package upload

import (
	"context"
	"storj.io/common/encryption"
	"storj.io/common/storj"
)

type Encrypter struct {
	transformer        encryption.Transformer
	blockNo            int64
	buffer             []byte
	output             EncryptedWriterLayer
	encryptedBlockSize int
}

type EncryptedWriterLayer interface {
	StartObject(ctx context.Context, info *StartObject) error
	StartSegment(ctx context.Context, index int) error
	EndSegment(ctx context.Context) error
	EndObject(ctx context.Context) error
	EncryptedWrite(context.Context, int, []byte) error
	EncryptedLastWrite(context.Context, int, []byte) error
	EncryptionParams(ctx context.Context, v EncryptionParams) error
}

var _ SegmentedWithKeyLayer = &Encrypter{}

func (e *Encrypter) Write(ctx context.Context, bytes []byte) error {
	out, err := e.transformer.Transform(e.buffer, bytes, 0)
	if err != nil {
		return err
	}
	return e.output.EncryptedWrite(ctx, len(bytes), out)
}

func (e *Encrypter) LastWrite(ctx context.Context, bytes []byte) error {
	out, err := e.transformer.Transform(e.buffer, bytes, 0)
	if err != nil {
		return err
	}
	return e.output.EncryptedLastWrite(ctx, len(bytes), out)
}

func (e *Encrypter) StartObject(ctx context.Context, info *StartObject) error {
	return e.output.StartObject(ctx, info)
}

func (e *Encrypter) StartSegment(ctx context.Context, index int) error {
	return e.output.StartSegment(ctx, index)
}

func (e *Encrypter) EndSegment(ctx context.Context) error {
	return e.output.EndSegment(ctx)
}

func (e *Encrypter) EndObject(ctx context.Context) error {
	return e.output.EndObject(ctx)
}

func (e *Encrypter) EncryptionParams(ctx context.Context, v EncryptionParams) (err error) {
	//TODO: what is the nonce here
	e.transformer, err = encryption.NewEncrypter(v.cipher, v.contentKey, &storj.Nonce{}, e.encryptedBlockSize)
	if err != nil {
		return err
	}
	return e.output.EncryptionParams(ctx, v)
}
