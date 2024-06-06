package core

import (
	"blockchain/crypto"
	"blockchain/types"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestBlockchain(t *testing.T) {
	bc, err := NewBlockchain(randomBlock(t, types.Hash{}, 0, nil))
	assert.Nil(t, err)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc, _ := NewBlockchain(randomBlock(t, types.Hash{}, 0, nil))
	assert.True(t, bc.HasBlock(uint32(0)))
	assert.False(t, bc.HasBlock(uint32(120)))
}

func TestAddBlock(t *testing.T) {
	bc, _ := NewBlockchain(randomBlock(t, types.Hash{}, 0, nil))
	lenBlocks := 10
	for i := 0; i < lenBlocks; i++ {
		b := randomBlock(t, getPrevBlockHash(t, bc, uint32(i+1)), uint32(i+1), nil)
		assert.Nil(t, bc.AddBlock(b))
	}
	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.blocks), lenBlocks+1)
	assert.NotNil(t, bc.AddBlock(randomBlock(t, types.Hash{}, uint32(223), nil)))
	assert.NotNil(t, bc.AddBlock(randomBlock(t, types.Hash{}, uint32(5223), nil)))
}

func TestGetBlock(t *testing.T) {
	bc, _ := NewBlockchain(randomBlock(t, types.Hash{}, 0, nil))
	lenBlocks := 10
	for i := 0; i < lenBlocks; i++ {
		b := randomBlock(t, getPrevBlockHash(t, bc, uint32(i+1)), uint32(i+1), nil)
		assert.Nil(t, bc.AddBlock(b))
		block, err := bc.GetBlock(uint32(i + 1))
		assert.Nil(t, err)
		assert.Equal(t, b, block)
	}
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) types.Hash {
	prevBlock, err := bc.GetBlock(height - 1)
	assert.Nil(t, err)
	return HeaderHasher{}.Hash(prevBlock.Header)
}

func TestTransferSuccess(t *testing.T) {
	bc, _ := NewBlockchain(CreateGenesisBlock())

	bob := crypto.GeneratePrivateKey()
	alice := crypto.GeneratePrivateKey()

	bc.accountsState.CreateAccount(bob.PublicKey().Address(), new(big.Int).SetUint64(100_000_000_000))

	tx := NewTransaction(nil)
	tx.To = alice.PublicKey()
	tx.Value = new(big.Int).SetUint64(3_000_000_000)
	tx.Sign(bob)

	block := randomBlock(t, getPrevBlockHash(t, bc, uint32(1)), uint32(1), []*Transaction{tx})
	assert.Nil(t, bc.AddBlock(block))

	bobBalance, _ := bc.accountsState.getBalance(bob.PublicKey().Address())
	assert.Equal(t, new(big.Int).SetUint64(97_000_000_000), bobBalance)
	aliceBalance, _ := bc.accountsState.getBalance(alice.PublicKey().Address())
	assert.Equal(t, new(big.Int).SetUint64(3_000_000_000), aliceBalance)
}

func TestTransferHacked(t *testing.T) {
	bc, _ := NewBlockchain(CreateGenesisBlock())

	bob := crypto.GeneratePrivateKey()
	alice := crypto.GeneratePrivateKey()

	bc.accountsState.CreateAccount(bob.PublicKey().Address(), new(big.Int).SetUint64(100_000_000_000))

	tx := NewTransaction(nil)
	tx.To = alice.PublicKey()
	tx.Value = new(big.Int).SetUint64(3_000_000_000)
	assert.Nil(t, tx.Sign(bob))

	hacker := crypto.GeneratePrivateKey()
	tx.To = hacker.PublicKey()

	block := randomBlock(t, getPrevBlockHash(t, bc, uint32(1)), uint32(1), []*Transaction{tx})
	assert.NotNil(t, bc.AddBlock(block))

	hackerBalance, err := bc.accountsState.getBalance(hacker.PublicKey().Address())
	assert.NotNil(t, err)
	fmt.Println("hacker:", hackerBalance)

	bobBalance, _ := bc.accountsState.getBalance(bob.PublicKey().Address())
	assert.Equal(t, new(big.Int).SetUint64(100_000_000_000), bobBalance)
	aliceBalance, err := bc.accountsState.getBalance(alice.PublicKey().Address())
	assert.NotNil(t, err)
	assert.Equal(t, new(big.Int), aliceBalance)
}
