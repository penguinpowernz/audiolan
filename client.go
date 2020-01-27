package audiolan

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

const ClientRxPort = 3457

type Client struct {
	Connected      bool
	CurrentAddress string
	ConnectedAt    time.Time

	conn    *websocket.Conn
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

func (cl *Client) ListenForAudio(ctx context.Context, conn *websocket.Conn) error {
	cl.conn = conn

	log.Println("listening for audio from", cl.conn.RemoteAddr())

	defer func() {
		log.Println("closing client WS listening port")
		err := cl.conn.Close()
		if err != nil {
			log.Println("failed to cleanly close client WS listening port:", err)
		}
	}()

	buffer := make([]float32, SampleRate*1)
	stream, err := portaudio.OpenDefaultStream(0, 1, SampleRate, len(buffer), func(out []float32) {
		mtype, bindata, err := cl.conn.ReadMessage()
		if err != nil {
			log.Println("failed to read from WS:", err)
			time.Sleep(time.Second / 5)
			return
		}

		if mtype != websocket.BinaryMessage {
			log.Println("ignoring non binary message")
			return
		}

		cl.bytesRx += len(bindata)
		fmt.Printf("received: %d bytes from %s\n", len(bindata), conn.RemoteAddr())

		r := bytes.NewReader(bindata)
		binary.Read(r, binary.BigEndian, &buffer)
		for i := range out {
			out[i] = buffer[i]
		}
	})

	if err != nil {
		log.Println("error when opening audio stream", err)
		cl.Disconnect()

	}

	if err := stream.Start(); err != nil {
		log.Println("error when starting audio stream", err)
		cl.Disconnect()
	}

	log.Println("after stream start")

	defer stream.Close()
	<-ctx.Done()
	log.Println("stopped listening for audio")

	return err
}

func (cl *Client) ConnectTo(address string) {
	fmt.Println("client connecting to", address)

	log.Println("asking server to send us audio")
	dialer := websocket.DefaultDialer
	dialer.ReadBufferSize = SampleRate * 4

	conn, res, err := dialer.Dial("ws://"+address+":3456/connect", nil)
	if err != nil {
		log.Println("failed to connect", err)

		go cl.onError(err)

		cl.Disconnect()
		return
	}

	if res.StatusCode == 101 {
		log.Println("connected successfully")
	} else {
		log.Println("failed to connect, got status code of", res.StatusCode)
		cl.Disconnect()
		return
	}

	cl.Connected = true
	cl.CurrentAddress = address
	cl.ConnectedAt = time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	cl.cancel = cancel
	if err := cl.ListenForAudio(ctx, conn); err != nil {
		log.Println("failed to start listening for audio:", err)
		return
	}
}
