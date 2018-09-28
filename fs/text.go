package fs

import (
	"fmt"
	"regexp"
)

type Text struct {
	FileName string
	Contents []byte
}

func (t *Text) Type() FileType {
	return FileTypeText
}

func (t *Text) Name() string {
	return t.FileName
}

var fileNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`)

func ValidateFileName(name string) error {
	if fileNameRegex.MatchString(name) {
		return nil
	}

	return fmt.Errorf("file name \"%v\" contains invalid characters", name)
}
