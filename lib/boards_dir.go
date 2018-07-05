package fourfuse

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	fourc "github.com/moshee/go-4chan-api/api"
	"golang.org/x/net/context"
)

type BoardsDir struct {
	boards map[string]*Board
}

const boardsDirInode uint64 = 1

func LoadBoards() *BoardsDir {
	return &BoardsDir{loadBoards()}
}

func (b *BoardsDir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = boardsDirInode
	a.Mode = os.ModeDir | 0555
	return nil
}

func (b *BoardsDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var dirDirs = make([]fuse.Dirent, 0, len(b.boards))
	for _, board := range b.boards {
		dirDirs = append(dirDirs, board.GetDirent())
	}
	return dirDirs, nil
}

func (b *BoardsDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if board, present := b.boards[name]; present {
		return board, nil
	} else {
		return nil, fuse.ENOENT
	}
}

func loadBoards() map[string]*Board {
	LogDebug("Loading Boards list")
	boards := make(map[string]*Board)
	fourcBoards, err := fourc.GetBoards()
	if err != nil {
		Log.Fatal(err)
	}

	for _, fourcBoard := range fourcBoards {
		board := NewBoard(fourcBoard.Board, fourcBoard.Title)
		boards[board.Slug()] = board
	}
	return boards
}
