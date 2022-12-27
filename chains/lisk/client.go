package lisk

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	lisk "github.com/sisu-network/deyes/chains/lisk/types"
	"github.com/sisu-network/lib/log"
)

// LiskClient A wrapper around socket so that we can mock in watcher tests.
type LiskClient interface {
	Close()
	WriteMessage(messageType int, data []byte)
	ReadMessage()
	UpdateSocket()
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

// UpdateSocket  A function to keep socket alive and  listen system event when socket interrupt
func (c *defaultLiskClient) UpdateSocket() {
	defer c.socket.Close()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			break
		case t := <-ticker.C:
			err := c.socket.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				panic(fmt.Errorf("write %v", err))
			}
		case <-interrupt:
			log.Infof("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				panic(fmt.Errorf("write close %v", err))

			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			panic(fmt.Errorf("write close %v", err))
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
