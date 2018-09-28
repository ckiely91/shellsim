package command

import (
	"fmt"
	"strings"
)

func tabCompletion(state *State, currentLine []rune) []rune {
	cmd, args, err := readLine(currentLine)
	if err != nil {
		// Just return the same line
		return currentLine
	}

	candidates := []string{}

	if len(args) == 0 {
		// Check if we can autocomplete the command
		for cmdName := range state.Commands {
			if strings.Index(cmdName, cmd) == 0 {
				candidates = append(candidates, cmdName)
			}
		}
	} else if theCmd, ok := state.Commands[cmd]; ok {
		arg := args[len(args)-1]
		for _, t := range theCmd.TabCompletionTypes {
			switch t {
			case TabCompletionTypeFile:
				for fileName := range state.CurrentDir.Files {
					if strings.Index(fileName, arg) == 0 {
						candidates = append(candidates, combineArgs(cmd, append(args[:len(args)-1], fileName)...))
					}
				}
			case TabCompletionTypeServer:
				for host := range state.CurrentHost.ConnectedHosts {
					if strings.Index(host, arg) == 0 {
						candidates = append(candidates, combineArgs(cmd, append(args[:len(args)-1], host)...))
					}
				}
			}
		}
	}

	if len(candidates) == 1 {
		return []rune(candidates[0])
	}

	return currentLine
}

func combineArgs(cmd string, args ...string) string {
	return fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
}
