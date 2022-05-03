package main

import (
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	//Time allowed to do a request and recive target response. Must be more than writeWait.
	requestWait = (writeWait * 11) / 10
)

type Client struct {
	conn      *websocket.Conn
	requests  chan []byte
	responses chan []byte
	close     chan struct{}
	closed    chan struct{}
}

func NewClient(conn *websocket.Conn) *Client {
	requests := make(chan []byte)
	responses := make(chan []byte)
	close := make(chan struct{})
	closed := make(chan struct{})

	c := &Client{
		conn:      conn,
		requests:  requests,
		responses: responses,
		close:     close,
		closed:    closed,
	}

	go c.waitToClose()
	go c.pumpRequests()
	go c.pumpResponses()

	return c
}

func (c *Client) pumpRequests() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	conn := c.conn
	for {
		select {
		case message, ok := <-c.requests:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.Close()
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("client: %v", err)
				c.Close()
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Close()
				return
			}
		case <-c.closed:
			return
		}
	}
}

func (c *Client) pumpResponses() {
	conn := c.conn
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("client: %v", err)
			}

			c.Close()
			break
		}

		select {
		case c.responses <- message:
		case <-c.closed:
			return
		}
	}
}

func (c *Client) waitToClose() {
	<-c.close
	close(c.closed)
	c.conn.Close()
}

//Notifie the client to close without blocking the current goroutine.
func (c *Client) Close() {
	select {
	case c.close <- struct{}{}:
	default:
	}
}

func (e *Client) Closed() bool {
	select {
	case <-e.closed:
		return true
	default:
		return false
	}
}

func (c *Client) Request(url string) ([]byte, error) {
	t := time.NewTimer(requestWait)
	select {
	case c.requests <- []byte(url):
		select {
		case data := <-c.responses:
			return data, nil
		case <-t.C:
			return nil, errors.New("request timeout")
		case <-c.closed:
			return nil, errors.New("client closed")
		}
	case <-t.C:
		return nil, errors.New("request timeout")
	case <-c.closed:
		return nil, errors.New("client closed")
	}
}

type Gateway struct {
	pool     []*Client
	register chan *Client
	require  chan *Client
	closed   chan bool
}

func NewGateway() *Gateway {
	g := Gateway{
		pool:     make([]*Client, 0),
		register: make(chan *Client),
		require:  make(chan *Client),
		closed:   make(chan bool),
	}

	go g.pump()

	return &g
}

func (g *Gateway) pump() {
	for {
		if len(g.pool) > 0 {
			select {
			case c := <-g.register:
				g.pool = append(g.pool, c)
			case g.require <- g.pool[0]:
				g.pool = g.pool[1:]
			case <-g.closed:
				for _, c := range g.pool {
					c.Close()
				}
				return
			}
		} else {
			select {
			case c := <-g.register:
				g.pool = append(g.pool, c)
			case <-g.closed:
				for _, c := range g.pool {
					c.Close()
				}
				return
			}
		}
	}
}

func (g *Gateway) Request(url string) ([]byte, error) {
	t := time.NewTimer(requestWait)

out:
	for {
		select {
		case c := <-g.require:
			if c.Closed() {
				continue out
			}
			resp, err := c.Request(url)
			if err != nil {
				log.Printf("gateway: %v", err)
				c.Close()
				continue out
			}

			g.register <- c

			return resp, nil
		case <-t.C:
			return nil, errors.New("request timeout")
		case <-g.closed:
			break out
		}
	}

	return nil, errors.New("gateway tiemout")
}

func (g *Gateway) Close() {
	close(g.closed)
}
