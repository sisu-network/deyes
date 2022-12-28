package lisk

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	lisk "github.com/sisu-network/deyes/chains/lisk/types"
	"github.com/sisu-network/lib/log"
	"net/url"
)

// LiskClient A wrapper around socket so that we can mock in watcher tests.
type LiskClient interface {
	Close()
	WriteMessage(messageType int, data []byte)
	ReadMessage()
	GetTransaction() chan *lisk.Payload
}

// defaultLiskClient
type defaultLiskClient struct {
	socket    *websocket.Conn
	payloadCh chan *lisk.Payload
}

// NewLiskClients Create new lisk client
func NewLiskClients(wss []string) []LiskClient {
	clients := make([]LiskClient, 0)

	for _, ws := range wss {
		client, err := dial(ws)
		if err == nil {
			clients = append(clients, client)
			log.Infof("Adding lisk client at ws: ", ws)
		}
	}

	return clients
}

// dial: Init socket
func dial(ws string) (LiskClient, error) {
	var addr = flag.String("addr", ws, "http service address")
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Infof("connecting to %s", u.String())

	socket, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Errorf("dial:", err)
	}

	payloadCh := make(chan *lisk.Payload, 1000)

	go ReadMessage(socket, payloadCh)

	return &defaultLiskClient{
		socket:    socket,
		payloadCh: payloadCh,
	}, nil
}

// ReadMessage Read and select message
func ReadMessage(socket *websocket.Conn, payloadCh chan *lisk.Payload) {
	for {
		_, message, err := socket.ReadMessage()
		if err != nil {
			log.Infof("read:", err)
		}

		var payload lisk.Payload
		json.Unmarshal([]byte(message), &payload)
		if payload.Method == "app:transaction:new" {
			payloadCh <- &payload
		}
	}
}

func (c *defaultLiskClient) Close() {
	c.socket.Close()
}

func (c *defaultLiskClient) WriteMessage(messageType int, data []byte) {
	err := c.socket.WriteMessage(messageType, data)
	if err != nil {
		log.Infof("write:", err)
		return
	}
}

func (c *defaultLiskClient) ReadMessage() {
	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			log.Infof("read:", err)
		}

		var payload lisk.Payload
		json.Unmarshal([]byte(message), &payload)

	}
}

func (c *defaultLiskClient) GetTransaction() chan *lisk.Payload {
	return c.payloadCh
}
