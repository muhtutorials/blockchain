package core

import (
	"errors"
	"fmt"
)

var ErrBlockAlreadyExists = errors.New("block already exists")

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (v *BlockValidator) ValidateBlock(block *Block) error {
	if v.bc.HasBlock(block.Height) {
		return ErrBlockAlreadyExists
	}
	if block.Height != v.bc.Height()+1 {
		return fmt.Errorf("block (%s) with height (%d) is too high => current height (%d)",
			block.HeaderHash(HeaderHasher{}), block.Height, v.bc.Height())
	}

	prevBlock, err := v.bc.GetBlock(block.Height - 1)
	if err != nil {
		return err
	}
	prevHeaderHash := HeaderHasher{}.Hash(prevBlock.Header)
	if prevHeaderHash != block.PrevHeaderHash {
		return fmt.Errorf("hash of the previous block header is invalid")
	}

	if err = block.Verify(); err != nil {
		return err
	}

	return nil
}
