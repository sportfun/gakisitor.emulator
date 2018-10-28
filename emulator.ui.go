package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"strings"
)

type CtrlKey struct {
		Key interface{}
		Modifier gocui.Modifier
		Handlers func(*gocui.Gui, *gocui.View) error
}
type UIKey struct {
	ShortName string
	LongName string

	Keys []CtrlKey
}

type Client struct {
	Id string
	Ip string
}
type UI struct {
	gui *gocui.Gui
	views struct{
		Command *gocui.View
		ClientList *gocui.View
		RequestMessages *gocui.View
		ResponseMessages *gocui.View
	}
	keys  []UIKey

	cache struct{
		allShortCommandSize int
		allLongCommandSize int
		nbClient int
	}
}

const (
	uiCommandListName = "cmd_list"
	uiClientListName    = "client_list"
	uiRequestMessageName = "request_message"
	uiResponseMessageName = "response_message"
)

var (
	ExitKey = UIKey{
		LongName: "Quit application",
		ShortName: "^Q/^C",
		Keys: []CtrlKey{
			{
				Key:      gocui.KeyCtrlQ,
				Modifier: gocui.ModNone,
				Handlers: func(ui *gocui.Gui, view *gocui.View) error { return gocui.ErrQuit },
			},
			{
				Key:      gocui.KeyCtrlC,
				Modifier: gocui.ModNone,
				Handlers: func(ui *gocui.Gui, view *gocui.View) error { return gocui.ErrQuit },
			},
		},
	}
)

// NewUI generates new UI instance
func NewUI() (*UI, error) {
	var err error

	ui := &UI{}
	ui.gui, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}

	ui.RegisterCommands(ExitKey)
	ui.gui.SetManagerFunc(func(gui *gocui.Gui) error { return ui.controllerView() })
	return ui, nil
}

// Display starts UI
func (ui *UI) Run() error {
	// Set keybindings
	for _, key := range ui.keys {
		for _, ctrlKey := range key.Keys {
			err := ui.gui.SetKeybinding("", ctrlKey.Key, ctrlKey.Modifier, ctrlKey.Handlers)
			if err != nil {
				return err
			}
		}
	}

	if err := ui.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

// RegisterCommands allows us to bind a new controller
func (ui *UI) RegisterCommands(keys ...UIKey) {
	for _, key := range keys {
		ui.cache.allShortCommandSize += len(key.ShortName)
		ui.cache.allLongCommandSize += len(key.LongName)
		ui.keys = append(ui.keys, key)
	}
}

// RefreshClients updates the client list view
func (ui *UI) RefreshClients(clients ...Client) {
	ui.cache.nbClient = len(clients)
	ui.gui.Update(func(gui *gocui.Gui) error {
		ui.views.ClientList.Clear()
		for _, client := range clients {
			fmt.Fprintf(ui.views.ClientList, "%s (%s)\n", client.Id, client.Ip)
		}
		return nil
	})
}

// AddRequestMessage writes a new request message
func (ui *UI) AddRequestMessage(message string) {
	ui.gui.Update(func(gui *gocui.Gui) error {
		fmt.Fprintln(ui.views.RequestMessages, message)
		return nil
	})
}

// AddRequestMessage writes a new response message
func (ui *UI) AddResponseMessage(message string) {
	ui.gui.Update(func(gui *gocui.Gui) error {
		fmt.Fprintln(ui.views.ResponseMessages, message)
		return nil
	})
}

// controllerView updates/generates the main ui
func (ui *UI) controllerView() error {
	var err error
	maxX, maxY := ui.gui.Size()

	// UI control list
	if ui.views.Command, err = ui.gui.SetView(uiCommandListName, 0, 0, maxX-1, 2); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	ui.views.Command.Clear()
	if maxX < (len(ui.keys) * 3) + ui.cache.allLongCommandSize + ui.cache.allShortCommandSize {
		spacer := strings.Repeat(" ", (maxX - ui.cache.allShortCommandSize) / len(ui.keys))
		for i := 0; i < len(ui.keys) - 1; i++ {
			fmt.Fprintf(ui.views.Command, "%s%s", ui.keys[i].ShortName, spacer)
		}
		fmt.Fprint(ui.views.Command, ui.keys[len(ui.keys)-1].ShortName)
	} else {
		spacer := strings.Repeat(" ", (maxX - ui.cache.allLongCommandSize) / len(ui.keys))
		for i := 0; i < len(ui.keys) - 1; i++ {
			fmt.Fprintf(ui.views.Command, "%s %s%s", ui.keys[i].ShortName, ui.keys[i].LongName, spacer)
		}
		fmt.Fprintf(ui.views.Command, "%s %s", ui.keys[len(ui.keys)-1].ShortName, ui.keys[len(ui.keys)-1].LongName)
	}

	// UI Client list view
	if ui.views.ClientList, err = ui.gui.SetView(uiClientListName, 0, 3, 35, maxY/2); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	ui.views.ClientList.Title = fmt.Sprintf("[ Client list (%d) ]", ui.cache.nbClient)
	ui.views.ClientList.Autoscroll = true
	ui.views.ClientList.Wrap = true

	// UI Server messages history view
	if ui.views.RequestMessages, err = ui.gui.SetView(uiRequestMessageName, 36, 3, maxX-1, maxY/2); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	ui.views.RequestMessages.Title = "[ Request messages ]"
	ui.views.RequestMessages.Autoscroll = true
	ui.views.RequestMessages.Wrap = true

	// UI Client messages history view
	if ui.views.ResponseMessages, err = ui.gui.SetView(uiResponseMessageName, 0, maxY/2+1, maxX-1, maxY-1); err != nil && err != gocui.ErrUnknownView {
		return err
	}
	ui.views.ResponseMessages.Title = "[ Response messages ]"
	ui.views.ResponseMessages.Autoscroll = true
	ui.views.ResponseMessages.Wrap = true

	return nil
}
