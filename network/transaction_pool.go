package network

import (
	"blockchain/core"
	"blockchain/types"
	"fmt"
	"slices"
	"sync"
)

type TransactionPool struct {
	all     *TransactionList
	pending *TransactionList
	// maxLength of the total pool
	// when the pool is full oldest transactions are pruned
	maxLength int
	hasher    core.Hasher[*core.Transaction]
}

func NewTransactionPool(maxLength int, hasher core.Hasher[*core.Transaction]) *TransactionPool {
	return &TransactionPool{
		all:       NewTransactionList(),
		pending:   NewTransactionList(),
		maxLength: maxLength,
		hasher:    hasher,
	}
}

func (p *TransactionPool) Add(tx *core.Transaction) error {
	if p.all.Count() == p.maxLength {
		oldest, err := p.all.First()
		if err != nil {
			return err
		}
		err = p.all.Delete(oldest.Hash(p.hasher))
		if err != nil {
			return err
		}
	}

	if !p.all.Contains(tx.Hash(p.hasher)) {
		err := p.all.Add(tx, p.hasher)
		if err != nil {
			return err
		}
		err = p.pending.Add(tx, p.hasher)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *TransactionPool) Contains(hash types.Hash) bool {
	return p.all.Contains(hash)
}

func (p *TransactionPool) Pending() []*core.Transaction {
	return p.pending.transactions
}

func (p *TransactionPool) PendingCount() int {
	return p.pending.Count()
}

func (p *TransactionPool) ClearPending() {
	p.pending.Clear()
}

type TransactionList struct {
	mu           sync.RWMutex
	lookup       map[types.Hash]*core.Transaction
	transactions []*core.Transaction
}

func NewTransactionList() *TransactionList {
	return &TransactionList{lookup: make(map[types.Hash]*core.Transaction)}
}

func (l *TransactionList) Add(tx *core.Transaction, hasher core.Hasher[*core.Transaction]) error {
	hash := tx.Hash(hasher)

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.lookup[hash]; ok {
		return fmt.Errorf("TransactionList.Add: transaction with hash (%s) already exists", hash)
	}

	l.lookup[hash] = tx
	l.transactions = append(l.transactions, tx)
	return nil
}

func (l *TransactionList) Get(h types.Hash) (*core.Transaction, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	tx, ok := l.lookup[h]
	if !ok {
		return nil, fmt.Errorf("TransactionList.Get: transaction with hash (%s) wasn't found", h)
	}

	return tx, nil
}

func (l *TransactionList) First() (*core.Transaction, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.transactions) == 0 {
		return nil, fmt.Errorf("TransactionList.First: transaction list is empty")
	}

	return l.transactions[0], nil
}

func (l *TransactionList) Contains(h types.Hash) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	_, ok := l.lookup[h]

	return ok
}

func (l *TransactionList) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return len(l.lookup)
}

func (l *TransactionList) Delete(h types.Hash) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.lookup[h]; !ok {
		return fmt.Errorf("TransactionList.Delete: transaction with hash (%s) wasn't found", h)
	}

	index := slices.Index(l.transactions, l.lookup[h])
	if index == -1 {
		return fmt.Errorf("TransactionList.Delete: transaction with hash (%s) wasn't found", h)
	}
	l.transactions = slices.Delete(l.transactions, index, index+1)

	delete(l.lookup, h)

	return nil
}

func (l *TransactionList) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lookup = make(map[types.Hash]*core.Transaction)
	l.transactions = []*core.Transaction{}
}
