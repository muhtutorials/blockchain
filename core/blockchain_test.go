package core

import (
	"blockchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockchain(t *testing.T) {
	bc, err := NewBlockchain(randomBlock(t, types.Hash{}, 0))
	assert.Nil(t, err)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc, _ := NewBlockchain(randomBlock(t, types.Hash{}, 0))
	assert.True(t, bc.HasBlock(uint32(0)))
	assert.False(t, bc.HasBlock(uint32(120)))
}

func TestAddBlock(t *testing.T) {
	bc, _ := NewBlockchain(randomBlock(t, types.Hash{}, 0))
	lenBlocks := 10
	for i := 0; i < lenBlocks; i++ {
		b := randomBlock(t, getPrevBlockHash(t, bc, uint32(i+1)), uint32(i+1))
		assert.Nil(t, bc.AddBlock(b))
	}
	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)
	assert.NotNil(t, bc.AddBlock(randomBlock(t, types.Hash{}, uint32(223))))
	assert.NotNil(t, bc.AddBlock(randomBlock(t, types.Hash{}, uint32(5223))))
}

func TestGetHeader(t *testing.T) {
	bc, _ := NewBlockchain(randomBlock(t, types.Hash{}, 0))
	lenBlocks := 10
	for i := 0; i < lenBlocks; i++ {
		b := randomBlock(t, getPrevBlockHash(t, bc, uint32(i+1)), uint32(i+1))
		assert.Nil(t, bc.AddBlock(b))
		header, err := bc.GetHeader(uint32(i + 1))
		assert.Nil(t, err)
		assert.Equal(t, b.Header, header)
	}
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}
