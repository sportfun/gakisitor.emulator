package main

import (
	"encoding/json"
	"fmt"
	"github.com/graarh/golang-socketio"
	"github.com/jroimartin/gocui"
	"gopkg.in/sportfun/gakisitor.v2/protocol/v1.0"
	"gopkg.in/urfave/cli.v1"
	"sync"
	"time"
)

type hardware struct {
	*sync.RWMutex
	*UI
	clients []Client
	sockets []*gosocketio.Channel
}

var (
	moveLeft = v1_0.DataPacket{
		Type: "hardware",
		Body: struct {
			Module string `json:"module"`
			Value  interface{} `json:"value"`
		}{Module: "controller", Value: 0},
	}
	moveRight = v1_0.DataPacket{
		Type: "hardware",
		Body: struct {
			Module string `json:"module"`
			Value  interface{} `json:"value"`
		}{Module: "controller", Value: 1},
	}
	rpm = v1_0.DataPacket{
		Type: "hardware",
		Body: struct {
			Module string `json:"module"`
			Value  interface{} `json:"value"`
		}{Module: "rpm", Value: 0.},
	}

)

func (h *hardware) Register(ui *UI, ctx *cli.Context, socket socketIO) {
	h.UI = ui
	h.RWMutex = &sync.RWMutex{}
	rpm.Body.Value = ctx.Float64("speed")

	// Setup LinkId
	link.LinkID = ctx.Parent().String("id")
	link.Type = "hardware"
	moveLeft.LinkId = ctx.Parent().String("id")
	moveRight.LinkId = ctx.Parent().String("id")
	rpm.LinkId = ctx.Parent().String("id")

	// Setup handler
	socket.On(gosocketio.OnConnection, func(c *gosocketio.Channel, a interface{}) {
		ui.gui.Update(func(gui *gocui.Gui) error {
			h.Lock()
			defer h.Unlock()

			// Log client connection
			ui.AddResponseMessage(fmt.Sprintf("client connection ('%s')", c.Id()))

			// Update client list
			h.clients = append(h.clients, Client{c.Id(), c.Ip()})
			h.sockets = append(h.sockets, c)
			ui.RefreshClients(h.clients...)

			// Start link
			ui.AddRequestMessage(fmt.Sprintf("%s > start linkage", c.Id()))
			c.Emit("command", link)
			return nil
		})
	})
	socket.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel, a interface{}) {
		ui.gui.Update(func(gui *gocui.Gui) error {
			h.Lock()
			defer h.Unlock()

			// Log client disconnection
			ui.AddResponseMessage(fmt.Sprintf("client disconnection ('%s')", c.Id()))

			// Find client index list
			var idx int
			var client Client
			for idx, client = range h.clients {
				if client.Id == c.Id() {
					break
				}
			}

			// Update client list
			h.clients[idx] = h.clients[len(h.clients)-1]
			h.clients = h.clients[:len(h.clients)-1]
			h.sockets[idx] = h.sockets[len(h.sockets)-1]
			h.sockets = h.sockets[:len(h.sockets)-1]
			ui.RefreshClients(h.clients...)
			return cli.NewExitError("", 0)
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
			ShortName: "↑",
			LongName: "More speed",
			Keys: []CtrlKey{
				{
					Key: gocui.KeyArrowUp,
					Modifier:gocui.ModNone,
					Handlers: h.speedUp,
				},
			},
		},
		UIKey{
			ShortName: "￬",
			LongName: "Less speed",
			Keys: []CtrlKey{
				{
					Key: gocui.KeyArrowDown,
					Modifier:gocui.ModNone,
					Handlers: h.speedDown,
				},
			},
		},
		UIKey{
			ShortName: "￩",
			LongName: "Move left",
			Keys: []CtrlKey{
				{
					Key: gocui.KeyArrowLeft,
					Modifier:gocui.ModNone,
					Handlers: h.moveLeft,
				},
			},
		},
		UIKey{
			ShortName: "￫",
			LongName: "Move right",
			Keys: []CtrlKey{
				{
					Key: gocui.KeyArrowRight,
					Modifier:gocui.ModNone,
					Handlers: h.moveRight,
				},
			},
		},
	)

	go func() {
		for ;; {
			h.RLock()
			for _, socket := range h.sockets {
				h.AddRequestMessage(fmt.Sprintf("%s > send RPM (%v)", socket.Id(), rpm))
				socket.Emit("data", rpm)
			}
			h.RUnlock()
			time.Sleep(ctx.Duration("refresh"))
		}
	}()
}

func (h *hardware) moveLeft(gui *gocui.Gui, view *gocui.View) error {
	h.RLock()
	defer h.RUnlock()

	for _, socket := range h.sockets {
		b, _ := json.Marshal(moveLeft)
		h.AddRequestMessage(fmt.Sprintf("%s > move left (%s)", socket.Id(), string(b)))
		socket.Emit("data", moveLeft)
	}
	return nil
}
func (h *hardware) moveRight(gui *gocui.Gui, view *gocui.View) error {
	h.RLock()
	defer h.RUnlock()

	for _, socket := range h.sockets {
		b, _ := json.Marshal(moveRight)
		h.AddRequestMessage(fmt.Sprintf("%s > move right (%s)", socket.Id(), string(b)))
		socket.Emit("data", moveRight)
	}
	return nil
}
func (h *hardware) speedUp(gui *gocui.Gui, view *gocui.View) error {
	h.RLock()
	defer h.RUnlock()

	if len(h.sockets) > 0 {
		rpm.Body.Value = rpm.Body.Value.(float64) + 5.
	}
	for _, socket := range h.sockets {
		h.AddRequestMessage(fmt.Sprintf("%s > speed up to %f", socket.Id(), rpm.Body.Value))
	}
	return nil
}
func (h *hardware) speedDown(gui *gocui.Gui, view *gocui.View) error {
	h.RLock()
	defer h.RUnlock()

	if len(h.sockets) > 0 && rpm.Body.Value.(float64) > 0 {
		rpm.Body.Value = rpm.Body.Value.(float64) - 5.
	}
	for _, socket := range h.sockets {
		h.AddRequestMessage(fmt.Sprintf("%s > speed down to %f", socket.Id(), rpm.Body.Value))
	}
	return nil
}