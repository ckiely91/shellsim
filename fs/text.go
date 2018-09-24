package fs

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
