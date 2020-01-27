package audiolan

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"reflect"
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

func Equal(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (strm *AudioStream) sendData(in []float32) {
	buffer := make([]float32, SampleRate*1)
	count := 0
	for i := range buffer {
		buffer[i] = in[i]
		if buffer[i] == 0 {
			count++
		}
	}
	log.Println(count, " are zeros")

	buf := bytes.NewBuffer(make([]byte, SampleRate))
	err := binary.Write(buf, binary.BigEndian, buffer)
	if err != nil {
		fmt.Println("while converting buffer to binary", buf.Len(), err)
		if !strm.errorTracker.Add() {
			log.Println("too many errors, client is gone")
			strm.Stop()
		}
		return
	}

	chkbuf := []float32{}
	binary.Read(buf, binary.BigEndian, chkbuf)
	log.Println("are the same?", reflect.DeepEqual(buffer, chkbuf))
	log.Println("are the eql?", Equal(buffer, chkbuf))

	log.Println("sending message")
	err = strm.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
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

func (strm *AudioStream) Start() {
	log.Println("starting stream for", strm.ip)
	buffer := make([]float32, SampleRate*1)
	stream, err := portaudio.OpenDefaultStream(1, 0, SampleRate, len(buffer), strm.sendData)

	if err != nil {
		panic(err)
	}
	stream.Start()
	defer stream.Close()

	defer strm.conn.Close()
	strm.streaming = true
	<-strm.ctx.Done()
}

func (strm *AudioStream) IsStreaming() bool {
	return strm.streaming
}

func (strm *AudioStream) Stop() {
	if strm.cancel != nil {
		strm.cancel()
	}
}
