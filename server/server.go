package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/sisu-network/lib/log"

	"github.com/ethereum/go-ethereum/rpc"
)

type Server struct {
	handler       *rpc.Server
	listenAddress string
}

func NewServer(handler *rpc.Server, port int) *Server {
	return &Server{
		handler:       handler,
		listenAddress: fmt.Sprintf("0.0.0.0:%d", port),
	}
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		panic(err)
	}

	srv := &http.Server{Handler: s.handler}
	log.Info("Running server at", s.listenAddress)
	srv.Serve(listener)
}
