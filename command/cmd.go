package command

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ckiely91/shellsim/fs"
)

type Command struct {
	ShortHelp string
	LongHelp  string
	Execute   func(state *State, args ...string) ([]byte, error)
}

func standardCommands() map[string]*Command {
	return map[string]*Command{
		"help":  HelpCommand,
		"ls":    LSCommand,
		"mkdir": MKDIRCommand,
		"cd":    CDCommand,
		"exit":  ExitCommand,
		"rmdir": RMDIRCommand,
	}
}

var HelpCommand = &Command{
	ShortHelp: "Display help for installed commands",
	LongHelp: `Display help for installed commands. Optionally include the command name for additional information.
Usage: help [command]`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("must supply zero or one arguments")
		}
		if len(args) == 0 {
			buf := bytes.NewBufferString("Commands\n")
			longest := 0
			for cmdName := range state.Commands {
				if len(cmdName) > longest {
					longest = len(cmdName)
				}
			}

			for cmdName, cmd := range state.Commands {
				buf.WriteString("  ")
				buf.WriteString(cmdName)
				for i := 0; i < longest-len(cmdName); i++ {
					buf.WriteRune(' ')
				}
				buf.WriteString(fmt.Sprintf(" - %s\n", cmd.ShortHelp))
			}
			return buf.Bytes(), nil
		}

		cmd, ok := state.Commands[args[0]]
		if !ok {
			return nil, fmt.Errorf("unknown command: %s", args[0])
		}

		return []byte(cmd.LongHelp), nil
	},
}

var LSCommand = &Command{
	ShortHelp: "List files in the current directory",
	LongHelp: `List files in the current directory.
Usage: ls`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		dirNames := []string{}
		fileNames := []string{}
		if state.CurrentDir.Parent != nil {
			dirNames = append(dirNames, "..")
		}

		for _, f := range state.CurrentDir.Files {
			if f.Type() == fs.FileTypeDirectory {
				dirNames = append(dirNames, fmt.Sprintf("%s/", f.Name()))
			} else {
				fileNames = append(fileNames, f.Name())
			}
		}

		sort.Strings(dirNames)
		sort.Strings(fileNames)

		all := []string{}
		all = append(all, dirNames...)
		all = append(all, fileNames...)

		if len(all) == 0 {
			return []byte("No files in the current directory"), nil
		}

		return []byte(strings.Join(all, "\n")), nil
	},
}

var MKDIRCommand = &Command{
	ShortHelp: "Create a new directory",
	LongHelp: `Create a new directory. New folders can only use alphanumeric characters separated by dashes or underscores. No spaces.
Usage: mkdir [name]`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("must supply only one argument - the directory name")
		}

		dirName := args[0]
		if err := fs.ValidateDirName(dirName); err != nil {
			return nil, err
		}

		dirNameLower := strings.ToLower(dirName)
		if _, ok := state.CurrentDir.Files[dirNameLower]; ok {
			return nil, fmt.Errorf("file or directory with that name already exists")
		}

		state.CurrentDir.Files[dirNameLower] = &fs.Directory{
			Parent:  state.CurrentDir,
			DirName: dirName,
			Files:   map[string]fs.File{},
		}

		return nil, nil
	},
}

var CDCommand = &Command{
	ShortHelp: "Change directory",
	LongHelp: `Change directory.
Usage: cd [relative or absolute path to directory]`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("must supply zero or one arguments")
		}

		if len(args) == 0 || args[0] == ".." {
			if state.CurrentDir.Parent == nil {
				return nil, fmt.Errorf("cannot go up a directory")
			}

			state.CurrentDir = state.CurrentDir.Parent
			return nil, nil
		}

		foundFile := fs.FindFileRelative(state.CurrentDir, state.RootDir, args[0])
		if foundFile == nil {
			return nil, fmt.Errorf("directory not found")
		}

		if foundFile.Type() != fs.FileTypeDirectory {
			return nil, fmt.Errorf("that is not a directory")
		}

		state.CurrentDir = foundFile.(*fs.Directory)

		return nil, nil
	},
}

var ExitCommand = &Command{
	ShortHelp: "Exit the current session",
	LongHelp: `Exit the current session.
Usage: exit`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		os.Exit(0)
		return nil, nil
	},
}

var RMDIRCommand = &Command{
	ShortHelp: "Remove a directory and its contents",
	LongHelp: `Remove a directory and its contents.
Usage: rmdir [directory name]`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("must supply directory name")
		}

		foundFile := fs.FindFileRelative(state.CurrentDir, state.RootDir, args[0])
		if foundFile == nil {
			return nil, fmt.Errorf("directory not found")
		}

		if foundFile.Type() != fs.FileTypeDirectory {
			return nil, fmt.Errorf("that is not a directory")
		}

		// Remove this file from its parent dir
		dir := foundFile.(*fs.Directory)
		if dir.Parent == nil {
			return nil, fmt.Errorf("cannot rmdir root")
		}

		delete(dir.Parent.Files, args[0])

		return nil, nil
	},
}

// func RMDIRCommand(state *State, args ...string) ([]byte, error) {
// 	if len(args) != 1 {
// 		return nil, fmt.Errorf("must supply one argument")
// 	}

// 	dirName := strings.ToLower(args[0])

// }
