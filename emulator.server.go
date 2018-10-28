package main

import (
	"fmt"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/jroimartin/gocui"
	"gopkg.in/urfave/cli.v1"
	"net/http"
	"sync"
)

type game struct {
	*sync.RWMutex
	*UI
	clients []Client
	sockets []*gosocketio.Channel
}

func (n *game) Register(ui *UI, c *cli.Context) {
	n.UI = ui
	socket := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	n.RWMutex = &sync.RWMutex{}

	// Setup handler
	socket.On(gosocketio.OnConnection, func(c *gosocketio.Channel, a interface{}) {
		ui.gui.Update(func(gui *gocui.Gui) error {
			n.Lock()
			defer n.Unlock()

			// Log client connection
			ui.AddResponseMessage(fmt.Sprintf("client connection ('%socket')", c.Id()))

			// Update client list
			n.clients = append(n.clients, Client{c.Id(), c.Ip()})
			ui.RefreshClients(n.clients...)
			return nil
		})
	})
	socket.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel, a interface{}) {
		ui.gui.Update(func(gui *gocui.Gui) error {
			n.Lock()
			defer n.Unlock()

			// Log client disconnection
			ui.AddResponseMessage(fmt.Sprintf("client disconnection ('%socket')", c.Id()))

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
					Handlers: n.startSession,
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
					Handlers: n.stopSession,
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

	// start game
	serveMux := http.NewServeMux()
	serveMux.Handle("/", socket)
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", c.Int("port")), serveMux)
}

func (n *game) startSession(gui *gocui.Gui, view *gocui.View) error {
	n.RLock()
	defer n.RUnlock()

	for _, socket := range n.sockets {
		n.AddRequestMessage(fmt.Sprintf("%s > start session", socket.Id()))
		socket.Emit("command", map[string]interface{}{
			"link_id": "00000000-0000-0000-00000000",
			"body": map[string]interface{}{
				"command": "start_game",
				"args":    []string{},
			},
		})
	}
	return nil
}

func (n *game) stopSession(gui *gocui.Gui, view *gocui.View) error {
	n.RLock()
	defer n.RUnlock()

	for _, socket := range n.sockets {
		n.AddRequestMessage(fmt.Sprintf("%s > stop session", socket.Id()))
		socket.Emit("command", map[string]interface{}{
			"link_id": "00000000-0000-0000-00000000",
			"body": map[string]interface{}{
				"command": "end_game",
				"args":    []string{},
			},
		})
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