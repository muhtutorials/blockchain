package network

import (
	"blockchain/core"
	"blockchain/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransactionPool_MaxLength(t *testing.T) {
	pool := NewTransactionPool(1, core.TransactionHasher{})
	err := pool.Add(utils.NewRandomTransaction(10))
	assert.Nil(t, err)
	assert.Equal(t, 1, pool.all.Count())

	err = pool.Add(utils.NewRandomTransaction(10))
	assert.Nil(t, err)
	err = pool.Add(utils.NewRandomTransaction(10))
	assert.Nil(t, err)
	err = pool.Add(utils.NewRandomTransaction(10))
	assert.Nil(t, err)
	tx := utils.NewRandomTransaction(100)
	err = pool.Add(tx)
	assert.Nil(t, err)
	assert.Equal(t, 1, pool.all.Count())
	assert.True(t, pool.Contains(tx.Hash(core.TransactionHasher{})))
}

func TestTransactionPool_MaxLength_2(t *testing.T) {
	var txs []*core.Transaction
	maxLength := 10
	n := 100

	pool := NewTransactionPool(maxLength, core.TransactionHasher{})

	for i := 0; i < n; i++ {
		tx := utils.NewRandomTransaction(100)
		err := pool.Add(tx)
		assert.Nil(t, err)

		if i > n-(maxLength+1) { // i > 89
			txs = append(txs, tx)
		}
	}

	assert.Equal(t, pool.all.Count(), maxLength)
	assert.Equal(t, len(txs), maxLength)

	for _, tx := range txs {
		assert.True(t, pool.Contains(tx.Hash(core.TransactionHasher{})))
	}
}

func TestTransactionPool_Add(t *testing.T) {
	pool := NewTransactionPool(11, core.TransactionHasher{})
	n := 10

	for i := 1; i <= n; i++ {
		tx := utils.NewRandomTransaction(100)
		err := pool.Add(tx)
		assert.Nil(t, err)
		err = pool.Add(tx)
		assert.Nil(t, err)

		assert.Equal(t, i, pool.PendingCount())
		assert.Equal(t, i, pool.all.Count())
		assert.Equal(t, i, pool.pending.Count())
	}
}

func TestTransactionList_Add_2(t *testing.T) {
	list := NewTransactionList()
	n := 100

	for i := 0; i < n; i++ {
		tx := utils.NewRandomTransaction(100)
		err := list.Add(tx, core.TransactionHasher{})
		assert.Nil(t, err)
		err = list.Add(tx, core.TransactionHasher{})
		assert.NotNil(t, err)

		assert.Equal(t, list.Count(), i+1)
		assert.True(t, list.Contains(tx.Hash(core.TransactionHasher{})))
		assert.Equal(t, len(list.lookup), len(list.transactions))
		txGet, _ := list.Get(tx.Hash(core.TransactionHasher{}))
		assert.Equal(t, txGet, tx)
	}

	list.Clear()
	assert.Equal(t, list.Count(), 0)
	assert.Equal(t, len(list.transactions), 0)
}

func TestTransactionList_First(t *testing.T) {
	list := NewTransactionList()
	first := utils.NewRandomTransaction(100)
	err := list.Add(first, core.TransactionHasher{})
	assert.Nil(t, err)
	err = list.Add(utils.NewRandomTransaction(10), core.TransactionHasher{})
	assert.Nil(t, err)
	err = list.Add(utils.NewRandomTransaction(10), core.TransactionHasher{})
	assert.Nil(t, err)
	firstGet, err := list.First()
	assert.Equal(t, firstGet, first)
}

func TestTransactionList_Delete(t *testing.T) {
	list := NewTransactionList()

	tx := utils.NewRandomTransaction(100)
	err := list.Add(tx, core.TransactionHasher{})
	assert.Nil(t, err)
	assert.Equal(t, list.Count(), 1)

	err = list.Delete(tx.Hash(core.TransactionHasher{}))
	assert.Nil(t, err)
	assert.Equal(t, list.Count(), 0)
	assert.False(t, list.Contains(tx.Hash(core.TransactionHasher{})))
}
