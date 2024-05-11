package network

import (
	"blockchain/core"
	"blockchain/crypto"
	"bytes"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"
)

var (
	defaultBlockTime = 5 * time.Second
)

type ServerOpts struct {
	ID                string
	InfoLog           *log.Logger
	ErrorLog          *log.Logger
	PrivateKey        *crypto.PrivateKey
	BlockTime         time.Duration
	DecodeRPCFunc     DecodeRPCFunc
	RPCProcessor      RPCProcessor
	TransactionHasher core.Hasher[*core.Transaction]
	Transports        []Transport
}

type Server struct {
	ServerOpts
	isValidator bool
	blockchain  *core.Blockchain
	memPool     *TransactionPool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	s := &Server{
		ServerOpts:  opts,
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}),
	}

	if s.InfoLog == nil {
		s.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	}

	if s.ErrorLog == nil {
		s.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	}

	if s.BlockTime == 0 {
		s.BlockTime = defaultBlockTime
	}

	if s.DecodeRPCFunc == nil {
		s.DecodeRPCFunc = DefaultDecodeRPCFunc
	}

	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}

	if s.TransactionHasher == nil {
		s.TransactionHasher = core.TransactionHasher{}
	}

	blockchain, err := core.NewBlockchain(core.CreateGenesisBlock())
	if err != nil {
		return nil, err
	}
	s.blockchain = blockchain

	s.memPool = NewTransactionPool(10, s.TransactionHasher)

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (s *Server) Start() {
	s.initTransports()

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.DecodeRPCFunc(rpc)
			if err != nil {
				fmt.Println(err)
			}
			if err = s.ProcessRPCMessage(msg); err != nil {
				fmt.Println(err)
			}
		case <-s.quitCh:
			break free
		}
	}

	s.InfoLog.Println("Server shutdown")
}

func (s *Server) initTransports() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}

func (s *Server) validatorLoop() {
	for range time.Tick(defaultBlockTime) {
		if err := s.createNewBlock(); err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.blockchain.GetHeader(s.blockchain.Height())
	if err != nil {
		return err
	}

	txs := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txs)
	if err != nil {
		return err
	}

	if err = block.Sign(s.PrivateKey); err != nil {
		return err
	}

	if err = s.blockchain.AddBlock(block); err != nil {
		return err
	}

	s.memPool.ClearPending()

	go func() {
		err = s.broadcastBlock(block)
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func (s *Server) ProcessRPCMessage(msg *DecodedRPCMessage) error {
	switch payload := msg.Payload.(type) {
	case *core.Transaction:
		return s.HandleTransaction(payload)
	case *core.Block:
		return s.HandleBlock(payload)
	}
	return nil
}

func (s *Server) HandleTransaction(tx *core.Transaction) error {
	hash := tx.Hash(s.TransactionHasher)
	if s.memPool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	go func() {
		err := s.broadcastTransaction(tx)
		if err != nil {
			fmt.Println(err)
		}
	}()

	slog.Info("adding new transaction to mempool", "hash", hash, "mempool length", s.memPool.PendingCount())
	return s.memPool.Add(tx)
}

func (s *Server) HandleBlock(block *core.Block) error {
	if err := s.blockchain.AddBlock(block); err != nil {
		return err
	}

	go func() {
		err := s.broadcastBlock(block)
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func (s *Server) broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) broadcastTransaction(tx *core.Transaction) error {
	buf := new(bytes.Buffer)
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}
	msg := NewRPCMessage(MessageTypeTransaction, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastBlock(block *core.Block) error {
	buf := new(bytes.Buffer)
	if err := block.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}
	msg := NewRPCMessage(MessageTypeBlock, buf.Bytes())
	return s.broadcast(msg.Bytes())
}
