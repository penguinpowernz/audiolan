package audiolan

import (
	"log"
	"net/http"
	"time"
)

// Server will answer client requests for audio and do whats needed
// to provide audio to them
type Server struct {
	running bool
	streams map[string]*AudioStream
	api     *http.Server
}

// NewServer will create a new server
func NewServer() *Server {
	return &Server{streams: map[string]*AudioStream{}}
}

// GetStreamFor will return the AudioStream object for the given IP
func (svr *Server) GetStreamFor(ip string) *AudioStream { return svr.streams[ip] }

// IsListening will return true if the server is currently listening for requests
func (svr *Server) IsListening() bool { return svr.running }

// HasClient will return true if there are any clients connected
func (svr *Server) HasClient() bool { return len(svr.streams) > 0 }

// ClientIP will return the IP of the first client
func (svr *Server) ClientIP() string {
	for ip := range svr.streams {
		return ip
	}
	return ""
}

// ClientConnectedAt will return the connection time of the first client
func (svr *Server) ClientConnectedAt() time.Time {
	for _, s := range svr.streams {
		return s.connectedAt
	}
	return time.Time{}
}

// Start will start the server on the given address
func (svr *Server) Start(addr string) {
	log.Println("starting on", addr)
	svr.running = true
	svr.startAPI(addr)
}

// Stop will stop the server
func (svr *Server) Stop() {
	log.Println("stopping")
	for ip, strm := range svr.streams {
		log.Println("closing connection to", ip)
		strm.Stop()
	}
	svr.stopAPI()
	svr.running = false
}
