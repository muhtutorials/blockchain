package network

import (
	"blockchain/core"
	"blockchain/types"
	"bytes"
	"encoding/hex"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
)

type APIConfig struct {
	ListenAddr string
	Logger     *slog.Logger
}

type API struct {
	APIConfig
	blockchain *core.Blockchain
	rpcCh      chan RPC
}

func NewAPI(cfg APIConfig, bc *core.Blockchain, rpcCh chan RPC) *API {
	s := &API{
		APIConfig:  cfg,
		blockchain: bc,
		rpcCh:      rpcCh,
	}

	if s.ListenAddr == "" {
		s.ListenAddr = ":8000"
	}

	return s
}

func (a *API) Start() {
	e := echo.New()

	e.GET("/block/:id", a.handleGetBlock)
	e.GET("/transaction/:hash", a.handleGetTransaction)
	e.POST("/transaction", a.handlePostTransaction)

	go func() {
		if err := e.Start(a.ListenAddr); err != nil {
			a.Logger.Error(err.Error(), "server address", a.ListenAddr)
		}
	}()
}

func (a *API) handleGetBlock(c echo.Context) error {
	var block *core.Block
	// id is block's height or hash
	id := c.Param("id")
	height, err := strconv.Atoi(id)
	if err == nil {
		block, err = a.blockchain.GetBlock(uint32(height))
		if err != nil {
			return c.JSON(http.StatusNotFound, ErrorRes{err.Error()})
		}
	} else {
		b, err := hex.DecodeString(id)
		if err != nil {
			return c.JSON(http.StatusNotFound, ErrorRes{err.Error()})
		}
		hash := types.HashFromBytes(b)
		block, err = a.blockchain.GetBlockByHeaderHash(hash)
		if err != nil {
			return c.JSON(http.StatusNotFound, ErrorRes{err.Error()})
		}
	}
	return c.JSON(http.StatusOK, ToBlockRes(block))
}

func (a *API) handleGetTransaction(c echo.Context) error {
	hashStr := c.Param("hash")
	b, err := hex.DecodeString(hashStr)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorRes{err.Error()})
	}
	hash := types.HashFromBytes(b)

	transaction, err := a.blockchain.GetTransaction(hash)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorRes{err.Error()})
	}
	return c.JSON(http.StatusOK, ToTransactionRes(transaction))
}

func (a *API) handlePostTransaction(c echo.Context) error {
	from, err := net.ResolveIPAddr("ip", c.Request().RemoteAddr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorRes{err.Error()})
	}

	buf := make([]byte, 1<<10)
	n, err := c.Request().Body.Read(buf)
	if err != nil && err != io.EOF {
		return c.JSON(http.StatusBadRequest, ErrorRes{err.Error()})
	}

	a.rpcCh <- RPC{
		From:    from,
		Payload: bytes.NewReader(buf[:n]),
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "transaction created successfully"})
}
