package core

import (
	"blockchain/crypto"
	"blockchain/types"
	"encoding/gob"
	"fmt"
	"math/big"
	"math/rand"
)

type Collection struct {
	MetaData []byte
	Fee      int64
}

type Mint struct {
	MetaData        []byte
	Fee             int64
	NFT             types.Hash
	Collection      types.Hash
	CollectionOwner crypto.PublicKey
	Signature       crypto.Signature
}

type Transaction struct {
	// for NFT
	Inner any
	// for VM
	Data      []byte
	From      crypto.PublicKey
	To        crypto.PublicKey
	Value     *big.Int
	Signature *crypto.Signature
	Nonce     uint64
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: rand.Uint64(),
	}
}

func (tx *Transaction) Sign(priv *crypto.PrivateKey) error {
	tx.From = priv.PublicKey()
	hash := tx.Hash(TransactionHasher{})
	sig, err := priv.Sign(hash[:])
	if err != nil {
		return err
	}
	tx.Signature = sig
	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction has no signature")
	}

	hash := tx.Hash(TransactionHasher{})
	if !tx.Signature.Verify(tx.From, hash[:]) {
		return fmt.Errorf("invalid transaction signature")
	}
	return nil
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	return hasher.Hash(tx)
}

func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}

func (tx *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(tx)
}

func init() {
	gob.Register(&Collection{})
	gob.Register(&Mint{})
}
