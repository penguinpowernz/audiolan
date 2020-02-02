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
	cv := &ClientView{model: client}
	cv.txtAddr = newEntry()
	cv.txtAddr.SetPlaceHolder("Set server address")
	cv.lblStatus = widget.NewLabel("Disconnected")
	cv.lblTx = widget.NewLabel("0 kB")
	cv.btnConnect = widget.NewButton("Connect", cv.onConnectTapped)

	cv.txtAddr.SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyReturn && !cv.txtAddr.Disabled() {
			fmt.Println("tapping............")
			cv.btnConnect.OnTapped()
		}
	})

	cv.model.OnError(func(err error) {
		if strings.Contains(err.Error(), "connection refused") {
			cv.lblStatus.SetText("Server down")
		}
	})

	return cv
}

type ClientView struct {
	model      *audiolan.Client
	txtAddr    *entry
	lblStatus  *widget.Label
	lblTx      *widget.Label
	btnConnect *widget.Button
}

func (cv *ClientView) Render() *fyne.Container {
	c := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		cv.txtAddr.Entry, cv.btnConnect, cv.lblStatus, cv.lblTx)

	go func() {
		for {
			cv.updateButton()
			time.Sleep(time.Second)
			cv.updateStatusBar()
		}
	}()

	return c
}

func (cv *ClientView) updateButton() {
	cv.btnConnect.Enable()
	if cv.model.IsConnectedTo(cv.txtAddr.Text) {
		cv.txtAddr.SetReadOnly(true)
		cv.btnConnect.SetText("Disconnect")
	} else {
		cv.btnConnect.SetText("Connect")
		cv.txtAddr.SetReadOnly(false)
	}
}

func (cv *ClientView) onConnectTapped() {
	if cv.txtAddr.Text == "" {
		cv.lblStatus.SetText("Invalid Address")
		return
	}

	if cv.model.IsConnectedTo(cv.txtAddr.Text) {
		log.Println("is connected, disconnecting")
		cv.model.Disconnect()
	} else {
		log.Println("not connected, connecting")
		go cv.model.ConnectTo(cv.txtAddr.Text)
	}

	cv.updateButton()
}

func (cv *ClientView) updateStatusBar() {
	txkB := 0.0
	if cv.model.Connected {
		cv.lblStatus.SetText(fmt.Sprintf("Connected for %0.fs", time.Since(cv.model.ConnectedAt).Seconds()))
		txkB = cv.model.BytesReceived() / 1024
	} else {
		txkB = 0
		cv.lblStatus.SetText("Disconnected")
	}

	cv.lblTx.SetText(fmt.Sprintf("%0.1f kB", txkB))
}
