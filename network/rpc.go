package network

import (
	"blockchain/core"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	AddSource: true,
	Level:     slog.LevelDebug,
}))

type MessageType byte

const (
	MessageTypeTransaction MessageType = iota
	MessageTypeBlock
	MessageTypeStatusRequest
	MessageTypeStatus
	MessageTypeSyncBlocksRequest
	MessageTypeMissingBlocks
)

type RPCMessage struct {
	Header MessageType
	Body   []byte
}

func NewRPCMessage(t MessageType, body []byte) *RPCMessage {
	return &RPCMessage{
		Header: t,
		Body:   body,
	}
}

func (m *RPCMessage) Bytes() []byte {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(m)
	return buf.Bytes()
}

type RPC struct {
	From    net.Addr
	Payload io.Reader
}

type DecodedRPCMessage struct {
	From    net.Addr
	Payload any
}

type DecodeRPCFunc func(RPC) (*DecodedRPCMessage, error)

func DefaultDecodeRPCFunc(rpc RPC) (*DecodedRPCMessage, error) {
	msg := new(RPCMessage)
	if err := gob.NewDecoder(rpc.Payload).Decode(msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s: %s", rpc.From, err)
	}

	switch msg.Header {
	case MessageTypeTransaction:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTransactionDecoder(bytes.NewReader(msg.Body))); err != nil {
			return nil, err
		}
		return &DecodedRPCMessage{
			From:    rpc.From,
			Payload: tx,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Body))); err != nil {
			return nil, err
		}
		return &DecodedRPCMessage{
			From:    rpc.From,
			Payload: block,
		}, nil
	case MessageTypeStatusRequest:
		emptyMessage := new(EmptyMessage)
		if err := emptyMessage.Decode(NewGobEmptyMessageDecoder(bytes.NewReader(msg.Body))); err != nil {
			return nil, err
		}
		return &DecodedRPCMessage{
			From:    rpc.From,
			Payload: emptyMessage,
		}, nil
	case MessageTypeStatus:
		status := new(Status)
		if err := status.Decode(NewGobStatusDecoder(bytes.NewReader(msg.Body))); err != nil {
			return nil, err
		}
		return &DecodedRPCMessage{
			From:    rpc.From,
			Payload: status,
		}, nil
	case MessageTypeSyncBlocksRequest:
		req := new(SyncBlocksRequest)
		if err := req.Decode(NewGobSyncBlocksRequestDecoder(bytes.NewReader(msg.Body))); err != nil {
			return nil, err
		}
		return &DecodedRPCMessage{
			From:    rpc.From,
			Payload: req,
		}, nil
	case MessageTypeMissingBlocks:
		blocks := new(Blocks)
		if err := blocks.Decode(NewGobBlocksDecoder(bytes.NewReader(msg.Body))); err != nil {
			return nil, err
		}
		return &DecodedRPCMessage{
			From:    rpc.From,
			Payload: blocks,
		}, nil
	default:
		return nil, fmt.Errorf("invalid message header")
	}
}

type RPCProcessor interface {
	ProcessRPCMessage(*DecodedRPCMessage) error
}
