package core

import (
	"blockchain/crypto"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

	//txDecoded := new(Transaction)
	//assert.Nil(t, txDecoded.Decode(NewGobTransactionDecoder(buf)))
	//assert.Equal(t, tx, txDecoded)
}

func randomTxWithSignature() *Transaction {
	privateKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("hey"),
	}
	tx.Sign(privateKey)
	return tx
}
