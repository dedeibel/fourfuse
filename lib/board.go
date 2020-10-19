package fourfuse

import (
	"os"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	fourc "github.com/moshee/go-4chan-api/api"
	"golang.org/x/net/context"
)

type Board struct {
	handle        string
	name          string
	slug          string
	inode         uint64
	threads       map[string]*Thread
	thumbnailsDir *ImageDir
	fetchMutex    sync.Mutex
}

func NewBoard(handle string, name string) *Board {
	return &Board{
		handle:        handle,
		name:          name,
		inode:         hashs(HASH_BOARD_PREFIX + handle),
		slug:          sanitizePathSegment(name),
		threads:       make(map[string]*Thread),
		thumbnailsDir: nil}
}

func (b *Board) Slug() string {
	return b.slug
}

func (b *Board) fetchThreads() {
	LogInfof("fetching: %s\n", b.handle)

	catalog, err := fourc.GetCatalog(b.handle)
	if err != nil {
		Log.Fatal(err)
	}

	thumbnails := NewImageList()
	for _, page := range catalog {
		for _, fourcThread := range page.Threads {
			thread := NewThread(fourcThread)
			b.threads[thread.Slug()] = thread
			thumbnails.Add(thread.GetThumbnail())
		}
	}

	thumbnails.SortByLocale()
	b.thumbnailsDir = NewImageDirFromImageList("thumbnails", thumbnails)
}

func (b *Board) ensureInitialized() {
	b.fetchMutex.Lock()
	defer b.fetchMutex.Unlock()
	if !b.hasBeenInitialized() {
		b.fetchThreads()
	}
}

func (b *Board) hasBeenInitialized() bool {
	return len(b.threads) > 0
}

func (b *Board) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = b.inode
	a.Mode = os.ModeDir | 0555
	return nil
}

func (b *Board) GetDirent() fuse.Dirent {
	return fuse.Dirent{
		Inode: b.inode,
		Name:  b.slug,
		Type:  fuse.DT_Dir}
}

func (b *Board) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	LogDebugf("read dir for %s\n", b.handle)
	b.ensureInitialized()

	threadsDirents := b.getThreadsDirents()

	var dirDirents = make([]fuse.Dirent, 0, 1+len(threadsDirents))
	dirDirents = append(dirDirents, threadsDirents...)
	dirDirents = append(dirDirents, b.thumbnailsDir.GetDirent())

	return dirDirents, nil
}

func (b *Board) getThreadsDirents() []fuse.Dirent {
	var entries = make([]fuse.Dirent, 0, len(b.threads))
	for _, thread := range b.threads {
		entries = append(entries, thread.GetDirent())
	}
	return entries
}

func (b *Board) Lookup(ctx context.Context, name string) (fs.Node, error) {
	b.ensureInitialized()

	if name == b.thumbnailsDir.Slug() {
		return b.thumbnailsDir, nil
	}

	if thread, present := b.threads[name]; present {
		return thread, nil
	} else {
		return nil, fuse.ENOENT
	}
}
