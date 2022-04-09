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

	//Time allowed to do a request and recive a response.
	requestWait = (writeWait * 11) / 10
)

type Exporter struct {
	conn      *websocket.Conn
	readChan  chan []byte
	writeChan chan []byte
	closed    chan struct{}
}

func NewExporter(conn *websocket.Conn) *Exporter {
	readChan := make(chan []byte)
	writeChan := make(chan []byte)
	closed := make(chan struct{})

	go func() {
		defer conn.Close()
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}

			select {
			case readChan <- message:
			case <-closed:
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			conn.Close()
		}()
		for {
			select {
			case message, ok := <-writeChan:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				conn.WriteMessage(websocket.TextMessage, message)
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-closed:
				return
			}
		}
	}()
	return &Exporter{
		conn:      conn,
		readChan:  readChan,
		writeChan: writeChan,
		closed:    closed,
	}
}

func (e *Exporter) Request(url string) ([]byte, error) {
	t := time.NewTimer(requestWait)
	select {
	case e.writeChan <- []byte(url):
		select {
		case data := <-e.readChan:
			return data, nil
		case <-t.C:
			return nil, errors.New("request timeout")
		case <-e.closed:
			return nil, errors.New("server closed")
		}
	case <-t.C:
		return nil, errors.New("request timeout")
	case <-e.closed:
		return nil, errors.New("server closed")
	}
}

func (e *Exporter) Close() {
	e.conn.Close()
	close(e.closed)
}

type ExporterPool struct {
	pool      []*Exporter
	addChan   chan *Exporter
	getChan   chan *Exporter
	closeChan chan bool
}

func NewExporterPool() *ExporterPool {
	pool := ExporterPool{
		pool:      make([]*Exporter, 0),
		addChan:   make(chan *Exporter),
		getChan:   make(chan *Exporter),
		closeChan: make(chan bool),
	}

	go func() {
		for {
			if len(pool.pool) > 0 {
				select {
				case exporter := <-pool.addChan:
					pool.pool = append(pool.pool, exporter)
				case pool.getChan <- pool.pool[0]:
					pool.pool = pool.pool[1:]
				case <-pool.closeChan:
					for _, exporter := range pool.pool {
						exporter.Close()
					}
				}
			} else {
				select {
				case exporter := <-pool.addChan:
					pool.pool = append(pool.pool, exporter)
				case <-pool.closeChan:
					for _, exporter := range pool.pool {
						exporter.Close()
					}
				}
			}
		}
	}()

	return &pool
}

func (pool *ExporterPool) Add(e *Exporter) {
	pool.addChan <- e
}

func (pool *ExporterPool) Get() (*Exporter, error) {
	select {
	case e := <-pool.getChan:
		return e, nil
	default:
		return nil, errors.New("get exporter timeout")
	}
}

func (pool *ExporterPool) Close() {
	pool.closeChan <- true
}
