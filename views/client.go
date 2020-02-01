package views

import (
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/penguinpowernz/audiolan"
)

func NewClientView(client *audiolan.Client) *ClientView {
	return &ClientView{cl: client}
}

type ClientView struct {
	cl      *audiolan.Client
	boxAddr *widget.Entry
}

func (cv *ClientView) Render() *fyne.Container {
	cv.boxAddr = widget.NewEntry()
	cv.boxAddr.SetPlaceHolder("Set server address")

	lblStatus := widget.NewLabel("Disconnected")
	lblTx := widget.NewLabel("0 kB")

	btnConnect := widget.NewButton("Connect", func() {})
	updateButton := func() {
		if cv.cl.IsConnectedTo(cv.boxAddr.Text) {
			cv.boxAddr.SetReadOnly(true)
			btnConnect.SetText("Disconnect")
		} else {
			btnConnect.SetText("Connect")
			cv.boxAddr.SetReadOnly(false)
		}
	}

	btnConnect.OnTapped = func() {
		if cv.boxAddr.Text == "" {
			lblStatus.SetText("Invalid Address")
			return
		}

		if cv.cl.IsConnectedTo(cv.boxAddr.Text) {
			log.Println("is connected, disconnecting")
			cv.cl.Disconnect()
		} else {
			log.Println("not connected, connecting")
			cv.cl.ConnectTo(cv.boxAddr.Text)
		}

		updateButton()
	}

	cv.cl.OnError(func(err error) {
		if strings.Contains(err.Error(), "connection refused") {
			lblStatus.SetText("Server down")
		}
	})

	go func() {
		txkB := 0.0
		for {
			updateButton()
			time.Sleep(time.Second)

			if cv.cl.Connected {
				lblStatus.SetText(fmt.Sprintf("Connected for %0.fs", time.Since(cv.cl.ConnectedAt).Seconds()))
				txkB = cv.cl.BytesReceived() / 1024
			} else {
				txkB = 0
				lblStatus.SetText("Disconnected")
			}

			lblTx.SetText(fmt.Sprintf("%0.1f kB", txkB))
		}
	}()

	return fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		cv.boxAddr, btnConnect, lblStatus, lblTx)
}
