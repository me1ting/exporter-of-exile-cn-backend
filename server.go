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
	gateway  *Gateway
	upgrader websocket.Upgrader
	config   *Config
}

func NewServer(config *Config) *Server {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return &Server{
		gateway:  NewGateway(),
		upgrader: upgrader,
		config:   config,
	}
}

func (s *Server) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}

	client := NewClient(conn)
	s.gateway.register <- client
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.URL.String()
	log.Print(url)

	if strings.HasPrefix(url, "/ws") {
		s.ServeWs(w, r)
	} else if strings.HasPrefix(url, "/character-window/get-characters") ||
		strings.HasPrefix(url, "/account/view-profile/") ||
		strings.HasPrefix(url, "/character-window/get-passive-skills") ||
		strings.HasPrefix(url, "/character-window/get-items") {
		data, err := s.gateway.Request(url)
		if err != nil {
			log.Printf("gateway: %v", err)
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			return
		}
		w.Write(data)
	} else if url == "/pob/patch" {
		//https://stackoverflow.com/questions/49333264/request-header-field-content-type-is-not-allowed-by-access-control-allow-headers
		w.Header().Add("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
		} else if r.Method == http.MethodPost {
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
