package main

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/network"
	"blockchain/types"
	"blockchain/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"
)

func main() {
	privateKey := crypto.GeneratePrivateKey()
	localNode := makeServer(":3000", ":8000", privateKey, nil)
	go localNode.Start()

	time.Sleep(2 * time.Second)
	remoteNode1 := makeServer(":3001", ":8001", nil, []string{":3000"})
	go remoteNode1.Start()

	remoteNode2 := makeServer(":3002", ":8002", nil, []string{":3000"})
	go remoteNode2.Start()

	remoteNode3 := makeServer(":3003", ":8003", nil, []string{":3000"})
	time.Sleep(12 * time.Second)
	go remoteNode3.Start()

	// causes EOF error because connection is closed after sending a tx
	//err := sendTransactionViaTCP(":3000")
	//if err != nil {
	//	fmt.Println(err)
	//}

	//collectionOwnerPrivateKey := crypto.GeneratePrivateKey()
	//collHash, err := createCollection(collectionOwnerPrivateKey, "http://localhost:8000/transaction")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//

	//go func() {
	//	for range time.Tick(time.Second) {
	//		err := mintNFT(collectionOwnerPrivateKey, collHash, "http://localhost:8000/transaction")
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//	}
	//}()

	//if err := sendCoins(privateKey, "http://localhost:8000/transaction"); err != nil {
	//	panic(err)
	//}

	select {}
}

func makeServer(addr, apiAddr string, pk *crypto.PrivateKey, seedNodes []string) *network.Server {
	opts := network.ServerOpts{
		Addr:       addr,
		APIAddr:    apiAddr,
		PrivateKey: pk,
		SeedNodes:  seedNodes,
	}
	server, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return server
}

func sendTransactionViaTCP(addr string) error {
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

func sendTransactionViaHTTP(addr string) error {
	privateKey := crypto.GeneratePrivateKey()
	ins := new(core.Instr)
	ins.Add(6, 1).String("hey").Store().Get("hey")

	tx := core.NewTransaction(ins.Bytes())
	err := tx.Sign(privateKey)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err = tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewRPCMessage(network.MessageTypeTransaction, buf.Bytes())

	req, err := http.NewRequest("POST", addr, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	buf2 := make([]byte, 1<<10)
	n, err := res.Body.Read(buf2)
	fmt.Println(string(buf2[:n]))

	return nil
}

func sendCoins(priv *crypto.PrivateKey, addr string) error {
	toPrivateKey := crypto.GeneratePrivateKey()
	tx := core.NewTransaction(nil)
	tx.To = toPrivateKey.PublicKey()
	tx.Value = big.NewInt(1_000_000)

	if err := tx.Sign(priv); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewRPCMessage(network.MessageTypeTransaction, buf.Bytes())

	req, err := http.NewRequest("POST", addr, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return err
	}

	client := http.DefaultClient
	_, err = client.Do(req)
	return err
}

func sendNFTTransactionViaHTTP(addr string) error {
	privateKey := crypto.GeneratePrivateKey()

	tx := core.NewTransaction(nil)
	tx.Inner = core.Collection{
		MetaData: []byte("Some stuff"),
		Fee:      150,
	}

	if err := tx.Sign(privateKey); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewRPCMessage(network.MessageTypeTransaction, buf.Bytes())

	req, err := http.NewRequest("POST", addr, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	buf2 := make([]byte, 1<<10)
	n, err := res.Body.Read(buf2)
	fmt.Println(string(buf2[:n]))

	return nil
}

func createCollection(priv *crypto.PrivateKey, addr string) (types.Hash, error) {
	tx := core.NewTransaction(nil)
	tx.Inner = core.Collection{
		MetaData: []byte("Some stuff"),
		Fee:      150,
	}

	if err := tx.Sign(priv); err != nil {
		return types.Hash{}, err
	}

	buf := new(bytes.Buffer)
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return types.Hash{}, err
	}

	msg := network.NewRPCMessage(network.MessageTypeTransaction, buf.Bytes())

	req, err := http.NewRequest("POST", addr, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return types.Hash{}, err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return types.Hash{}, err
	}

	buf2 := make([]byte, 1<<10)
	n, err := res.Body.Read(buf2)
	fmt.Println(string(buf2[:n]))

	return tx.Hash(core.TransactionHasher{}), nil
}

func mintNFT(priv *crypto.PrivateKey, coll types.Hash, addr string) error {
	metaData := map[string]any{
		"power":  8,
		"health": 100,
		"color":  "green",
		"rare":   true,
	}
	metaBuf := new(bytes.Buffer)
	if err := json.NewEncoder(metaBuf).Encode(metaData); err != nil {
		return err
	}

	tx := core.NewTransaction(nil)
	tx.Inner = core.Mint{
		MetaData:        metaBuf.Bytes(),
		Fee:             150,
		NFT:             utils.RandomHash(),
		Collection:      coll,
		CollectionOwner: priv.PublicKey(),
	}

	if err := tx.Sign(priv); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewRPCMessage(network.MessageTypeTransaction, buf.Bytes())

	req, err := http.NewRequest("POST", addr, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	buf2 := make([]byte, 1<<10)
	n, err := res.Body.Read(buf2)
	fmt.Println(string(buf2[:n]))

	return nil
}
