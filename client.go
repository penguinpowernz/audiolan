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

type Client struct {
	Connected      bool
	CurrentAddress string
	ConnectedAt    time.Time

	conn   *net.UDPConn
	cancel func()
}

func (cl *Client) Disconnect() {
	fmt.Println("client disconnecting")
	if cl.cancel != nil {
		cl.cancel()
	}
	cl.Connected = false
}

func (cl *Client) ConnectedTo(address string) bool {
	return cl.Connected && cl.CurrentAddress == address
}

func (cl *Client) ListenForAudio(ctx context.Context) {
	var err error
	cl.conn, err = net.ListenUDP("udp", &net.UDPAddr{
		Port: 3456,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		panic(err)
	}

	defer cl.conn.Close()
	fmt.Printf("client listening for audio %s\n", cl.conn.LocalAddr().String())

	for {
		select {
		case <-ctx.Done():
			cl.conn.Close()
			return
		default:
		}
		message := make([]byte, 20)
		rlen, remote, err := cl.conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}

		data := strings.TrimSpace(string(message[:rlen]))
		fmt.Printf("received: %s from %s\n", data, remote)
	}

}

func (cl *Client) ConnectTo(address string) {
	fmt.Println("client connecting to", address)

	ctx, cancel := context.WithCancel(context.Background())
	cl.cancel = cancel
	go cl.ListenForAudio(ctx)

	log.Println("requesting connection")
	res, err := http.Get("http://" + address + ":3456/connect")
	if err != nil {
		log.Println("failed to connect", err)
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
