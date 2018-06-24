package fourfuse

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type ImageDir struct {
	slug      string
	inode     uint64
	imageList *ImageList
}

func NewImageDir(slug string, remoteFiles []*RemoteFile) *ImageDir {
	imageList := NewImageList()
	imageList.Add(remoteFiles...)
	return newImageDir(slug, imageList)
}

/* ImageList will be copied */
func NewImageDirFromImageList(slug string, imageList *ImageList) *ImageDir {
	return newImageDir(slug, imageList.Copy())
}

func newImageDir(slug string, imageList *ImageList) *ImageDir {
	imageList.SortByLocale()
	imageDir := &ImageDir{
		slug:      slug,
		inode:     hashs(HASH_IMAGE_DIR_PREFIX + slug),
		imageList: imageList}
	return imageDir
}

func (d *ImageDir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = d.inode
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d *ImageDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return d.GetContentDirents(), nil
}

func (d *ImageDir) GetContentDirents() []fuse.Dirent {
	return d.imageList.GetContentDirents()
}

func (d *ImageDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	file, filePresent := d.imageList.Get(name)
	if filePresent {
		return file, nil
	}

	return nil, fuse.ENOENT
}

func (d *ImageDir) Slug() string {
	return d.slug
}

func (d *ImageDir) GetDirent() fuse.Dirent {
	return fuse.Dirent{Inode: d.inode, Name: d.slug, Type: fuse.DT_Dir}
}
