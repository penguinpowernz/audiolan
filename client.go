package audiolan

import (
	"fmt"
	"time"
)

type Client struct {
	Connected      bool
	CurrentAddress string
	ConnectedAt    time.Time
}

func (cl *Client) Disconnect() {
	fmt.Println("client disconnecting")
	cl.Connected = false
}

func (cl *Client) ConnectedTo(address string) bool {
	return cl.Connected && cl.CurrentAddress == address
}

func (cl *Client) ConnectTo(address string) {
	fmt.Println("client connecting to", address)
	cl.Connected = true
	cl.CurrentAddress = address
	cl.ConnectedAt = time.Now()
}
