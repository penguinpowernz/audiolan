package audiolan

import (
	"log"
	"time"
)

type Server struct {
	running         bool
	clientConnected bool
	clientIP        string
	connectedAt     time.Time
}

func (svr *Server) IsListening() bool            { return svr.running }
func (svr *Server) HasClient() bool              { return svr.clientConnected }
func (svr *Server) ClientIP() string             { return svr.clientIP }
func (svr *Server) ClientConnectedAt() time.Time { return svr.connectedAt }

func (svr *Server) Start(addr string) {
	log.Println("starting on", addr)
	svr.running = true

	go func() {
		time.Sleep(3 * time.Second)
		svr.clientIP = "10.0.0.1"
		svr.clientConnected = true
		svr.connectedAt = time.Now()
	}()
}

func (svr *Server) Stop() {
	log.Println("stopping")
	svr.running = false
	svr.clientIP = ""
	svr.clientConnected = false
}
