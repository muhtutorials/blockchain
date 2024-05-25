package network

import (
	"blockchain/core"
	"blockchain/crypto"
	"bytes"
	"log/slog"
	"net"
	"os"
	"time"
)

var (
	defaultBlockTime = 5 * time.Second
)

type ServerOpts struct {
	ID                string
	Addr              string
	PrivateKey        *crypto.PrivateKey
	SeedNodes         []string
	Logger            *slog.Logger
	BlockTime         time.Duration
	DecodeRPCFunc     DecodeRPCFunc
	RPCProcessor      RPCProcessor
	TransactionHasher core.Hasher[*core.Transaction]
	Transport         *TCPTransport
}

type Server struct {
	ServerOpts
	isValidator bool
	blockchain  *core.Blockchain
	memPool     *TransactionPool
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	s := &Server{
		ServerOpts:  opts,
		isValidator: opts.PrivateKey != nil,
		quitCh:      make(chan struct{}),
	}

	if s.Logger == nil {
		handlerOpts := &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}
		s.Logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOpts))
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

	s.Transport = NewTCPTransport(s.Addr)

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
	err := s.Transport.Start()
	if err != nil {
		s.Logger.Error(err.Error(), "serverID", s.ID)
	}

	s.bootstrapNetwork()

free:
	for {
		select {
		case rpc := <-s.Transport.rpcCh:
			msg, err := s.DecodeRPCFunc(rpc)
			if err != nil {
				s.Logger.Error(err.Error(), "serverID", s.ID)
				continue
			}
			if err = s.ProcessRPCMessage(msg); err != nil {
				if err != core.ErrBlockAlreadyExists {
					s.Logger.Error(err.Error(), "serverID", s.ID)
				}
			}
		case peer := <-s.Transport.addPeerCh:
			err = s.sendStatusRequest(peer.conn.RemoteAddr())
			if err != nil {
				s.Logger.Error(err.Error(), "serverID", s.ID)
			}
		case <-s.quitCh:
			break free
		}
	}

	s.Logger.Info("server shutdown")
}

func (s *Server) bootstrapNetwork() {
	for _, addr := range s.SeedNodes {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			s.Logger.Error(err.Error(), "serverID", s.ID)
			continue
		}
		s.Transport.AddPeer(conn, false)
	}
}

func (s *Server) validatorLoop() {
	for range time.Tick(defaultBlockTime) {
		if err := s.createNewBlock(); err != nil {
			s.Logger.Error(err.Error(), "serverID", s.ID)
		}
	}
}

func (s *Server) createNewBlock() error {
	currentBlock, err := s.blockchain.GetBlock(s.blockchain.Height())
	if err != nil {
		return err
	}

	txs := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentBlock.Header, txs)
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
			s.Logger.Error(err.Error(), "serverID", s.ID)
		}
	}()

	return nil
}

// SyncBlocksLoop not used ATM
func (s *Server) SyncBlocksLoop(addr net.Addr) error {
	for range time.Tick(3 * time.Second) {
		req := &SyncBlocksRequest{
			FromHeight: s.blockchain.Height(),
		}
		buf := new(bytes.Buffer)
		if err := req.Encode(NewGobSyncBlocksRequestEncoder(buf)); err != nil {
			return err
		}
		rpcMessage := NewRPCMessage(MessageTypeSyncBlocksRequest, buf.Bytes())
		if err := s.Transport.SendMessage(addr, rpcMessage.Bytes()); err != nil {
			s.Logger.Error(err.Error(), "serverID", s.ID)
		}
	}
	return nil
}

func (s *Server) ProcessRPCMessage(msg *DecodedRPCMessage) error {
	switch payload := msg.Payload.(type) {
	case *core.Transaction:
		return s.receiveTransaction(payload)
	case *core.Block:
		return s.receiveBlock(payload)
	case *Status:
		return s.receiveStatus(msg.From, payload)
	case *SyncBlocksRequest:
		return s.receiveSyncBlocksRequest(msg.From, payload)
	case *Blocks:
		return s.receiveMissingBlocks(msg.From, payload)
	case *EmptyMessage:
		switch payload.Type {
		case MessageTypeStatusRequest:
			return s.receiveStatusRequest(msg.From)
		default:
			s.Logger.Error("Unknown message type", "serverID", s.ID)
		}
	default:
		s.Logger.Error("Unknown message type", "serverID", s.ID)
	}
	return nil
}

