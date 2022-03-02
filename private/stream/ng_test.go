package stream

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"storj.io/common/encryption"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"testing"
)

func TestNg(t *testing.T) {
	rsScheme := &pb.RedundancyScheme{
		Type:             pb.RedundancyScheme_RS,
		MinReq:           10,
		Total:            20,
		RepairThreshold:  11,
		SuccessThreshold: 20,
		ErasureShareSize: 1024,
	}

	encParam := storj.EncryptionParameters{
		CipherSuite: storj.EncAESGCM,
		BlockSize:   rsScheme.MinReq * rsScheme.ErasureShareSize,
	}

	contentKey := storj.Key{}
	_, err := rand.Read(contentKey[:])
	require.NoError(t, err)

	contentNonce := storj.Nonce{}
	_, err = encryption.Increment(&contentNonce, int64(0)+1)
	require.NoError(t, err)

	pieceGroupWriter, err := NewPieceGroupWriter()
	require.NoError(t, err)

	ecWriter, err := NewECWRiter(pieceGroupWriter, rsScheme)
	require.NoError(t, err)

	s, err := NewEncryptedWriter(ecWriter.Write, encParam.CipherSuite, &contentKey, &contentNonce, int(encParam.BlockSize))
	require.NoError(t, err)

	c, err := NewChunkedWriter(s, int(rsScheme.MinReq*rsScheme.ErasureShareSize)-s.transformer.OutBlockSize()+s.transformer.InBlockSize())
	bytes := make([]byte, c.size)
	for i := 0; i < 10; i++ {
		_, err = c.Write(bytes)
		require.NoError(t, err)
	}

}
