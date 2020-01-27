package audiolan

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

type AudioStream struct {
	ip           string
	connectedAt  time.Time
	bytesTx      int
	cancel       func()
	conn         *net.UDPConn
	errs         int
	ctx          context.Context
	streaming    bool
	errorTracker *rateTrack
}

func NewAudioStream(ip string) (*AudioStream, error) {
	addr := ip + ":" + strconv.Itoa(ClientRxPort)

	log.Println("transmitting audio to", addr)
	clientAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", localAddr, clientAddr)
	if err != nil {
		return nil, err
	}

	strm := new(AudioStream)
	strm.conn = conn
	strm.connectedAt = time.Now()
	strm.ip = ip
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
	defer strm.conn.Close()
	strm.streaming = true
	i := 0
	for {
		select {
		case <-strm.ctx.Done():
			strm.streaming = false
			log.Println("stream stopped for", strm.ip)
			return
		default:
			time.Sleep(time.Second / 5)
			msg := strconv.Itoa(i)
			i++
			buf := []byte(msg)
			n, err := strm.conn.Write(buf)
			strm.bytesTx += n
			if err != nil {
				fmt.Println(msg, err, strm.errs)
				if !strm.errorTracker.Add() {
					log.Println("too many errors, client is gone")
					strm.Stop()
				}
				continue
			}
			strm.errs = 0
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
