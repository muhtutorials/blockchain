package network

import (
	"blockchain/core"
	"blockchain/crypto"
	"encoding/hex"
	"slices"
)

type ErrorRes struct {
	Error string `json:"error"`
}

type BlockRes struct {
	Version          uint32       `json:"version"`
	TransactionsHash string       `json:"transactions_hash"`
	PrevHeaderHash   string       `json:"prev_header_hash"`
	Height           uint32       `json:"height"`
	Timestamp        int64        `json:"timestamp"`
	Transactions     []string     `json:"transactions"`
	Validator        string       `json:"validator"`
	Signature        SignatureRes `json:"signature"`
	HeaderHash       string       `json:"header_hash"`
}

func ToBlockRes(b *core.Block) *BlockRes {
	blockRes := &BlockRes{
		Version:          b.Version,
		TransactionsHash: b.TransactionsHash.String(),
		PrevHeaderHash:   b.PrevHeaderHash.String(),
		Height:           b.Height,
		Timestamp:        b.Timestamp,
		Validator:        hex.EncodeToString(b.Validator),
		Signature:        ToSignatureRes(b.Signature),
		HeaderHash:       b.HeaderHash(core.HeaderHasher{}).String(),
	}

	var transactions []string
	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Hash(core.TransactionHasher{}).String())
	}
	blockRes.Transactions = transactions

	return blockRes
}

type TransactionRes struct {
	Data      []byte       `json:"data"`
	From      string       `json:"from"`
	Signature SignatureRes `json:"signature"`
	Hash      string       `json:"hash"`
}

func ToTransactionRes(tx *core.Transaction) *TransactionRes {
	return &TransactionRes{
		Data:      tx.Data,
		From:      hex.EncodeToString(tx.From),
		Signature: ToSignatureRes(tx.Signature),
		Hash:      tx.Hash(core.TransactionHasher{}).String(),
	}
}

type SignatureRes string

func ToSignatureRes(s *crypto.Signature) SignatureRes {
	signatureBytes := slices.Concat(s.R.Bytes(), s.S.Bytes())
	return SignatureRes(hex.EncodeToString(signatureBytes))
}
