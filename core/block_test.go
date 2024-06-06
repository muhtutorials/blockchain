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
	block := randomBlock(t, types.Hash{}, 0, nil)
	assert.Nil(t, block.Sign(privateKey))
	assert.Nil(t, block.Verify())

	otherPrivateKey := crypto.GeneratePrivateKey()
	block.Validator = otherPrivateKey.PublicKey()
	assert.NotNil(t, block.Verify())
}

func TestBlock_EncodeAndDecode(t *testing.T) {
	block := randomBlock(t, types.Hash{}, 0, nil)
	buf := new(bytes.Buffer)
	assert.Nil(t, block.Encode(NewGobBlockEncoder(buf)))

	decodedBlock := new(Block)
	err := decodedBlock.Decode(NewGobBlockDecoder(buf))
	assert.Nil(t, err)
	assert.Equal(t, block, decodedBlock)
}

func TestBlock_Hacked(t *testing.T) {
	block := randomBlock(t, types.Hash{}, 0, nil)
	assert.Nil(t, block.Verify())

	block.Height = 1
	assert.NotNil(t, block.Verify())
}

func randomBlock(t *testing.T, prevHeaderHash types.Hash, height uint32, txs []*Transaction) *Block {
	privateKey := crypto.GeneratePrivateKey()
	h := &Header{
		Version:        1,
		PrevHeaderHash: prevHeaderHash,
		Height:         height,
		Timestamp:      time.Now().UnixNano(),
	}

	if txs == nil {
		tx := randomTxWithSignature()
		txs = []*Transaction{tx}
	}

	block := NewBlock(h, txs)
	transactionsHash, err := HashTransactions(block.Transactions)
	block.Header.TransactionsHash = transactionsHash
	assert.Nil(t, err)

	assert.Nil(t, block.Sign(privateKey))

	return block
}
