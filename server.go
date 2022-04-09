package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type APIResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Server struct {
	exporterPool *ExporterPool
	upgrader     websocket.Upgrader
	config       *Config
}

func NewServer(config *Config) *Server {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return &Server{
		exporterPool: NewExporterPool(),
		upgrader:     upgrader,
		config:       config,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.URL.String()
	log.Print(url)

	if strings.HasPrefix(url, "/ws") {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print(err)
			return
		}

		exporter := NewExporter(conn)
		s.exporterPool.Add(exporter)

	} else if strings.HasPrefix(url, "/character-window/get-characters") ||
		strings.HasPrefix(url, "/account/view-profile/") ||
		strings.HasPrefix(url, "/character-window/get-passive-skills") ||
		strings.HasPrefix(url, "/character-window/get-items") {
		exporter, err := s.exporterPool.Get()
		if err != nil {
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			return
		}
		data, err := exporter.Request(url)
		if err != nil {
			log.Printf("error: %v", err)
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			exporter.Close()
			return
		}
		s.exporterPool.Add(exporter)
		w.Write(data)
	} else if strings.HasPrefix(url, "/patch") {
		if r.Method == http.MethodPost {
			filePath := r.Form.Get("filePath")
			err := Patch(filePath, fmt.Sprintf("http://localhost:%v/", s.config.ListenPort))
			if err != nil {
				data, _ := json.Marshal(APIResp{Code: 400, Msg: err.Error()})
				w.Write(data)
				return
			}

			data, _ := json.Marshal(APIResp{Code: 200, Msg: "success"})
			w.Write(data)
			return
		}
	}
}
