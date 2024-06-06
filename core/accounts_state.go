package core

import (
	"blockchain/types"
	"fmt"
	"math/big"
	"sync"
)

type Account struct {
	Address types.Address
	Balance *big.Int
}

type AccountsState struct {
	mu       sync.RWMutex
	accounts map[types.Address]*Account
}

func NewAccountsState() *AccountsState {
	return &AccountsState{
		accounts: make(map[types.Address]*Account),
	}
}

func (s *AccountsState) CreateAccount(addr types.Address, balance *big.Int) *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account := &Account{
		Address: addr,
		Balance: balance,
	}
	s.accounts[addr] = account

	return account
}

func (s *AccountsState) GetAccount(addr types.Address) (*Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getAccount(addr)
}

func (s *AccountsState) getAccount(addr types.Address) (*Account, error) {
	account, ok := s.accounts[addr]
	if !ok {
		return nil, fmt.Errorf("account (%s) not found", addr)
	}
	return account, nil
}

func (s *AccountsState) GetBalance(addr types.Address) (*big.Int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getBalance(addr)
}

func (s *AccountsState) getBalance(addr types.Address) (*big.Int, error) {
	account, err := s.getAccount(addr)
	if err != nil {
		return &big.Int{}, err
	}
	return account.Balance, nil
}

func (s *AccountsState) Transfer(from, to types.Address, amount *big.Int) error {
	//s.mu.Lock()
	//defer s.mu.Unlock()

	if err := s.SubBalance(from, amount); err != nil {
		return err
	}

	s.AddBalance(to, amount)

	fmt.Printf("transferred (%s) from (%s) to (%s)", amount, from, to)

	return nil
}

func (s *AccountsState) AddBalance(addr types.Address, amount *big.Int) {
	if s.accounts[addr] == nil {
		s.accounts[addr] = &Account{
			Address: addr,
			Balance: new(big.Int),
		}
	}
	s.accounts[addr].Balance.Add(s.accounts[addr].Balance, amount)
}

func (s *AccountsState) SubBalance(addr types.Address, amount *big.Int) error {
	balance, err := s.getBalance(addr)
	if err != nil {
		return err
	}

	if balance.Cmp(amount) == -1 {
		return fmt.Errorf(
			"account (%s) doesn't have enough balance (balance = %d, required amount = %d)",
			addr, balance, amount)
	}

	s.accounts[addr].Balance.Sub(s.accounts[addr].Balance, amount)

	return nil
}
