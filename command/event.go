package command

import (
	"fmt"
	"strings"

	"github.com/ckiely91/shellsim/screen"
	termbox "github.com/nsf/termbox-go"
)

type EventType uint8

const (
	EventTypeCommand EventType = iota
	EventTypeExit
)

type Event struct {
	Type EventType
	Text string
}

func EventLoop(state *State, screen *screen.Screen) {
mainLoop:
	for {
		evt, ok := <-state.EventChan
		if !ok {
			break
		}

		switch evt.Type {
		case EventTypeExit:
			break mainLoop
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
			screen.CurPath = state.CurrentDir.FullPath()
			screen.Redraw()
		}
	}
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