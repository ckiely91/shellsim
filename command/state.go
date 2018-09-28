package command

import (
	"fmt"

	"github.com/ckiely91/shellsim/fs"
)

type State struct {
	CurrentDir     *fs.Directory
	LocalHost      *Host
	CurrentHost    *Host
	Commands       map[string]*Command
	CommandHistory *CommandHistory
	EventChan      chan *Event
}

func NewState() *State {
	localHost := NewHost("192.168.1.1")

	otherHost1 := NewHost("200.12.1.29")
	otherHost2 := NewHost("129.21.230.12")

	localHost.ConnectedHosts[otherHost1.Hostname] = otherHost1
	localHost.ConnectedHosts[otherHost2.Hostname] = otherHost2

	return &State{
		CurrentDir:     localHost.RootDir,
		LocalHost:      localHost,
		CurrentHost:    localHost,
		Commands:       standardCommands(),
		CommandHistory: &CommandHistory{},
		EventChan:      make(chan *Event),
	}
}

func (s *State) Logf(line string, args ...interface{}) {
	go sendEvent(s.EventChan, EventTypeLog, fmt.Sprintf(line, args...))
}
