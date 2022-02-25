package eestream

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/vivint/infectious"
	"io"
	"io/ioutil"
	"storj.io/common/sync2"
	"storj.io/common/testcontext"
	"testing"
)

func TestContentReader(t *testing.T) {
	c := &ContentReader{MaxSize: 12}
	r, err := c.Read(make([]byte, 10))
	require.Equal(t, 10, r)
	require.NoError(t, err)

	r, err = c.Read(make([]byte, 2))
	require.Equal(t, 2, r)
	require.NoError(t, err)

	r, err = c.Read(make([]byte, 3))
	require.Equal(t, 0, r)
	require.Error(t, err)
}

func TestEncodeReader2(t *testing.T) {
	ctx := testcontext.New(t)

	fc, err := infectious.NewFEC(3, 6)

	require.NoError(t, err)
	es := NewRSScheme(fc, 1024)

	rs, err := NewRedundancyStrategy(es, 0, 0)
	require.NoError(t, err)

	readers, err := EncodeReader2(ctx, &ContentReader{MaxSize: 1024 * 3}, rs)
	require.NoError(t, err)

	wg := sync2.WorkGroup{}

	for _, r := range readers {
		_, err := io.Copy(ioutil.Discard, r)
		require.NoError(t, err)
	}

	wg.Wait()
}

type ContentReader struct {
	MaxSize  int
	position int
}

func (c *ContentReader) Read(p []byte) (n int, err error) {
	fmt.Printf("read %d\n", len(p))
	if c.position >= c.MaxSize {
		return 0, io.EOF
	}
	n = len(p)
	if n > c.MaxSize-c.position {
		n = c.MaxSize - c.position
	}
	c.position += n
	return n, nil
}

var _ io.Reader = &ContentReader{}
