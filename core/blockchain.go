package core

import (
	"blockchain/crypto"
	"blockchain/types"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
)

type Blockchain struct {
	blocksMu        sync.RWMutex
	blocks          []*Block
	blocksMap       map[types.Hash]*Block
	transactionsMap map[types.Hash]*Transaction
	collectionsMap  map[types.Hash]*Collection
	mintsMap        map[types.Hash]*Mint
	validator       Validator
	accountsState   *AccountsState
	contractState   *State
	store           Storage
}

func NewBlockchain(genesisBlock *Block) (*Blockchain, error) {
	accountState := NewAccountsState()
	coinBase := crypto.PublicKey{}
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	accountState.CreateAccount(coinBase.Address(), balance)

	bc := &Blockchain{
		blocksMap:       make(map[types.Hash]*Block),
		transactionsMap: make(map[types.Hash]*Transaction),
		collectionsMap:  make(map[types.Hash]*Collection),
		mintsMap:        make(map[types.Hash]*Mint),
		// read from some DB on startup
		accountsState: accountState,
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
	return bc.saveBlock(b)
}

func (bc *Blockchain) handleTransfer(tx *Transaction) error {
	return bc.accountsState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)
}

func (bc *Blockchain) handleNFT(tx *Transaction) error {
	hash := tx.Hash(TransactionHasher{})

	switch v := tx.Inner.(type) {
	case *Collection:
		bc.collectionsMap[hash] = v
		fmt.Println("created new NFT collection:", hash)
	case *Mint:
		_, ok := bc.collectionsMap[v.Collection]
		if !ok {
			return fmt.Errorf("collection (%s) doesn't exist on the blockchain", v.Collection)
		}
		bc.mintsMap[hash] = v
		fmt.Printf("created new NFT (%s), collection (%s)\n", v.NFT, v.Collection)
	default:
		return fmt.Errorf("unsupported transaction type: (%s)", v)
	}

	return nil
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("height (%d) is too high", height)
	}

	bc.blocksMu.Lock()
	defer bc.blocksMu.Unlock()

	return bc.blocks[height], nil
}

func (bc *Blockchain) GetBlockByHeaderHash(hash types.Hash) (*Block, error) {
	bc.blocksMu.Lock()
	defer bc.blocksMu.Unlock()

	block, ok := bc.blocksMap[hash]
	if !ok {
		return nil, fmt.Errorf("block with header hash (%s) couldn't be found", hash)
	}

	return block, nil
}

func (bc *Blockchain) GetTransaction(hash types.Hash) (*Transaction, error) {
	transaction, ok := bc.transactionsMap[hash]
	if !ok {
		return nil, fmt.Errorf("transaction with hash (%s) couldn't be found", hash)
	}
	return transaction, nil
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

func (bc *Blockchain) handleTransaction(tx *Transaction) error {
	if tx.Data != nil {
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
		slog.Info("Contract state:", "result", vm.contractState)
		//slog.Info("VM:", "result", deserializeInt64(vm.stack.Pop().([]byte)))
	}

	if tx.Inner != nil {
		if err := bc.handleNFT(tx); err != nil {
			return err
		}
	}

	if tx.Value != nil {
		if tx.Value.Cmp(new(big.Int)) == 1 {
			if err := bc.handleTransfer(tx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (bc *Blockchain) saveBlock(b *Block) error {
	for _, tx := range b.Transactions {
		if err := bc.handleTransaction(tx); err != nil {
			fmt.Println(err)
			continue
		}
		bc.transactionsMap[tx.Hash(TransactionHasher{})] = tx
	}

	slog.Info(
		"adding new block",
		"height", b.Height,
		"hash", b.HeaderHash(HeaderHasher{}),
	)

	bc.blocksMu.Lock()
	defer bc.blocksMu.Unlock()

	bc.blocks = append(bc.blocks, b)
	bc.blocksMap[b.HeaderHash(HeaderHasher{})] = b

	return bc.store.Put(b)
}
