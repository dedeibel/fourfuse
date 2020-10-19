package fourfuse

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	fourc "github.com/moshee/go-4chan-api/api"
	"golang.org/x/net/context"
)

const discussionSlug string = "discussion.txt"

type Thread struct {
	*Post
	board         string
	id            int64
	slug          string
	inode         uint64
	posts         map[string]*Post
	postsDir      *PostsDir
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
		postsDir:      nil}
}

func (t *Thread) fetchPosts() {
	LogInfof("fetching: %s\n", t.slug)

	fourcThread, err := fourc.GetThread(t.board, t.id)
	if err != nil {
		Log.Fatal(err)
	}
	LogDebugf("api call done for: %s\n", t.slug)

	thumbnails := NewImageList()
	for _, fourcPost := range fourcThread.Posts {
		post := NewPost(fourcPost, t)
		t.posts[post.Slug()] = post
		thumbnails.Add(post.GetSamePrefixedSlugThumbnail())
	}

	t.postsDir = NewPostsDir(t.slug+" posts", t.posts)

	t.imageDir = NewImageDir(t.slug+" images", t.allPostsImages())

	thumbnails.SortByLocale()
	t.thumbnailsDir = NewImageDirFromImageList(t.slug+" thumbnails", thumbnails)

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

func (t *Thread) getDiscussionSanitized() string {
	var discussion = ""

	posts := make([]*Post, 0, len(t.posts))
	for _, v := range t.posts {
		posts = append(posts, v)
	}
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Time().Before(posts[j].Time())
	})

	for _, post := range t.posts {
		user := post.GetUserName()
		if len(user) > 0 {
			discussion += "U: " + user
			discussion += "\n"
		}
		subject := post.GetSubjectSanitized()
		if len(subject) > 0 {
			discussion += "S: " + subject
			discussion += "\n"
		}
		discussion += "T: " + post.Time().Format(time.RFC3339)
		discussion += "\n"
		discussion += post.GetCommentSanitized()
		discussion += "\n------\n"
		discussion += "\n"
	}
	return discussion
}

func (t *Thread) GetThumbnail() *RemoteFile {
	return t.Post.GetSamePrefixedSlugThumbnail()
}

func (t *Thread) getDiscussiontDirent() fuse.Dirent {
	return fuse.Dirent{
		Inode: hashs(HASH_THREAD_INFO_PREFIX + t.slug + "discussion"),
		Name:  discussionSlug,
		Type:  fuse.DT_File}
}

func (t *Thread) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	LogDebugf("read dir for %s\n", t.slug)
	t.ensureInitialized()

	var dirDirs, err = t.Post.ReadDirAll(ctx)
	if err != nil {
		return nil, err
	}

	dirDirs = append(dirDirs, t.getDiscussiontDirent())
	dirDirs = append(dirDirs, t.postsDir.GetDirent())
	dirDirs = append(dirDirs, t.imageDir.GetDirent())
	dirDirs = append(dirDirs, t.thumbnailsDir.GetDirent())

	return dirDirs, nil
}

func (t *Thread) Lookup(ctx context.Context, name string) (fs.Node, error) {
	t.ensureInitialized()

	if name == discussionSlug {
		return NewFile(
			hashs(HASH_THREAD_DISCUSSION+t.slug),
			t.getDiscussionSanitized()), nil
	}

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

	return t.Post.Lookup(ctx, name)
}
