package core

import (
	"blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestAccountsState(t *testing.T) {
	accountState := NewAccountsState()

	coinBase := crypto.PublicKey{}.Address()
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	accountState.CreateAccount(coinBase, balance)
	account, err := accountState.getAccount(coinBase)
	assert.Nil(t, err)
	assert.Equal(t, account.Address, coinBase)
}

func TestAccountsStateTransferSuccess(t *testing.T) {
	accountState := NewAccountsState()

	coinBase := crypto.PublicKey{}.Address()
	balance := new(big.Int).SetUint64(10_000_000_000)
	accountState.CreateAccount(coinBase, balance)
	account, err := accountState.getAccount(coinBase)
	assert.Nil(t, err)
	assert.Equal(t, account.Address, coinBase)

	to := crypto.GeneratePrivateKey().PublicKey().Address()
	amount := new(big.Int).SetUint64(5_000_000_000)
	assert.Nil(t, accountState.Transfer(coinBase, to, amount))

	balance, err = accountState.getBalance(to)
	assert.Nil(t, err)
	assert.Equal(t, amount, balance)
}

func TestAccountsStateTransferFail(t *testing.T) {
	accountState := NewAccountsState()

	coinBase := crypto.PublicKey{}.Address()
	balance, _ := new(big.Int).SetString("1000000000", 10)
	accountState.CreateAccount(coinBase, balance)
	account, err := accountState.getAccount(coinBase)
	assert.Nil(t, err)
	assert.Equal(t, account.Address, coinBase)

	to := crypto.GeneratePrivateKey().PublicKey().Address()
	amount, _ := new(big.Int).SetString("5000000000", 10)

	err = accountState.Transfer(coinBase, to, amount)
	assert.NotNil(t, err)
}
