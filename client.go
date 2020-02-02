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

// Client models and audiolan client that requests audio from a
// remote server and plays it locally
type Client struct {
	Connected      bool
	CurrentAddress string
	ConnectedAt    time.Time

	connMu  *sync.Mutex
	conn    *websocket.Conn
	bytesRx int
	cancel  func()
	onError func(err error)

	audioIn chan []float32
}

// NewClient will return a new client
func NewClient() *Client {
	cl := &Client{
		onError: func(err error) {},
		connMu:  new(sync.Mutex),
		audioIn: make(chan []float32, 10),
	}

	return cl
}

// BytesReceived will return how many bytes have been recieved so far
func (cl *Client) BytesReceived() float64 {
	return float64(cl.bytesRx)
}

// OnError will register a callback that is given an error whenever
// specific errors happen:
// - when connecting to a remote server fails
func (cl *Client) OnError(cb func(error)) {
	cl.onError = cb
}

// ConnectTo will connect the client to the given address
func (cl *Client) ConnectTo(address string) {
	cl.connMu.Lock()
	defer cl.connMu.Unlock()

	log.Printf("asking %s to send us audio", address)

	cl.CurrentAddress = address
	conn, err := cl.openSocket()
	if err != nil {
		log.Println("failed to connect", err)
		go cl.onError(err)
		cl.Disconnect()
	}

	cl.Connected = true
	cl.ConnectedAt = time.Now()
	log.Println("connected successfully")

	var ctx context.Context
	ctx, cl.cancel = context.WithCancel(context.Background())
	go cl.playAudio(ctx)

	if err := cl.listenForAudio(ctx, conn); err != nil {
		log.Println("failed to start listening for audio:", err)
	}

	cl.Disconnect()
}

func (cl *Client) openSocket() (*websocket.Conn, error) {
	dialer := websocket.DefaultDialer
	dialer.ReadBufferSize = FrameLength

	conn, res, err := dialer.Dial("ws://"+cl.CurrentAddress+":3456/connect", nil)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 101 {
		return nil, fmt.Errorf("bad status %d", res.StatusCode)
	}

	return conn, nil
}

// Disconnect will close all connections and stop playing audio
func (cl *Client) Disconnect() {
	log.Println("client disconnecting")
	if cl.cancel != nil {
		cl.cancel()
	}

	cl.Connected = false
}

// IsConnectedTo will return if the client is connected to the given address
func (cl *Client) IsConnectedTo(address string) bool {
	return cl.Connected && cl.CurrentAddress == address
}

// playAudio will setup list for any audio decoded by listenForAudio
func (cl *Client) playAudio(ctx context.Context) error {
	stream, err := portaudio.OpenDefaultStream(0, 1, SampleRate, FrameLength, cl.handleStream)
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
	defer stream.Stop()

	<-ctx.Done()
	return nil
}

// listenForAudio will start listening for audio coming over the websockets connection
func (cl *Client) listenForAudio(ctx context.Context, conn *websocket.Conn) error {
	cl.conn = conn
	defer conn.Close()

	log.Println("listening for audio from", cl.conn.RemoteAddr())

	for {
		if ctx.Err() != nil {
			break
		}

		mtype, bindata, err := cl.conn.ReadMessage()
		if _, ok := err.(*websocket.CloseError); ok {
			log.Println("connection was closed")
			return err
		}

		if err != nil {
			log.Println("failed to read from WS:", err)
			return err
		}

		if mtype != websocket.BinaryMessage {
			// log.Println("ignoring non binary message")
			continue
		}

		cl.bytesRx += len(bindata)
		log.Printf("received: %d bytes from %s\n", len(bindata), cl.conn.RemoteAddr())

		data, err := cl.decodeAudio(bindata)
		if err != nil {
			log.Println("failed to read the buffer from binary", err)
			continue
		}

		cl.audioIn <- data
	}

	log.Println("stopped listening for audio")
	return nil
}

func (cl *Client) decodeAudio(data []byte) ([]float32, error) {
	r := bytes.NewReader(data)
	audio := make([]float32, FrameLength)
	err := binary.Read(r, binary.BigEndian, &audio)
	return audio, err
}

// handleStream is the callback given to portaudio that decodes messages
// from the websockets connection and passes them up
func (cl *Client) handleStream(out []float32) {
	buffer := <-cl.audioIn

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
