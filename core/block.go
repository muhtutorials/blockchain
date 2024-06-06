package core

import (
	"blockchain/crypto"
	"blockchain/types"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
	"time"
)

type Header struct {
	Version          uint32
	TransactionsHash types.Hash
	PrevHeaderHash   types.Hash
	Height           uint32
	Timestamp        int64
}

func (h Header) Bytes() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(h)
	return buf.Bytes()
}

type Block struct {
	*Header
	Transactions []*Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature
}

func NewBlock(h *Header, txs []*Transaction) *Block {
	return &Block{
		Header:       h,
		Transactions: txs,
	}
}

func NewBlockFromPrevHeader(prevHeader *Header, txs []*Transaction) (*Block, error) {
	transactionsHash, err := HashTransactions(txs)
	if err != nil {
		return nil, err
	}

	header := &Header{
		Version:          1,
		TransactionsHash: transactionsHash,
		PrevHeaderHash:   HeaderHasher{}.Hash(prevHeader),
		Height:           prevHeader.Height + 1,
		Timestamp:        time.Now().UnixNano(),
	}

	return NewBlock(header, txs), nil
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}

func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func (b *Block) HeaderHash(hasher Hasher[*Header]) types.Hash {
	return hasher.Hash(b.Header)
}

func (b *Block) Sign(priv *crypto.PrivateKey) error {
	hash := b.HeaderHash(HeaderHasher{})
	sig, err := priv.Sign(hash.Bytes())
	if err != nil {
		return err
	}
	b.Validator = priv.PublicKey()
	b.Signature = sig
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block has no signature")
	}

	hash := b.HeaderHash(HeaderHasher{})
	if !b.Signature.Verify(b.Validator, hash.Bytes()) {
		return fmt.Errorf("invalid block signature")
	}

	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	transactionsHash, err := HashTransactions(b.Transactions)
	if err != nil {
		return err
	}

	if transactionsHash != b.TransactionsHash {
		return fmt.Errorf("block (%s) has invalid transactions hash", b.HeaderHash(HeaderHasher{}))
	}

	return nil
}

func HashTransactions(txs []*Transaction) (types.Hash, error) {
	buf := new(bytes.Buffer)
	for _, tx := range txs {
		if err := tx.Encode(NewGobTransactionEncoder(buf)); err != nil {
			return types.Hash{}, err
		}
	}
	hash := sha256.Sum256(buf.Bytes())
	return hash, nil
}

func CreateGenesisBlock() *Block {
	// transactions hash shouldn't be saved in header because of the random nonce
	// which makes genesis blocks differ on every server
	h := &Header{
		Version: 1,
	}

	coinBase := crypto.PublicKey{}
	tx := NewTransaction(nil)
	tx.From = coinBase
	tx.To = coinBase
	value, _ := new(big.Int).SetString("1000000000000000000", 10)
	tx.Value = value

	b := NewBlock(h, []*Transaction{tx})

	return b
}

//func (h *Header) EncodeBinary(w io.Writer) error {
//	if err := binary.Write(w, binary.LittleEndian, &h.Version); err != nil {
//		return err
//	}
//	if err := binary.Write(w, binary.LittleEndian, &h.PrevBlock); err != nil {
//		return err
//	}
//	if err := binary.Write(w, binary.LittleEndian, &h.Height); err != nil {
//		return err
//	}
//	if err := binary.Write(w, binary.LittleEndian, &h.Nonce); err != nil {
//		return err
//	}
//	if err := binary.Write(w, binary.LittleEndian, &h.Timestamp); err != nil {
//		return err
//	}
//	return nil
//}
//
//func (h *Header) DecodeBinary(r io.Reader) error {
//	if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
//		return err
//	}
//	if err := binary.Read(r, binary.LittleEndian, &h.PrevBlock); err != nil {
//		return err
//	}
//	if err := binary.Read(r, binary.LittleEndian, &h.Height); err != nil {
//		return err
//	}
//	if err := binary.Read(r, binary.LittleEndian, &h.Nonce); err != nil {
//		return err
//	}
//	if err := binary.Read(r, binary.LittleEndian, &h.Timestamp); err != nil {
//		return err
//	}
//	return nil
//}

//func (b *Block) Hash() (types.Hash, error) {
//	buf := new(bytes.Buffer)
//	if err := b.Header.EncodeBinary(buf); err != nil {
//		return [32]uint8{}, err
//	}
//
//	if b.hash.IsZero() {
//		b.hash = sha256.Sum256(buf.Bytes())
//	}
//
//	return b.hash, nil
//}
//
//func (b *Block) EncodeBinary(w io.Writer) error {
//	if err := b.Header.EncodeBinary(w); err != nil {
//		return err
//	}
//
//	for _, tx := range b.Transactions {
//		if err := tx.EncodeBinary(w); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (b *Block) DecodeBinary(r io.Reader) error {
//	if err := b.Header.DecodeBinary(r); err != nil {
//		return err
//	}
//
//	for _, tx := range b.Transactions {
//		if err := tx.DecodeBinary(r); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
