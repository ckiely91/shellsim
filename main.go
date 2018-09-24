package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ckiely91/shellsim/command"

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

	screen := &Screen{
		Lines:      []Line{},
		Editing:    true,
		CurPath:    state.CurrentDir.FullPath(),
		CursorPosX: 0,
	}
	screen.Redraw()

	history := &CommandHistory{}

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowLeft:
				screen.MoveCursorLeft(true)
			case termbox.KeyArrowRight:
				screen.MoveCursorRight(true)
			case termbox.KeyArrowUp:
				screen.SetEditLine(true, []rune(history.Up()))
			case termbox.KeyArrowDown:
				screen.SetEditLine(true, []rune(history.Down()))
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				screen.BackspaceAtCursor(true)
			case termbox.KeyCtrlC, termbox.KeyEsc:
				os.Exit(0)
			case termbox.KeyCtrlV:
				text, _ := clipboard.ReadAll()
				screen.AppendAtCursor(true, []rune(text)...)
			case termbox.KeyEnter:
				if len(screen.EditLine) == 0 {
					break
				}
				history.Append(string(screen.EditLine))
				screen.AppendLines(false, termbox.ColorDefault, fmt.Sprintf("%v > %v", screen.CurPath, string(screen.EditLine)))
				inputCmd, args, err := readLine(screen.EditLine)
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
			case termbox.KeySpace:
				screen.AppendAtCursor(true, ' ')
			default:
				screen.AppendAtCursor(true, ev.Ch)
			}
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
