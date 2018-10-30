package main

import (
	"fmt"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/satori/go.uuid"
	"gopkg.in/urfave/cli.v1"
	"net/http"
	"os"
"github.com/pkg/errors"
	"time"
)

type network interface {
	Register(*UI, *cli.Context, socketIO)
}
type socketIO interface {
	On(string, interface{}) error
}

var link_id = uuid.Must(uuid.NewV4()).String()

func main() {
	app := cli.NewApp()
	app.Name = "Sportfun emulator"
	app.Description = "Sportfun Hardware and software emulator"

	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port, p",
			Usage: "communication port",
			Value: 8080,
		},
		cli.StringFlag{
			Name:  "host",
			Usage: "socket.io host for connection",
			Value: "localhost",
		},
		cli.BoolFlag{
			Name:  "server",
			Usage: "create socket.io server",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "hardware",
			Aliases: []string{"hdw"},
			Usage:   "emulate hardware commands",
			Action:  command(&hardware{}),
			Flags: []cli.Flag{
				&cli.Float64Flag{
					Name:  "speed, s",
					Usage: "starting RPM value",
					Value: 60,
				},
				&cli.DurationFlag{
					Name:  "refresh, r",
					Usage: "refreshing time between two RPM value",
					Value: 250*time.Millisecond,
				},
			},
		},
		{
			Name:   "game",
			Usage:  "emulate game commands",
			Action: command(&game{}),
		},
	}

	_ = app.Run(os.Args)
}

func command(network network) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		ui, err := NewUI()
		if err != nil {
			return cli.NewExitError(errors.Wrap(err, "failed to generate CLI"), 0x11)
		}

		var socket socketIO
		if ctx.Parent().Bool("server") {
			socket = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
		} else {
			url := gosocketio.GetUrl(ctx.Parent().String("host"), ctx.Parent().Int("port"), false)
			if socket, err = gosocketio.Dial(
				url,
				transport.GetDefaultWebsocketTransport(),
			); err != nil {
				return cli.NewExitError(errors.Wrapf(err, "failed to connect to %s", url), 0x8)
			}
		}

		network.Register(ui, ctx, socket)

		if ctx.Parent().Bool("server") {
			serveMux := http.NewServeMux()
			serveMux.Handle("/", socket.(*gosocketio.Server))
			go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", ctx.Parent().Int("port")), serveMux)
		}
		return ui.Run()
	}
}