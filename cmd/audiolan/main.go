package main

import (
	"log"
	"os"

	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/gordonklaus/portaudio"
	"github.com/penguinpowernz/audiolan"
	"github.com/penguinpowernz/audiolan/views"
)

func main() {
	app := app.New()

	w := app.NewWindow("AudioLAN")

	portaudio.Initialize()
	defer portaudio.Terminate()

	client := audiolan.NewClient()
	svr := audiolan.NewServer()

	w.SetOnClosed(func() {
		log.Println("closed")
		svr.Stop()
		client.Disconnect()
		os.Exit(0)
	})

	w.SetContent(widget.NewTabContainer(
		widget.NewTabItem("Client", views.NewClientView(client).Render()),
		widget.NewTabItem("Server", views.NewServerView(svr).Render()),
	))

	w.ShowAndRun()
}
