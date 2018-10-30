package main

import (
	"fmt"
	"github.com/graarh/golang-socketio"
	"github.com/jroimartin/gocui"
	"gopkg.in/sportfun/gakisitor.v2/protocol/v1.0"
	"gopkg.in/urfave/cli.v1"
	"sync"
)


type game struct {
	*sync.RWMutex
	*UI
	clients []Client
	sockets []*gosocketio.Channel
}

var (
	moveRight = v1_0.CommandPacket{
		Type: "game",
		LinkID: link_id,
		Body: struct {
			Command string `json:"command"`
			Args    []interface{} `json:"args"`
		}{Command: "start_game", Args: nil},
	}
	endGame = v1_0.CommandPacket{
		Type: "game",
		LinkID: link_id,
		Body: struct {
			Command string `json:"command"`
			Args    []interface{} `json:"args"`
		}{Command: "end_game", Args: nil},
	}
)

func (n *game) Register(ui *UI, _ *cli.Context, socket socketIO) {
	n.UI = ui
	n.RWMutex = &sync.RWMutex{}

	// Setup handler
	socket.On(gosocketio.OnConnection, func(c *gosocketio.Channel, a interface{}) {
		ui.gui.Update(func(gui *gocui.Gui) error {
			n.Lock()
			defer n.Unlock()

			// Log client connection
			ui.AddResponseMessage(fmt.Sprintf("client connection ('%s')", c.Id()))

			// Update client list
			n.clients = append(n.clients, Client{c.Id(), c.Ip()})
			n.sockets = append(n.sockets, c)
			ui.RefreshClients(n.clients...)
			return nil
		})
	})
	socket.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel, a interface{}) {
		ui.gui.Update(func(gui *gocui.Gui) error {
			n.Lock()
			defer n.Unlock()

			// Log client disconnection
			ui.AddResponseMessage(fmt.Sprintf("client disconnection ('%s')", c.Id()))

			// Find client index list
			var idx int
			var client Client
			for idx, client = range n.clients {
				if client.Id == c.Id() {
					break
				}
			}

			// Update client list
			n.clients[idx] = n.clients[len(n.clients)-1]
			n.clients = n.clients[:len(n.clients)-1]
			n.sockets[idx] = n.sockets[len(n.sockets)-1]
			n.sockets = n.sockets[:len(n.sockets)-1]
			ui.RefreshClients(n.clients...)
			return nil
		})
	})
	socket.On("command", func(c *gosocketio.Channel, a interface{}) {
		ui.AddResponseMessage(fmt.Sprintf("[%socket]{command}   \033[33m%#v\033[0m", c.Id(), a))
	})
	socket.On("data", func(c *gosocketio.Channel, a interface{}) {
		ui.AddResponseMessage(fmt.Sprintf("[%socket]{data}      \033[36m%#v\033[0m", c.Id(), a))
	})
	socket.On("error", func(c *gosocketio.Channel, a interface{}) {
		ui.AddResponseMessage(fmt.Sprintf("[%socket]{error}     \033[31m%#v\033[0m", c.Id(), a))
	})

	// Setup keybinding
	ui.RegisterCommands(
		UIKey{
			ShortName: "S^",
			LongName:  "Start acquisition session",
			Keys: []CtrlKey{
				{
					Key:      gocui.KeyCtrlS,
					Modifier: gocui.ModNone,
					Handlers: n.startAcquisitionSession,
				},
			},
		},
		UIKey{
			ShortName: "X^",
			LongName:  "Stop acquisition session",
			Keys: []CtrlKey{
				{
					Key:      gocui.KeyCtrlX,
					Modifier: gocui.ModNone,
					Handlers: n.stopAcquisitionSession,
				},
			},
		},
		UIKey{
			ShortName: "D^",
			LongName:  "Disconnect client",
			Keys: []CtrlKey{
				{
					Key:      gocui.KeyCtrlD,
					Modifier: gocui.ModNone,
					Handlers: n.disconnect,
				},
			},
		},
	)
}

func (n *game) startAcquisitionSession(gui *gocui.Gui, view *gocui.View) error {
	n.RLock()
	defer n.RUnlock()

	for _, socket := range n.sockets {
		n.AddRequestMessage(fmt.Sprintf("%s > start acquisition session", socket.Id()))
		socket.Emit("command", moveRight)
	}
	return nil
}

func (n *game) stopAcquisitionSession(gui *gocui.Gui, view *gocui.View) error {
	n.RLock()
	defer n.RUnlock()

	for _, socket := range n.sockets {
		n.AddRequestMessage(fmt.Sprintf("%s > stop session", socket.Id()))
		socket.Emit("command", endGame)
	}
	return nil
}

func (n *game) disconnect(gui *gocui.Gui, view *gocui.View) error {
	n.RLock()
	defer n.RUnlock()

	for _, socket := range n.sockets {
		n.AddRequestMessage(fmt.Sprintf("%s > start session", socket.Id()))
		socket.Close()
	}
	return nil
}