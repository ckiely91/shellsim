package command

import "github.com/ckiely91/shellsim/fs"

type Host struct {
	Hostname       string
	RootDir        *fs.Directory
	ConnectedHosts map[string]*Host
}

func NewHost(hostname string) *Host {
	return &Host{
		Hostname: hostname,
		RootDir: &fs.Directory{
			Parent: nil,
			Files:  map[string]fs.File{},
		},
		ConnectedHosts: map[string]*Host{},
	}
}
