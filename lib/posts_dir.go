package fourfuse

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type PostsDir struct {
	slug  string
	inode uint64
	posts map[string]*Post
}

func NewPostsDir(slug string, posts map[string]*Post) *PostsDir {
	postsDir := &PostsDir{
		slug:  slug,
		inode: hashs(HASH_POSTS_DIR_PREFIX + slug),
		posts: posts}
	return postsDir
}

func (d *PostsDir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = d.inode
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d *PostsDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var dirDirs = make([]fuse.Dirent, 0, len(d.posts))
	for _, post := range d.posts {
		dirDirs = append(dirDirs, post.GetDirent())
	}
	return dirDirs, nil
}

func (d *PostsDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for _, post := range d.posts {
		if post.Slug() == name {
			return post, nil
		}
	}

	return nil, fuse.ENOENT
}

func (d *PostsDir) Slug() string {
	return d.slug
}

func (d *PostsDir) GetDirent() fuse.Dirent {
	return fuse.Dirent{Inode: d.inode, Name: d.slug, Type: fuse.DT_Dir}
}
