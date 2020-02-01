package audiolan

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (svr *Server) startAPI(addr string) {
	r := gin.Default()

	r.GET("/connect", svr.handleConnect)
	r.GET("/disconnect", svr.handleDisconnect)

	svr.api = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := svr.api.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

func (svr *Server) stopAPI() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
	defer cancel()
	if err := svr.api.Shutdown(ctx); err != nil {
		log.Fatal("API stopped unexpectedly:", err)
	}

	<-ctx.Done()
	log.Println("API stopped")
}

func (svr *Server) handleConnect(c *gin.Context) {
	ip := strings.Split(c.Request.RemoteAddr, ":")[0]

	if strm, found := svr.streams[ip]; found {
		strm.Stop()
	}

	conn, err := websocket.Upgrade(c.Writer, c.Request, c.Writer.Header(), 1024, SampleRate)
	if err != nil {
		c.AbortWithError(500, err)
	}

	strm, err := NewAudioStream(conn)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	svr.streams[ip] = strm

	go strm.Start()

	c.Status(200)
}

func (svr *Server) handleDisconnect(c *gin.Context) {
	ip := strings.Split(c.Request.RemoteAddr, ":")[0]

	if strm, found := svr.streams[ip]; found {
		strm.Stop()
	} else {
		c.AbortWithStatus(404)
		return
	}

	c.Status(200)
}
