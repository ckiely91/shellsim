package main

import (
	"fmt"

	"github.com/ckiely91/shellsim/command"
	"github.com/ckiely91/shellsim/screen"

	termbox "github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	state := command.NewState()

	termbox.SetInputMode(termbox.InputEsc)

	screen := screen.NewScreen(fmt.Sprintf("%v:%v", state.CurrentHost.Hostname, state.CurrentDir.FullPath()))
	screen.Redraw()

	command.EventLoop(state, screen)
}
