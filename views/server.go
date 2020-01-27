package views

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/penguinpowernz/audiolan"
)

func NewServerView(svr *audiolan.Server) *ServerView {
	return &ServerView{svr: svr}
}

type ServerView struct {
	svr *audiolan.Server
}

func (sv *ServerView) Render() *fyne.Container {
	txtPort := widget.NewEntry()
	txtPort.SetText(":3456")

	btnConnect := widget.NewButton("Start", func() {})
	updateButton := func() {
		if sv.svr.IsListening() {
			txtPort.SetReadOnly(true)
			btnConnect.SetText("Stop")
		} else {
			btnConnect.SetText("Start")
			txtPort.SetReadOnly(false)
		}
	}

	btnConnect.OnTapped = func() {
		if sv.svr.IsListening() {
			log.Println("server is started, stopping")
			sv.svr.Stop()
		} else {
			log.Println("not started, starting")
			go sv.svr.Start(txtPort.Text)
		}
		updateButton()
	}

	lblStatus := widget.NewLabel("No Client")
	lblRx := widget.NewLabel("0 kB")

	go func() {
		rxkB := 0.0
		secs := 0.0

		for {
			updateButton()
			time.Sleep(time.Second)

			if sv.svr.HasClient() {
				ip := sv.svr.ClientIP()
				strm := sv.svr.GetStreamFor(ip)

				lblStatus.SetText(fmt.Sprintf("Client: %s", ip))
				rxkB = strm.BytesSent() / 1024
				secs = strm.ConnectedSecs()
			} else {
				rxkB = 0.0
				secs = 0.0
				lblStatus.SetText("No Client")
			}

			lblRx.SetText(fmt.Sprintf("%0.fs / %0.1f kB", secs, rxkB))
		}
	}()

	return fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		txtPort, btnConnect, lblStatus, lblRx)
}
