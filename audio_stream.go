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

// SampleRate is the sample rate used when reading or writing audio
const SampleRate = 44100

// FrameLength is how many samples are sent in a single packet
const FrameLength = SampleRate * 1

// AudioStream models a stream of audio across a websockets connection
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

// NewAudioStream creates a new AudioStream over the given websockets connection
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

// BytesSent will return how many bytes have been sent so far
func (strm *AudioStream) BytesSent() float64 {
	return float64(strm.bytesTx)
}

// ConnectedSecs returns how long the stream has been open for
func (strm *AudioStream) ConnectedSecs() float64 {
	return time.Since(strm.connectedAt).Seconds()
}

func (strm *AudioStream) encode(in []float32) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.BigEndian, in)
	return buf.Bytes(), err
}

func (strm *AudioStream) sendData(in []float32) {
	buffer := make([]float32, FrameLength)
	count := 0
	for i := range buffer {
		buffer[i] = in[i]
		if buffer[i] == 0 {
			count++
		}
	}
	log.Println(count, " are zeros")

	if count == len(buffer) {
		return
	}

	data, err := strm.encode(buffer)
	if err != nil {
		log.Println("encoding data failed:", err)
		return
	}

	// Verify(buffer, data)

	log.Println("sending message")
	err = strm.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		fmt.Println("while writing message to websocket", len(buffer), err)
		if !strm.errorTracker.Add() {
			log.Println("too many errors, client is gone")
			strm.Stop()
		}
		return
	}

	strm.bytesTx += len(buffer)
}

// Start starts streaming audio across the websockets connection
func (strm *AudioStream) Start() {
	log.Println("starting stream for", strm.ip)

	stream, err := portaudio.OpenDefaultStream(1, 0, SampleRate, FrameLength, strm.sendData)
	if err != nil {
		panic(err)
	}

	if err := stream.Start(); err != nil {
		panic(err)
	}

	defer stream.Stop()
	defer stream.Close()
	defer strm.conn.Close()

	strm.streaming = true
	<-strm.ctx.Done()
	strm.streaming = false
}

// IsStreaming returns true if the stream is currently streaming
func (strm *AudioStream) IsStreaming() bool {
	return strm.streaming
}

// Stop will stop the stream
func (strm *AudioStream) Stop() {
	if strm.cancel != nil {
		strm.cancel()
	}
}
