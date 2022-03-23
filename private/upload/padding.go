package upload

import (
	"context"
)

type Padding struct {
	chunkSize int
	output    WritingLayer
}

func (p Padding) Write(ctx context.Context, bytes []byte) error {
	return p.output.Write(ctx, bytes)
}

func (p Padding) LastWrite(ctx context.Context, bytes []byte) error {
	//TODO: do the padding here
	return p.output.LastWrite(ctx, make([]byte, p.chunkSize))
}

var _ WritingLayer = &Padding{}
