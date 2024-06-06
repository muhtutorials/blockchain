package core

import (
	"blockchain/types"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type HeaderHasher struct{}

func (HeaderHasher) Hash(h *Header) types.Hash {
	return sha256.Sum256(h.Bytes())
}

type TransactionHasher struct{}

func (TransactionHasher) Hash(tx *Transaction) types.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tx.Data)
	binary.Write(buf, binary.LittleEndian, tx.From)
	binary.Write(buf, binary.LittleEndian, tx.To)
	binary.Write(buf, binary.LittleEndian, tx.Value)
	binary.Write(buf, binary.LittleEndian, tx.Nonce)
	return sha256.Sum256(buf.Bytes())
}
