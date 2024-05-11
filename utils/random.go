package utils

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/types"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func RandomBytes(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func RandomHash() types.Hash {
	return types.HashFromBytes(RandomBytes(32))
}

func NewRandomTransaction(size int) *core.Transaction {
	return core.NewTransaction(RandomBytes(size))
}

func NewRandomTransactionWithSignature(t *testing.T, size int, privKey *crypto.PrivateKey) *core.Transaction {
	tx := NewRandomTransaction(size)
	assert.Nil(t, tx.Sign(privKey))
	return tx
}

func NewRandomBlock(t *testing.T, prevBlockHash types.Hash, height uint32) *core.Block {
	privateKey := crypto.GeneratePrivateKey()
	tx := NewRandomTransactionWithSignature(t, 100, privateKey)
	h := &core.Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Timestamp:     time.Now().UnixNano(),
	}

	b := core.NewBlock(h, []*core.Transaction{tx})
	err := b.Sign(privateKey)
	assert.Nil(t, err)

	blockHash, err := core.HashBlock(b.Transactions)
	b.Header.BlockHash = blockHash
	assert.Nil(t, err)

	return b
}

func NewRandomBlockWithSignature(t *testing.T, pk *crypto.PrivateKey, prevBlockHash types.Hash, height uint32) *core.Block {
	b := NewRandomBlock(t, prevBlockHash, height)
	assert.Nil(t, b.Sign(pk))
	return b
}
