package main

import (
	"github.com/ckiely91/shellsim/command"
	"github.com/ckiely91/shellsim/screen"

	"github.com/atotto/clipboard"
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

	screen := screen.NewScreen(state.CurrentDir.FullPath())
	screen.Redraw()

	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyArrowLeft:
					screen.MoveCursorLeft(true)
				case termbox.KeyArrowRight:
					screen.MoveCursorRight(true)
				case termbox.KeyArrowUp:
					screen.SetEditLine(true, []rune(state.CommandHistory.Up()))
				case termbox.KeyArrowDown:
					screen.SetEditLine(true, []rune(state.CommandHistory.Down()))
				case termbox.KeyBackspace, termbox.KeyBackspace2:
					screen.BackspaceAtCursor(true)
				case termbox.KeyCtrlC, termbox.KeyEsc:
					state.EventChan <- &command.Event{Type: command.EventTypeExit}
				case termbox.KeyCtrlV:
					text, _ := clipboard.ReadAll()
					screen.AppendAtCursor(true, []rune(text)...)
				case termbox.KeyEnter:
					if len(screen.EditLine) == 0 {
						break
					}
					state.EventChan <- &command.Event{Type: command.EventTypeCommand, Text: string(screen.EditLine)}
				case termbox.KeySpace:
					screen.AppendAtCursor(true, ' ')
				default:
					screen.AppendAtCursor(true, ev.Ch)
				}
			}
		}
	}()

	command.EventLoop(state, screen)
}
