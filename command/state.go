package command

import (
	"github.com/ckiely91/shellsim/fs"
)

type State struct {
	RootDir        *fs.Directory
	CurrentDir     *fs.Directory
	Commands       map[string]*Command
	CommandHistory *CommandHistory
	EventChan      chan *Event
}

func NewState() *State {
	rootDir := &fs.Directory{
		Parent: nil,
		Files:  map[string]fs.File{},
	}

	return &State{
		RootDir:        rootDir,
		CurrentDir:     rootDir,
		Commands:       standardCommands(),
		CommandHistory: &CommandHistory{},
		EventChan:      make(chan *Event),
	}
}
