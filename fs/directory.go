package fs

import (
	"fmt"
	"regexp"
	"strings"
)

type Directory struct {
	Parent  *Directory
	DirName string
	Files   map[string]File
}

func (d *Directory) Type() FileType {
	return FileTypeDirectory
}

func (d *Directory) Name() string {
	return d.DirName
}

func (d *Directory) FullPath() string {
	dirNames := []string{}
	currentDir := d
	for currentDir != nil {
		if currentDir.DirName != "" {
			dirNames = append([]string{currentDir.DirName}, dirNames...)
		}
		currentDir = currentDir.Parent
	}

	return "/" + strings.Join(dirNames, "/")
}

var dirNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-\_]+$`)

func ValidateDirName(name string) error {
	if dirNameRegex.MatchString(name) {
		return nil
	}

	return fmt.Errorf("directory name contains invalid characters")
}
