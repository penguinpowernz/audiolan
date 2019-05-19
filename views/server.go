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

	btnConnect.OnTapped = func() {
		if sv.svr.IsListening() {
			log.Println("server is started, stopping")
			sv.svr.Stop()
			btnConnect.SetText("Start")
			txtPort.SetReadOnly(false)
		} else {
			log.Println("not started, starting")
			sv.svr.Start(txtPort.Text)
			txtPort.SetReadOnly(true)
			btnConnect.SetText("Stop")
		}
	}

	lblStatus := widget.NewLabel("No Client")
	lblRx := widget.NewLabel("0 kB")

	go func() {
		rxkB := 0.0
		secs := 0.0
		for {
			time.Sleep(time.Second)
			if sv.svr.HasClient() {
				lblStatus.SetText(fmt.Sprintf("Client: %s", sv.svr.ClientIP()))
				rxkB += 44.1
				secs = time.Since(sv.svr.ClientConnectedAt()).Seconds()
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
