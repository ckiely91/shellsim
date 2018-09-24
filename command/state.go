package command

import (
	"github.com/ckiely91/shellsim/fs"
)

type State struct {
	RootDir    *fs.Directory
	CurrentDir *fs.Directory
	Commands   map[string]*Command
	Exiting    bool
}

func NewState() *State {
	rootDir := &fs.Directory{
		Parent: nil,
		Files:  map[string]fs.File{},
	}

	return &State{
		RootDir:    rootDir,
		CurrentDir: rootDir,
		Commands:   standardCommands(),
		Exiting:    false,
	}
}
