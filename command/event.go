package command

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/ckiely91/shellsim/screen"
	termbox "github.com/nsf/termbox-go"
)

type EventType uint8

const (
	EventTypeCommand EventType = iota
	EventTypeLog
	EventTypeExit
	EventTypeArrowLeft
	EventTypeArrowRight
	EventTypeArrowUp
	EventTypeArrowDown
	EventTypeBackspace
	EventTypePaste
	EventTypeEnter
	EventTypeTab
	EventTypeChar
)

type Event struct {
	Type EventType
	Text string
}

func EventLoop(state *State, screen *screen.Screen) {
	go keyEventLoop(state.EventChan)
	mainEventLoop(state, screen)
}

func keyEventLoop(ch chan *Event) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowLeft:
				sendEvent(ch, EventTypeArrowLeft, "")
			case termbox.KeyArrowRight:
				sendEvent(ch, EventTypeArrowRight, "")
			case termbox.KeyArrowUp:
				sendEvent(ch, EventTypeArrowUp, "")
			case termbox.KeyArrowDown:
				sendEvent(ch, EventTypeArrowDown, "")
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				sendEvent(ch, EventTypeBackspace, "")
			case termbox.KeyCtrlC, termbox.KeyEsc:
				sendEvent(ch, EventTypeExit, "")
			case termbox.KeyCtrlV:
				sendEvent(ch, EventTypePaste, "")
			case termbox.KeyEnter:
				sendEvent(ch, EventTypeEnter, "")
			case termbox.KeyTab:
				sendEvent(ch, EventTypeTab, "")
			case termbox.KeySpace:
				sendEvent(ch, EventTypeChar, " ")
			default:
				sendEvent(ch, EventTypeChar, string(ev.Ch))
			}
		}
	}
}

func mainEventLoop(state *State, screen *screen.Screen) {
mainLoop:
	for {
		evt, ok := <-state.EventChan
		if !ok {
			break
		}

		switch evt.Type {
		case EventTypeCommand:
			state.CommandHistory.Append(evt.Text)
			screen.AppendLines(false, termbox.ColorDefault, fmt.Sprintf("%v > %v", screen.CurPath, evt.Text))
			inputCmd, args, err := readLine([]rune(evt.Text))
			if err != nil {
				screen.AppendLines(false, termbox.ColorRed, fmt.Sprintf("error: %v", err))
			} else {
				if cmd, ok := state.Commands[inputCmd]; ok {
					output, err := cmd.Execute(state, args...)
					if err != nil {
						screen.AppendLines(false, termbox.ColorRed, fmt.Sprintf("error: %v", err))
					} else if output != nil {
						screen.AppendLines(false, termbox.ColorWhite, strings.Split(string(output), "\n")...)
					}
				} else {
					screen.AppendLines(false, termbox.ColorRed, fmt.Sprintf("invalid command: %s", inputCmd))
				}
			}

			screen.SetEditLine(false, []rune{})
			// And set our current directory in case it changed
			screen.CurPath = fmt.Sprintf("%v:%v", state.CurrentHost.Hostname, state.CurrentDir.FullPath())
			screen.Redraw()
		case EventTypeLog:
			screen.AppendLines(true, termbox.ColorDefault, evt.Text)
		case EventTypeExit:
			break mainLoop
		case EventTypeArrowLeft:
			screen.MoveCursorLeft(true)
		case EventTypeArrowRight:
			screen.MoveCursorRight(true)
		case EventTypeArrowUp:
			screen.SetEditLine(true, []rune(state.CommandHistory.Up()))
		case EventTypeArrowDown:
			screen.SetEditLine(true, []rune(state.CommandHistory.Down()))
		case EventTypeBackspace:
			screen.BackspaceAtCursor(true)
		case EventTypePaste:
			text, _ := clipboard.ReadAll()
			screen.AppendAtCursor(true, []rune(text)...)
		case EventTypeTab:
			screen.SetEditLine(true, tabCompletion(state, screen.EditLine))
		case EventTypeEnter:
			if len(screen.EditLine) == 0 {
				break
			}
			go sendEvent(state.EventChan, EventTypeCommand, string(screen.EditLine))
		case EventTypeChar:
			screen.AppendAtCursor(true, []rune(evt.Text)...)
		}
	}
}

func sendEvent(ch chan *Event, evt EventType, text string) {
	ch <- &Event{Type: evt, Text: text}
}

func readLine(line []rune) (cmd string, args []string, err error) {
	curArg := ""
	quoteOpen := false

	for i := 0; i < len(line); i++ {
		c := line[i]

		switch c {
		case ' ':
			if quoteOpen {
				curArg += " "
			} else if curArg != "" {
				if cmd == "" {
					cmd = curArg
				} else {
					args = append(args, curArg)
				}
				curArg = ""
			}
		case '"':
			if quoteOpen {
				quoteOpen = false
				args = append(args, curArg)
				curArg = ""
			} else {
				quoteOpen = true
			}
		case '\\':
			// Read the next rune into curArg no matter what it is
			if i == len(line)-1 {
				return "", []string{}, fmt.Errorf("multiline not supported")
			}
			curArg += string(line[i+1])
			i++
		default:
			curArg += string(c)
		}
	}

	if quoteOpen {
		return "", []string{}, fmt.Errorf("invalid syntax: unterminated \"")
	}

	if curArg != "" {
		if cmd == "" {
			cmd = curArg
		} else {
			args = append(args, curArg)
		}
	}

	if cmd == "" {
		return "", []string{}, fmt.Errorf("empty command")
	}

	return cmd, args, nil
}
