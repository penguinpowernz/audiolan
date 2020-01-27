package audiolan

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	running bool
	streams map[string]*AudioStream
}

func NewServer() *Server {
	return &Server{streams: map[string]*AudioStream{}}
}

func (svr *Server) IsListening() bool { return svr.running }
func (svr *Server) HasClient() bool   { return len(svr.streams) > 0 }
func (svr *Server) ClientIP() string {
	for ip := range svr.streams {
		return ip
	}
	return ""
}
func (svr *Server) ClientConnectedAt() time.Time {
	for _, s := range svr.streams {
		return s.connectedAt
	}
	return time.Time{}
}

func (svr *Server) StartAPI(addr string) {
	api := gin.Default()

	api.GET("/connect", func(c *gin.Context) {
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]

		if strm, found := svr.streams[ip]; found {
			strm.Stop()
		}

		strm, err := NewAudioStream(ip)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		svr.streams[ip] = strm

		go strm.Start()

		c.Status(200)
	})

	api.GET("/disconnect", func(c *gin.Context) {
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]

		if strm, found := svr.streams[ip]; found {
			strm.Stop()
		} else {
			c.AbortWithStatus(404)
			return
		}

		c.Status(200)
	})

	api.Run(addr)
}

func (svr *Server) Start(addr string) {
	log.Println("starting on", addr)
	svr.running = true
	svr.StartAPI(addr)
}

func (svr *Server) Stop() {
	log.Println("stopping")
	for ip, strm := range svr.streams {
		log.Println("closing connection to", ip)
		strm.Stop()
	}
	svr.running = false
}
