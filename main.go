package main

type Network interface {
	Register(ui *UI)
}

func main() {
	ui, err := NewUI()
	if err != nil {
		panic(err)
	}

	n := network{}
	n.Register(ui)

	if err := ui.Display(); err != nil {
		panic(err)
	}
}