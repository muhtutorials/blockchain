package main

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/network"
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func main() {
	localTransport := network.NewLocalTransport("local")
	remoteTransport1 := network.NewLocalTransport("remote1")
	remoteTransport2 := network.NewLocalTransport("remote2")
	remoteTransport3 := network.NewLocalTransport("remote3")

	err := localTransport.Connect(remoteTransport1)
	if err != nil {
		fmt.Println(err)
	}
	err = remoteTransport1.Connect(localTransport)
	if err != nil {
		fmt.Println(err)
	}
	err = remoteTransport1.Connect(remoteTransport2)
	if err != nil {
		fmt.Println(err)
	}
	err = remoteTransport2.Connect(remoteTransport3)
	if err != nil {
		fmt.Println(err)
	}

	initRemoteServers([]network.Transport{
		remoteTransport1,
		remoteTransport2,
		remoteTransport3,
	})

	go func() {
		for {
			if err := sendTransaction(remoteTransport1, localTransport.Addr()); err != nil {
				fmt.Println(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		time.Sleep(7 * time.Second)
		remoteLateTransport := network.NewLocalTransport("remoteLate")
		err = remoteTransport3.Connect(remoteLateTransport)
		if err != nil {
			fmt.Println(err)
		}
		remoteLateServer := makeServer("remoteLate", remoteLateTransport, nil)
		remoteLateServer.Start()
	}()

	privateKey := crypto.GeneratePrivateKey()
	server := makeServer("local", localTransport, &privateKey)
	server.Start()
}

func makeServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
	opts := network.ServerOpts{
		ID:         id,
		PrivateKey: pk,
		Transports: []network.Transport{tr},
	}
	server, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return server
}

func initRemoteServers(trs []network.Transport) {
	for i := 0; i < len(trs); i++ {
		id := fmt.Sprintf("remote_%d", i)
		server := makeServer(id, trs[i], nil)
		go server.Start()
	}
}

func sendTransaction(t network.Transport, to network.NetAddr) error {
	privateKey := crypto.GeneratePrivateKey()
	data := []byte(fmt.Sprintf("%d", rand.Intn(1000)))

	tx := core.NewTransaction(data)
	err := tx.Sign(privateKey)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err = tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewRPCMessage(network.MessageTypeTransaction, buf.Bytes())

	return t.SendMessage(to, msg.Bytes())

}
