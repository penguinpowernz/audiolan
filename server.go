package audiolan

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	running         bool
	clientConnected bool
	clientIP        string
	connectedAt     time.Time

	clientCancels map[string]func()
}

func NewServer() *Server {
	return &Server{clientCancels: map[string]func(){}}
}

func (svr *Server) IsListening() bool            { return svr.running }
func (svr *Server) HasClient() bool              { return svr.clientConnected }
func (svr *Server) ClientIP() string             { return svr.clientIP }
func (svr *Server) ClientConnectedAt() time.Time { return svr.connectedAt }

func (svr *Server) WaitForHandshake(addr string) {
	api := gin.Default()

	api.GET("/connect", func(c *gin.Context) {
		ctx, cancel := context.WithCancel(context.Background())
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]
		if cancel, _ := svr.clientCancels[ip]; cancel != nil {
			cancel()
		}
		svr.clientCancels[ip] = cancel
		svr.clientIP = ip
		svr.clientConnected = true
		svr.connectedAt = time.Now()
		err := svr.TransmitAudio(ctx, svr.clientIP)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Status(200)
	})

	api.GET("/disconnect", func(c *gin.Context) {
		ip := strings.Split(c.Request.RemoteAddr, ":")[0]
		if cancel, _ := svr.clientCancels[ip]; cancel != nil {
			cancel()
			delete(svr.clientCancels, ip)
		}

		svr.clientIP = ""
		svr.clientConnected = false
		c.Status(200)
	})

	api.Run(addr)
}

func (svr *Server) Start(addr string) {
	log.Println("starting on", addr)
	svr.running = true

	svr.WaitForHandshake(addr)
}

func (svr *Server) StopTransmitAudio(ip string) {
	if cancel, _ := svr.clientCancels[ip]; cancel != nil {
		cancel()
		delete(svr.clientCancels, ip)
	}
}

func (svr *Server) TransmitAudio(ctx context.Context, clientIP string) error {
	cl := clientIP + ":3456"
	log.Println("transmitting audio to", cl)
	clientAddr, err := net.ResolveUDPAddr("udp", cl)
	if err != nil {
		return err
	}

	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", localAddr, clientAddr)
	if err != nil {
		return err
	}

	go func() {
		errs := 0
		defer conn.Close()
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second * 1)
				msg := strconv.Itoa(i)
				i++
				buf := []byte(msg)
				_, err := conn.Write(buf)
				if err != nil {
					errs++
					fmt.Println(msg, err)
					if errs > 10 {
						log.Println("failed to transmit last 10 packets, client is gone")
						svr.StopTransmitAudio(clientIP)
					}
					continue
				}
				errs = 0
			}
		}
	}()

	return nil
}

func (svr *Server) Stop() {
	log.Println("stopping")
	for ip, cancel := range svr.clientCancels {
		log.Println("closing connection to", ip)
		cancel()
	}
	svr.running = false
	svr.clientIP = ""
	svr.clientConnected = false
}
