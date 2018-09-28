package command

import (
	"bytes"
	"fmt"
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
		"help":    HelpCommand,
		"ls":      LSCommand,
		"mkdir":   MKDIRCommand,
		"cd":      CDCommand,
		"exit":    ExitCommand,
		"rmdir":   RMDIRCommand,
		"append":  AppendCommand,
		"cat":     CatCommand,
		"replace": ReplaceCommand,
		"scan":    ScanCommand,
		"connect": ConnectCommand,
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

		foundFile := fs.FindFileRelative(state.CurrentDir, state.CurrentHost.RootDir, args[0])
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
		if state.CurrentHost == state.LocalHost {
			go sendEvent(state.EventChan, EventTypeExit, "")
			return nil, nil
		}

		state.CurrentHost = state.LocalHost
		state.CurrentDir = state.LocalHost.RootDir

		return []byte("Disconnected."), nil
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

		foundFile := fs.FindFileRelative(state.CurrentDir, state.CurrentHost.RootDir, args[0])
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

var AppendCommand = &Command{
	ShortHelp: "Append text to an existing or new file",
	LongHelp: `Append text to a file. If it does not exist, it will be created.
Usage: create [filename] "your text here"`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("must supply at least directory name")
		}

		filePath := args[0]

		var textFile *fs.Text
		var createdNew bool
		foundFile := fs.FindFileRelative(state.CurrentDir, state.CurrentHost.RootDir, filePath)

		if foundFile == nil {
			// we must create the new file
			pathParts := strings.Split(filePath, "/")
			newFilename := filePath

			creatingInDir := state.CurrentDir
			if strings.HasSuffix(filePath, "/") || len(pathParts) > 1 {
				newFilename = pathParts[len(pathParts)-1]
				// This is a path to a file, we must first find the directory referenced
				foundDir := fs.FindFileRelative(state.CurrentDir, state.CurrentHost.RootDir, strings.TrimSuffix(filePath, newFilename))
				if foundDir == nil || foundDir.Type() != fs.FileTypeDirectory {
					return nil, fmt.Errorf("file path not valid - directory does not exist")
				}
				creatingInDir = foundDir.(*fs.Directory)
			}

			if err := fs.ValidateFileName(newFilename); err != nil {
				return nil, err
			}

			createdNew = true

			creatingInDir.Files[newFilename] = &fs.Text{FileName: newFilename}
			textFile = creatingInDir.Files[newFilename].(*fs.Text)
		} else if foundFile.Type() != fs.FileTypeText {
			return nil, fmt.Errorf("cannot append to non-text file")
		} else {
			textFile = foundFile.(*fs.Text)
		}

		if len(args) > 1 && args[1] != "" {
			textFile.Contents = append(textFile.Contents, []byte(args[1])...)
		}

		if createdNew {
			return []byte(fmt.Sprintf("created new file at %v", args[0])), nil
		}

		return []byte("appended to existing file"), nil
	},
}

var CatCommand = &Command{
	ShortHelp: "View contents of a file",
	LongHelp: `View contents of a file.
Usage: cat [path to file]`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("must supply a file path")
		}

		filePath := args[0]

		foundFile := fs.FindFileRelative(state.CurrentDir, state.CurrentHost.RootDir, filePath)
		if foundFile == nil {
			return nil, fmt.Errorf("file not found")
		}
		if foundFile.Type() != fs.FileTypeText {
			return nil, fmt.Errorf("%v is not a readable file", filePath)
		}

		return foundFile.(*fs.Text).Contents, nil
	},
}

var ReplaceCommand = &Command{
	ShortHelp: "Replace all instances of a string in a file",
	LongHelp: `Replace all instances of a string in a file.
Usage: replace [path to file] [search text] [replace text]`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("must supply path to file, search text and replace text")
		}

		filePath := args[0]

		foundFile := fs.FindFileRelative(state.CurrentDir, state.CurrentHost.RootDir, filePath)
		if foundFile == nil {
			return nil, fmt.Errorf("file not found")
		}
		if foundFile.Type() != fs.FileTypeText {
			return nil, fmt.Errorf("%v is not a writeable file", filePath)
		}

		file := foundFile.(*fs.Text)
		new := strings.Replace(string(file.Contents), args[1], args[2], -1)

		file.Contents = []byte(new)

		return nil, nil
	},
}

var ScanCommand = &Command{
	ShortHelp: "Scan other hosts connected to the current hosts",
	LongHelp: `Scan other hosts connected to the current hosts.
Usage: scan`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(state.CurrentHost.ConnectedHosts) == 0 {
			return []byte("No connected hosts."), nil
		}

		otherHosts := []string{}
		for host := range state.CurrentHost.ConnectedHosts {
			otherHosts = append(otherHosts, host)
		}

		return []byte(strings.Join(otherHosts, "\n")), nil
	},
}

var ConnectCommand = &Command{
	ShortHelp: "Connect to another host",
	LongHelp: `Connect to another host.
Usage: connect [hostname]`,
	Execute: func(state *State, args ...string) ([]byte, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("must supply a hostname")
		}

		host, ok := state.CurrentHost.ConnectedHosts[args[0]]
		if !ok {
			return nil, fmt.Errorf("host %s not found", args[0])
		}

		state.CurrentHost = host
		state.CurrentDir = host.RootDir

		return []byte(fmt.Sprintf("connected to %s", host.Hostname)), nil
	},
}
