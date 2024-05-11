package core

import (
	"fmt"
	"log/slog"
	"sync"
)

type Blockchain struct {
	mu        sync.RWMutex
	headers   []*Header
	validator Validator
	store     Storage
}

func NewBlockchain(genesisBlock *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers: []*Header{},
		store:   NewMemoryStore(),
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.saveBlock(genesisBlock)
	if err != nil {
		return nil, err
	}
	return bc, nil
}

func (bc *Blockchain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *Blockchain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}
	slog.Info(
		"adding new block",
		"height", b.Height,
		"hash", b.Hash(BlockHasher{}),
	)
	return bc.saveBlock(b)
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) is too high", height)
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	return bc.headers[height], nil
}

// Height returns number of blocks in the blockchain.
// First block is the genesis block which is not included in height
func (bc *Blockchain) Height() uint32 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) saveBlock(b *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.headers = append(bc.headers, b.Header)
	return bc.store.Put(b)
}
