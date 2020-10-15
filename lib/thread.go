package fourfuse

import (
	"strconv"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	fourc "github.com/moshee/go-4chan-api/api"
	"golang.org/x/net/context"
)

type Thread struct {
	*Post
	board         string
	id            int64
	slug          string
	inode         uint64
	posts         map[string]*Post
  postsDir      *PostsDir
	thumbnails    *ImageList
	imageDir      *ImageDir
	thumbnailsDir *ImageDir
	fetchMutex    sync.Mutex
}

func NewThread(fourc *fourc.Thread) *Thread {
	return &Thread{
		board:         fourc.Board,
		id:            fourc.OP.Id,
		slug:          GeneratePostSlug(fourc.OP),
		inode:         hashs(HASH_THREAD_PREFIX + strconv.FormatInt(fourc.OP.Id, 10)),
		imageDir:      nil,
		thumbnailsDir: nil,
		Post:          NewPost(fourc.OP, nil),
		posts:         make(map[string]*Post),
    postsDir:      nil,
		thumbnails:    NewImageList()}
}

func (t *Thread) fetchPosts() {
	LogInfof("fetching: %s\n", t.slug)

	fourcThread, err := fourc.GetThread(t.board, t.id)
	if err != nil {
		Log.Fatal(err)
	}
	LogDebugf("api call done for: %s\n", t.slug)

	for _, fourcPost := range fourcThread.Posts {
		post := NewPost(fourcPost, t)
		t.posts[post.Slug()] = post
		t.thumbnails.Add(post.GetSamePrefixedSlugThumbnail())
	}
	t.postsDir = NewPostsDir("posts", t.posts)
	t.thumbnails.SortByLocale()

	t.imageDir = NewImageDir("images", t.allPostsImages())
	t.thumbnailsDir = NewImageDirFromImageList("thumbnails", t.thumbnails)
	LogDebugf("fetch done for: %s\n", t.slug)
}

func (t *Thread) allPostsImages() []*RemoteFile {
	files := make([]*RemoteFile, 0, len(t.posts))
	for _, post := range t.posts {
		if file := post.GetImage(); file != nil {
			files = append(files, file)
		}
	}

	return files
}

func (t *Thread) ensureInitialized() {
	LogTrace("thread init, lock")

	t.fetchMutex.Lock()
	defer t.fetchMutex.Unlock()

	LogTrace("thread init: ", t.hasBeenInitialized())
	if !t.hasBeenInitialized() {
		t.fetchPosts()
	}
	LogTrace("thread init done")
}

func (t *Thread) hasBeenInitialized() bool {
	return len(t.posts) > 0
}

func (p *Thread) GetThumbnail() *RemoteFile {
	return p.Post.GetSamePrefixedSlugThumbnail()
}

func (t *Thread) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	LogDebugf("read dir for %s\n", t.slug)
	t.ensureInitialized()

	var dirDirs, err = t.Post.ReadDirAll(ctx)
	if err != nil {
		return nil, err
	}

	dirDirs = append(dirDirs, t.postsDir.GetDirent())
	dirDirs = append(dirDirs, t.imageDir.GetDirent())
	dirDirs = append(dirDirs, t.thumbnailsDir.GetDirent())

	return dirDirs, nil
}

func (t *Thread) Lookup(ctx context.Context, name string) (fs.Node, error) {
	t.ensureInitialized()

	if name == t.postsDir.Slug() {
		return t.postsDir, nil
	}

	if name == t.imageDir.Slug() {
		return t.imageDir, nil
	}

	if name == t.thumbnailsDir.Slug() {
		return t.thumbnailsDir, nil
	}

	if post, postPresent := t.posts[name]; postPresent {
		return post, nil
	}

	if thumbnail, thumbnailPresent := t.thumbnails.Get(name); thumbnailPresent {
		return thumbnail, nil
	}

	return t.Post.Lookup(ctx, name)
}
