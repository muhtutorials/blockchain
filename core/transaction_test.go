package core

import (
	"blockchain/crypto"
	"bytes"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func randomTxWithSignature() *Transaction {
	privateKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("hey"),
	}
	tx.Sign(privateKey)
	return tx
}

func TestTransaction_SignAndVerify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("hey"),
	}
	assert.Nil(t, tx.Sign(privateKey))
	assert.Nil(t, tx.Verify())

	otherPrivateKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivateKey.PublicKey()
	assert.NotNil(t, tx.Verify())
}

func TestTransaction_EncodeDecode(t *testing.T) {
	tx := randomTxWithSignature()
	buf := new(bytes.Buffer)
	assert.Nil(t, tx.Encode(NewGobTransactionEncoder(buf)))

	txDecoded := new(Transaction)
	assert.Nil(t, txDecoded.Decode(NewGobTransactionDecoder(buf)))
	assert.Equal(t, tx, txDecoded)
}

func TestNFTTransaction(t *testing.T) {
	coll := &Collection{
		MetaData: []byte("ok"),
		Fee:      200,
	}
	privateKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Inner: coll,
	}
	assert.Nil(t, tx.Sign(privateKey))

	buf := new(bytes.Buffer)
	assert.Nil(t, tx.Encode(NewGobTransactionEncoder(buf)))

	txDecoded := new(Transaction)
	assert.Nil(t, txDecoded.Decode(NewGobTransactionDecoder(buf)))
	assert.Equal(t, tx, txDecoded)
}

func TestTransferTransaction(t *testing.T) {
	fromPrivateKey := crypto.GeneratePrivateKey()
	toPrivateKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		To:    toPrivateKey.PublicKey(),
		Value: new(big.Int).SetUint64(1_000_000_000_000),
	}
	assert.Nil(t, tx.Sign(fromPrivateKey))
}

func TestTransactionHacked(t *testing.T) {
	bob := crypto.GeneratePrivateKey()
	alice := crypto.GeneratePrivateKey()
	hacker := crypto.GeneratePrivateKey()

	tx := NewTransaction(nil)
	tx.From = bob.PublicKey()
	tx.To = alice.PublicKey()
	tx.Value = new(big.Int).SetInt64(1_000_000_000)

	assert.Nil(t, tx.Sign(bob))

	tx.To = hacker.PublicKey()

	assert.NotNil(t, tx.Verify())
}