func (s *Server) broadcastTransaction(tx *core.Transaction) error {
	buf := new(bytes.Buffer)
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}
	rpcMessage := NewRPCMessage(MessageTypeTransaction, buf.Bytes())
	return s.Transport.Broadcast(rpcMessage.Bytes())
}

func (s *Server) receiveTransaction(tx *core.Transaction) error {
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
			s.Logger.Error(err.Error(), "serverID", s.ID)
		}
	}()

	s.Logger.Info("adding new transaction to mempool",
		"hash", hash, "mempool length", s.memPool.PendingCount())
	return s.memPool.Add(tx)
}

func (s *Server) broadcastBlock(block *core.Block) error {
	buf := new(bytes.Buffer)
	if err := block.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}
	rpcMessage := NewRPCMessage(MessageTypeBlock, buf.Bytes())
	return s.Transport.Broadcast(rpcMessage.Bytes())
}

func (s *Server) receiveBlock(block *core.Block) error {
	if err := s.blockchain.AddBlock(block); err != nil {
		return err
	}
	go func() {
		if err := s.broadcastBlock(block); err != nil {
			s.Logger.Error(err.Error(), "serverID", s.ID)
		}
	}()
	return nil
}

func (s *Server) sendStatusRequest(addr net.Addr) error {
	s.Logger.Info("sent status request", "serverID", s.ID, "to", addr)
	emptyMessage := &EmptyMessage{
		Type: MessageTypeStatusRequest,
	}
	buf := new(bytes.Buffer)
	if err := emptyMessage.Encode(NewGobEmptyMessageEncoder(buf)); err != nil {
		return err
	}
	rpcMessage := NewRPCMessage(MessageTypeStatusRequest, buf.Bytes())
	return s.Transport.SendMessage(addr, rpcMessage.Bytes())
}

func (s *Server) receiveStatusRequest(addr net.Addr) error {
	s.Logger.Info("received status request", "serverID", s.ID, "from", addr)
	status := &Status{
		ID:     s.ID,
		Height: s.blockchain.Height(),
	}
	buf := new(bytes.Buffer)
	if err := status.Encode(NewGobStatusEncoder(buf)); err != nil {
		return err
	}
	rpcMessage := NewRPCMessage(MessageTypeStatus, buf.Bytes())
	return s.Transport.SendMessage(addr, rpcMessage.Bytes())
}

func (s *Server) receiveStatus(addr net.Addr, status *Status) error {
	if status.Height <= s.blockchain.Height() {
		s.Logger.Info("no need to sync blocks with this node", "serverID", s.ID, "from", addr)
		return nil
	}
	req := &SyncBlocksRequest{
		FromHeight: s.blockchain.Height(),
	}
	buf := new(bytes.Buffer)
	if err := req.Encode(NewGobSyncBlocksRequestEncoder(buf)); err != nil {
		return err
	}
	rpcMessage := NewRPCMessage(MessageTypeSyncBlocksRequest, buf.Bytes())
	return s.Transport.SendMessage(addr, rpcMessage.Bytes())
}

func (s *Server) receiveSyncBlocksRequest(addr net.Addr, req *SyncBlocksRequest) error {
	s.Logger.Info("received sync blocks request", "serverID", s.ID, "from", addr)

	blocks := Blocks{}

	if req.FromHeight == 0 {
		req.FromHeight = 1
	}
	if req.ToHeight == 0 {
		req.ToHeight = s.blockchain.Height()
	}

	for i := req.FromHeight; i <= req.ToHeight; i++ {
		block, err := s.blockchain.GetBlock(i)
		if err != nil {
			return err
		}
		blocks = append(blocks, block)
	}

	buf := new(bytes.Buffer)
	if err := blocks.Encode(NewGobBlocksEncoder(buf)); err != nil {
		return err
	}
	rpcMessage := NewRPCMessage(MessageTypeMissingBlocks, buf.Bytes())
	return s.Transport.SendMessage(addr, rpcMessage.Bytes())
}

func (s *Server) receiveMissingBlocks(from net.Addr, blocks *Blocks) error {
	s.Logger.Info("received missing blocks", "serverID", s.ID, "from", from)
	for _, block := range *blocks {
		if err := s.blockchain.AddBlock(block); err != nil {
			return err
		}
	}
	return nil
}
