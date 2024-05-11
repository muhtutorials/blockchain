package network

import (
	"blockchain/core"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log/slog"
)

type MessageType byte

const (
	MessageTypeTransaction MessageType = iota
	MessageTypeBlock
	MessageTypeSyncBlocks
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
	From    NetAddr
	Payload io.Reader
}

type DecodedRPCMessage struct {
	From    NetAddr
	Payload any
}

type DecodeRPCFunc func(RPC) (*DecodedRPCMessage, error)

func DefaultDecodeRPCFunc(rpc RPC) (*DecodedRPCMessage, error) {
	msg := new(RPCMessage)
	if err := gob.NewDecoder(rpc.Payload).Decode(msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s: %s", rpc.From, err)
	}

	slog.Info("received new message", "from", rpc.From)

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
	default:
		return nil, fmt.Errorf("invalid message header")
	}
}

type RPCProcessor interface {
	ProcessRPCMessage(*DecodedRPCMessage) error
}
