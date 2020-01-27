package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/penguinpowernz/audiolan"
)

func main() {
	var addr string
	var serve bool
	flag.BoolVar(&serve, "s", true, "serve audio from this computer to network clients")
	flag.StringVar(&addr, "c", "", "IP address to listen for audio from")
	flag.Parse()

	if !serve && addr == "" {
		log.Println("must use as client or server")
		os.Exit(1)
	}

	portaudio.Initialize()
	defer portaudio.Terminate()

	switch {
	case addr != "":
		for {
			client := audiolan.NewClient()
			client.ConnectTo(addr)
			time.Sleep(time.Second)
		}

	case serve:
		for {
			svr := audiolan.NewServer()
			svr.Start("0.0.0.0:3456")
		}
	}
}
