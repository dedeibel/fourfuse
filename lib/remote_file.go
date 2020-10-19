package fourfuse

import (
	"regexp"
	"sync"

	"io/ioutil"
	"net/http"

	"bazil.org/fuse"
	"golang.org/x/net/context"
)

// Download remote files early on "Attr" call. This will help with
// file browsers that are otherwise confused with 0 size thumbnails.
// The posts main images are not affacted because their size is available
// as meta data in advance.
// Disable it if you do not need thumbnails.
const EARLY_REMOTE_DOWNLOAD = true

type RemoteFile struct {
	slug               string
	url                string
	inode              uint64
	content            []byte
	size               uint64
	fileAccessCallback FileAccessCallback
	downloadMutex      sync.RWMutex
}

func NewRemoteFile(inode uint64, slug string, url string, size uint64) *RemoteFile {
	return &RemoteFile{
		url:                url,
		slug:               slug,
		inode:              inode,
		size:               size,
		fileAccessCallback: NewFileAccessCallbackNoop()}
}

func (f *RemoteFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.inode
	a.Mode = 0444

	// assume 0 means not downloaded, post images have their size already set,
	// they will be downloaded on ReadAll
	if EARLY_REMOTE_DOWNLOAD && f.size == 0 {
		f.Download(true)
		f.fileAccessCallback.onAccess(f)
	}

	a.Size = uint64(f.size)
	return nil
}

func (f *RemoteFile) ReadAll(ctx context.Context) ([]byte, error) {
	f.Download(true)
	f.fileAccessCallback.onAccess(f)

	return []byte(f.content), nil
}

func (f *RemoteFile) Download(hasPriority bool) {
	LogTrace("remote file download called, read lock")

	f.downloadMutex.RLock()
	if f.content == nil {
		f.downloadMutex.RUnlock()
		f.doDownload(hasPriority)
	} else {
		f.downloadMutex.RUnlock()
	}

	LogTrace("remote file download call done")
}

func (f *RemoteFile) doDownload(hasPriority bool) {
	LogTrace("remote file not present, write lock")
	f.downloadMutex.Lock()
	defer f.downloadMutex.Unlock()

	cachedContent, cacheHit := GetCache().Lookup(f.url)

	if cacheHit {
		LogDebugf("cache hit for url: %s\n", f.url)

		f.content = cachedContent
		f.size = uint64(len(f.content))
	} else {
		if hasPriority {
			LogInfof("-> downloading '%s' url: %s with 'priority'", f.slug, f.url)
			GetPrefetchWorkerPool().Pause()
			defer GetPrefetchWorkerPool().Resume()
		} else {
			LogInfof("-> downloading '%s' url: %s", f.slug, f.url)
		}

		resp, err := http.Get(f.url)
		if err != nil {
			Log.Fatal(err)
		}
		defer resp.Body.Close()
		f.content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			Log.Panic(err)
		}
		f.size = uint64(len(f.content))

		LogDebugf("<- downloaded '%s' url: %s", f.slug, f.url)

		GetCache().Store(f.url, f.content)
	}
}

func (f *RemoteFile) IsDownloaded() bool {
	if f.content != nil {
		return true
	}

	_, cacheHit := GetCache().Lookup(f.url)
	return cacheHit
}

func (f *RemoteFile) Slug() string {
	return f.slug
}

func (f *RemoteFile) AddSuffix(suffix string) {
	var re = regexp.MustCompile(`^(.*)\.(\.*)$`)
	f.slug = re.ReplaceAllString(f.slug, `$1`+suffix+`.$2`)
}

func (f *RemoteFile) AddPrefix(prefix string) {
	f.slug = prefix + f.slug
}

func (f *RemoteFile) GetDirent() fuse.Dirent {
	return fuse.Dirent{Inode: f.inode, Name: f.slug, Type: fuse.DT_File}
}

func (f *RemoteFile) SetAccessCallback(fac FileAccessCallback) {
	f.fileAccessCallback = fac
}
