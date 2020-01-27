package audiolan

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const ClientRxPort = 3457

type Client struct {
	Connected      bool
	CurrentAddress string
	ConnectedAt    time.Time

	conn    *net.UDPConn
	bytesRx int
	cancel  func()
	onError func(err error)
}

func NewClient() *Client {
	return &Client{
		onError: func(err error) {},
	}
}

func (cl *Client) BytesReceived() float64 {
	return float64(cl.bytesRx)
}

func (cl *Client) OnError(cb func(error)) {
	cl.onError = cb
}

func (cl *Client) Disconnect() {
	log.Println("client disconnecting")
	if cl.cancel != nil {
		cl.cancel()
	}

	cl.CurrentAddress = ""
	cl.Connected = false
}

func (cl *Client) ConnectedTo(address string) bool {
	return cl.Connected && cl.CurrentAddress == address
}

func (cl *Client) ListenForAudio(ctx context.Context) error {
	var err error
	cl.conn, err = net.ListenUDP("udp", &net.UDPAddr{
		Port: ClientRxPort,
		IP:   net.ParseIP("0.0.0.0"),
	})

	if err != nil {
		return err
	}

	fmt.Printf("listening for audio at %s\n", cl.conn.LocalAddr().String())

	go func() {
		defer func() {
			log.Println("closing client UDP listening port")
			err := cl.conn.Close()
			if err != nil {
				log.Println("failed to cleanly close client UDP listening port:", err)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("stopped listening for audio")
				return
			default:
				message := make([]byte, 20)
				rlen, remote, err := cl.conn.ReadFromUDP(message[:])
				if err != nil {
					log.Println("failed to read from UDP port:", err)
					time.Sleep(time.Second / 5)
					continue
				}

				cl.bytesRx += rlen
				data := strings.TrimSpace(string(message[:rlen]))
				fmt.Printf("received: %s from %s\n", data, remote)
			}
		}
	}()

	return nil
}

func (cl *Client) ConnectTo(address string) {
	fmt.Println("client connecting to", address)

	ctx, cancel := context.WithCancel(context.Background())
	cl.cancel = cancel
	if err := cl.ListenForAudio(ctx); err != nil {
		log.Println("failed to start listening for audio:", err)
		return
	}

	log.Println("asking server to send us audio")
	res, err := http.Get("http://" + address + ":3456/connect")
	if err != nil {
		log.Println("failed to connect", err)

		go cl.onError(err)

		cl.Disconnect()
		return
	}

	if res.StatusCode == 200 {
		log.Println("connected successfully")
	} else {
		log.Println("failed to connect, got status code of", res.StatusCode)
		cl.Disconnect()
		return
	}

	cl.Connected = true
	cl.CurrentAddress = address
	cl.ConnectedAt = time.Now()
}
