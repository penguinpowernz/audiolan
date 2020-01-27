package audiolan

import (
	"sync"
	"time"
)

type rateTrack struct {
	buf  chan bool
	mu   *sync.Mutex
	rate time.Duration
}

func newRateTrack(max int, rate time.Duration) *rateTrack {
	rt := &rateTrack{
		make(chan bool, max),
		new(sync.Mutex),
		rate,
	}
	go rt.drain()
	return rt
}

func (rt *rateTrack) drain() {
	for {
		<-rt.buf
		time.Sleep(rt.rate)
	}
}

func (rt *rateTrack) Add() bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.buf) == cap(rt.buf) {
		return false
	}

	rt.buf <- true
	return true
}
