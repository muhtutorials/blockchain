package core

import (
	"blockchain/types"
	"crypto/sha256"
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
	return sha256.Sum256(tx.Data)
}
