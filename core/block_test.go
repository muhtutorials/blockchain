package core

import (
	"blockchain/crypto"
	"blockchain/types"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBlock_SignAndVerify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(t, types.Hash{}, 0)
	assert.Nil(t, block.Sign(privateKey))
	assert.Nil(t, block.Verify())

	otherPrivateKey := crypto.GeneratePrivateKey()
	block.Validator = otherPrivateKey.PublicKey()
	assert.NotNil(t, block.Verify())

	block.Height = 1
	assert.NotNil(t, block.Verify())
}

func TestBlock_EncodeAndDecode(t *testing.T) {
	block := randomBlock(t, types.Hash{}, 0)
	buf := new(bytes.Buffer)
	assert.Nil(t, block.Encode(NewGobBlockEncoder(buf)))

	decodedBlock := new(Block)
	err := decodedBlock.Decode(NewGobBlockDecoder(buf))
	assert.Nil(t, err)
	assert.Equal(t, block, decodedBlock)
}

func randomBlock(t *testing.T, prevHeaderHash types.Hash, height uint32) *Block {
	privateKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature()
	h := &Header{
		Version:        1,
		PrevHeaderHash: prevHeaderHash,
		Height:         height,
		Timestamp:      time.Now().UnixNano(),
	}

	block := NewBlock(h, []*Transaction{tx})
	err := block.Sign(privateKey)
	assert.Nil(t, err)

	transactionsHash, err := HashTransactions(block.Transactions)
	block.Header.TransactionsHash = transactionsHash
	assert.Nil(t, err)

	return block
}
