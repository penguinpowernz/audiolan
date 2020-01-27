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

const SampleRate = 44100

type AudioStream struct {
	ip           string
	connectedAt  time.Time
	bytesTx      int
	cancel       func()
	conn         *websocket.Conn
	ctx          context.Context
	streaming    bool
	errorTracker *rateTrack
}

func NewAudioStream(conn *websocket.Conn) (*AudioStream, error) {
	log.Println("transmitting audio to", conn.RemoteAddr())

	strm := new(AudioStream)
	strm.conn = conn
	strm.connectedAt = time.Now()
	strm.ip = conn.RemoteAddr().String()
	strm.errorTracker = newRateTrack(5, 10*time.Second)

	strm.ctx, strm.cancel = context.WithCancel(context.Background())

	return strm, nil
}

func (strm *AudioStream) BytesSent() float64 {
	return float64(strm.bytesTx)
}

func (strm *AudioStream) ConnectedSecs() float64 {
	return time.Since(strm.connectedAt).Seconds()
}

func (strm *AudioStream) Start() {

	buffer := make([]float32, SampleRate*1)
	stream, err := portaudio.OpenDefaultStream(1, 0, SampleRate, len(buffer), func(in []float32) {
		for i := range buffer {
			buffer[i] = in[i]
		}
	})

	if err != nil {
		panic(err)
	}
	stream.Start()
	defer stream.Close()

	defer strm.conn.Close()
	strm.streaming = true
	for {
		select {
		case <-strm.ctx.Done():
			strm.streaming = false
			log.Println("stream stopped for", strm.ip)
			return
		default:
			time.Sleep(time.Second / 5)

			buf := bytes.NewBuffer([]byte{})
			err := binary.Write(buf, binary.BigEndian, buffer)
			if err != nil {
				fmt.Println("while converting buffer to binary", len(buffer), err)
				if !strm.errorTracker.Add() {
					log.Println("too many errors, client is gone")
					strm.Stop()
				}
				continue
			}

			log.Println("sending message")
			err = strm.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
			if err != nil {
				fmt.Println("while writing message to websocket", len(buffer), err)
				if !strm.errorTracker.Add() {
					log.Println("too many errors, client is gone")
					strm.Stop()
				}
				continue
			}

			strm.bytesTx += len(buffer)
		}
	}
}

func (strm *AudioStream) IsStreaming() bool {
	return strm.streaming
}

func (strm *AudioStream) Stop() {
	if strm.cancel != nil {
		strm.cancel()
	}
}
