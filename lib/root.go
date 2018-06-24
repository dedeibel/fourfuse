package fourfuse

import (
	"bazil.org/fuse/fs"
)

type FS struct {
	boardsDir *BoardsDir
}

func NewFs(boardsDir *BoardsDir) *FS {
	return &FS{boardsDir}
}

func (fs FS) Root() (fs.Node, error) {
	return fs.boardsDir, nil
}
