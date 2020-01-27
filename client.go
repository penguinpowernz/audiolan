package audiolan

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

const ClientRxPort = 3457

type Client struct {
	Connected      bool
	CurrentAddress string
	ConnectedAt    time.Time

	connMu  *sync.Mutex
	conn    *websocket.Conn
	bytesRx int
	cancel  func()
	onError func(err error)
}

func NewClient() *Client {
	cl := &Client{
		onError: func(err error) {},
		connMu:  new(sync.Mutex),
	}

	return cl
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

	cl.Connected = false
}

func (cl *Client) ConnectedTo(address string) bool {
	return cl.Connected && cl.CurrentAddress == address
}

func (cl *Client) ListenForAudio(ctx context.Context, conn *websocket.Conn) error {
	cl.conn = conn
	defer conn.Close()

	log.Println("listening for audio from", cl.conn.RemoteAddr())

	defer func() {
		log.Println("closing client WS listening port")
		err := cl.conn.Close()
		if err != nil {
			log.Println("failed to cleanly close client WS listening port:", err)
		}
	}()

	stream, err := portaudio.OpenDefaultStream(0, 1, SampleRate, SampleRate, cl.handleStream)
	if err != nil {
		log.Println("error when opening audio stream", err)
		cl.Disconnect()
		return nil
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Println("error when starting audio stream", err)
		cl.Disconnect()
		return nil
	}

	<-ctx.Done()
	log.Println("stopped listening for audio")

	return err
}

func (cl *Client) handleStream(out []float32) {
	buffer := make([]float32, SampleRate*1)
	mtype, bindata, err := cl.conn.ReadMessage()
	if _, ok := err.(*websocket.CloseError); ok {
		log.Println("connection was closed")
		cl.Disconnect()
		return
	}

	if err != nil {
		log.Println("failed to read from WS:", err)
		return
	}

	if mtype != websocket.BinaryMessage {
		log.Println("ignoring non binary message")
		return
	}

	cl.bytesRx += len(bindata)
	log.Printf("received: %d bytes from %s\n", len(bindata), cl.conn.RemoteAddr())

	r := bytes.NewReader(bindata)
	if err := binary.Read(r, binary.BigEndian, &buffer); err != nil {
		log.Println("failed to read the buffer from binary", err)
		return
	}

	log.Println("buffer length ", len(buffer))

	count := 0
	for i := range out {
		out[i] = buffer[i]
		if buffer[i] == 0 {
			count++
		}
	}

	log.Println(count, " are zeros")
}

func (cl *Client) ConnectTo(address string) {
	cl.connMu.Lock()
	defer cl.connMu.Unlock()

	fmt.Println("client connecting to", address)

	log.Println("asking server to send us audio")
	dialer := websocket.DefaultDialer
	dialer.ReadBufferSize = SampleRate

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
