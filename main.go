package main

import (
	"gopkg.in/urfave/cli.v1"
	"os"
)

type Network interface {
	Register(ui *UI, c* cli.Context)
}

func main() {
	app := cli.NewApp()
	app.Name = "Sportfun emulator"
	app.Description = "Sportfun Hardware and software emulator"

	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name: "port, p",
			Usage: "communication port",
			Value: 8080,
		},
	}
	app.Commands = []cli.Command{
		{
			Name: "hardware",
			Aliases: []string{"hdw"},
			Usage: "emulate hardware commands",
			Action: func(c *cli.Context) error {
				ui, err := NewUI()
				if err != nil {
					return cli.NewExitError(err.Error(), 5)
				}

				(&hardware{}).Register(ui, c)
				return ui.Run()
			},
		},
		{
			Name: "game",
			Usage: "emulate game commands",
			Action: func(c *cli.Context) error {
				ui, err := NewUI()
				if err != nil {
					return cli.NewExitError(err.Error(), 5)
				}

				(&game{}).Register(ui, c)
				return ui.Run()
			},
		},
	}

	_ = app.Run(os.Args)
}