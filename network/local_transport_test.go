package network

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnect(t *testing.T) {
	tr1 := NewLocalTransport("1")
	tr2 := NewLocalTransport("2")

	_ = tr1.Connect(tr2)
	_ = tr2.Connect(tr1)

	assert.Equal(t, tr1.peers[tr2.addr], tr2)
	assert.Equal(t, tr2.peers[tr1.addr], tr1)
}

func TestSendMessage(t *testing.T) {
	tr1 := NewLocalTransport("me")
	tr2 := NewLocalTransport("you")

	_ = tr1.Connect(tr2)
	_ = tr2.Connect(tr1)

	msg := []byte("hey")
	err := tr1.SendMessage(tr2.Addr(), msg)
	assert.Nil(t, err)

	rpc := <-tr2.Consume()
	assert.Equal(t, rpc.From, tr1.Addr())
	assert.Equal(t, rpc.Payload, bytes.NewReader(msg))
}

func TestBroadcast(t *testing.T) {
	tr1 := NewLocalTransport("me")
	tr2 := NewLocalTransport("you")
	tr3 := NewLocalTransport("other guy")

	_ = tr1.Connect(tr2)
	_ = tr1.Connect(tr3)

	msg := []byte("foo")
	err := tr1.Broadcast(msg)
	assert.Nil(t, err)

	rpc2 := <-tr2.Consume()
	rpc3 := <-tr3.Consume()

	assert.Equal(t, rpc2.Payload, bytes.NewReader(msg))
	assert.Equal(t, rpc3.Payload, bytes.NewReader(msg))
	assert.Equal(t, rpc2.From, tr1.Addr())
	assert.Equal(t, rpc3.From, tr1.Addr())
}
