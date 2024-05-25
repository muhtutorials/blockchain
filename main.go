package main

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/network"
	"bytes"
	"log"
	"net"
	"time"
)

func main() {
	privateKey := crypto.GeneratePrivateKey()
	localNode := makeServer("local_node", ":3000", privateKey, nil)
	go localNode.Start()

	time.Sleep(1 * time.Second)

	remoteNode1 := makeServer("remote_node_1", ":3001", nil, []string{":3000"})
	go remoteNode1.Start()

	remoteNode2 := makeServer("remote_node_2", ":3002", nil, []string{":3000"})
	go remoteNode2.Start()

	remoteNode3 := makeServer("remote_node_3", ":3003", nil, []string{":3000"})
	time.Sleep(12 * time.Second)
	go remoteNode3.Start()

	// causes EOF error because connection is closed after sending a tx
	//err := sendTransaction(":3000")
	//if err != nil {
	//	fmt.Println(err)
	//}

	select {}
}

func makeServer(id, addr string, pk *crypto.PrivateKey, seedNodes []string) *network.Server {
	opts := network.ServerOpts{
		ID:         id,
		Addr:       addr,
		PrivateKey: pk,
		SeedNodes:  seedNodes,
	}
	server, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return server
}

func sendTransaction(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	privateKey := crypto.GeneratePrivateKey()
	ins := new(core.Instr)
	ins.Add(2, 3).String("hey").Store().Get("hey")

	tx := core.NewTransaction(ins.Bytes())
	err = tx.Sign(privateKey)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err = tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewRPCMessage(network.MessageTypeTransaction, buf.Bytes())

	_, err = conn.Write(msg.Bytes())
	if err != nil {
		return err
	}

	return nil
}
