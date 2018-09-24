package fs

import (
	"strings"
)

type FileType uint8

const (
	FileTypeDirectory FileType = iota
	FileTypeText
)

type File interface {
	Type() FileType
	Name() string
}

func findFileRelativeToDir(dir *Directory, paths []string) File {
	curPath := strings.ToLower(paths[0])
	paths = paths[1:]
	if len(paths) == 0 {
		// We're looking for the actual file here
		if curPath == ".." {
			if dir.Parent == nil {
				// Can't go further up. Return nothing
				return nil
			}

			return dir.Parent
		}

		if file, ok := dir.Files[curPath]; ok {
			return file
		}

		return nil
	}

	// Else we're looking for a directory to drill into
	if curPath == ".." {
		if dir.Parent == nil {
			// Can't go further up. Return nothing
			return nil
		}

		return findFileRelativeToDir(dir.Parent, paths)
	}

	file, ok := dir.Files[curPath]
	if !ok {
		return nil
	}

	dir, ok = file.(*Directory)
	if !ok {
		return nil
	}

	return findFileRelativeToDir(dir, paths)
}

// Returns nil if path invalid or file not found
func FindFileRelative(currentDir *Directory, rootDir *Directory, path string) File {
	var startingDir *Directory
	if path == "/" {
		return rootDir
	}

	if strings.HasPrefix(path, "/") {
		startingDir = rootDir
	} else {
		startingDir = currentDir
	}

	path = strings.Trim(path, "/")
	paths := strings.Split(path, "/")

	return findFileRelativeToDir(startingDir, paths)
}
