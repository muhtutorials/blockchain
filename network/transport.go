package network

type NetAddr string

type Transport interface {
	Addr() NetAddr
	Connect(Transport) error
	Consume() <-chan RPC
	SendMessage(NetAddr, []byte) error
	Broadcast([]byte) error
}
