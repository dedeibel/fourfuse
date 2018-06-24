package fourfuse

import (
	"bazil.org/fuse"
	"golang.org/x/net/context"
)

type File struct {
	inode   uint64
	content string
}

func NewFile(inode uint64, content string) *File {
	return &File{
		inode,
		content}
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.inode
	a.Mode = 0444
	a.Size = uint64(len(f.content))
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(f.content), nil
}
