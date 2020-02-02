package views

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

func newEntry() *entry {
	return &entry{Entry: widget.NewEntry()}
}

type entry struct {
	*widget.Entry
	onTypedKey func(*fyne.KeyEvent)
}

func (en *entry) SetOnTypedKey(cb func(*fyne.KeyEvent)) {
	en.onTypedKey = cb
}

func (en *entry) TypedKey(ev *fyne.KeyEvent) {
	en.Entry.TypedKey(ev)
	if en.onTypedKey != nil {
		en.onTypedKey(ev)
	}
}
