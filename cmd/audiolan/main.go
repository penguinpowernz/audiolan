package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/penguinpowernz/audiolan"
	"github.com/penguinpowernz/audiolan/views"
)

func main() {
	app := app.New()

	w := app.NewWindow("AudioLAN")

	client := audiolan.NewClient()
	svr := audiolan.NewServer()

	w.SetContent(widget.NewTabContainer(
		widget.NewTabItem("Client", views.NewClientView(client).Render()),
		widget.NewTabItem("Server", views.NewServerView(svr).Render()),
	))

	w.ShowAndRun()
}
