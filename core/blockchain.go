package core

import (
	"fmt"
	"log/slog"
	"sync"
)

type Blockchain struct {
	blocksMu      sync.RWMutex
	blocks        []*Block
	validator     Validator
	contractState *State
	store         Storage
}

func NewBlockchain(genesisBlock *Block) (*Blockchain, error) {
	bc := &Blockchain{
		contractState: NewState(),
		store:         NewMemoryStore(),
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

	for _, tx := range b.Transactions {
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
		slog.Info("Contract state:", "result", vm.contractState)
		//slog.Info("VM:", "result", deserializeInt64(vm.stack.Pop().([]byte)))
	}

	slog.Info(
		"adding new block",
		"height", b.Height,
		"hash", b.HeaderHash(HeaderHasher{}),
	)
	return bc.saveBlock(b)
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) is too high", height)
	}

	bc.blocksMu.Lock()
	defer bc.blocksMu.Unlock()

	return bc.blocks[height], nil
}

// Height returns number of blocks in the blockchain.
// First block is the genesis block which is not included
func (bc *Blockchain) Height() uint32 {
	bc.blocksMu.RLock()
	defer bc.blocksMu.RUnlock()

	return uint32(len(bc.blocks) - 1)
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) saveBlock(b *Block) error {
	bc.blocksMu.Lock()
	defer bc.blocksMu.Unlock()

	bc.blocks = append(bc.blocks, b)
	return bc.store.Put(b)
}
