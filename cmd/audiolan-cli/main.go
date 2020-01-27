package main

import (
	"flag"
	"log"
	"os"

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

	if addr != "" {
		client := new(audiolan.Client)
		client.ConnectTo(addr)
		for {
		}
	}

	if serve {
		svr := audiolan.NewServer()
		svr.Start("0.0.0.0:3456")
		for {
		}
	}

}
