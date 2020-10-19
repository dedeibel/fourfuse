package fourfuse

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
)

/*
 * bazil fuse is caching results too, but we make images with
 * the same urls available at different locations and want to
 * prevent multiple downloads
 */

const maxCacheEntries = 500

type FileCache struct {
	cache *lru.TwoQueueCache
}

var (
	fileCache     *FileCache
	fileCacheOnce sync.Once
)

func GetCache() *FileCache {
	fileCacheOnce.Do(func() {
		cache, err := lru.New2Q(maxCacheEntries)
		if err != nil {
			Log.Panic(err)
		}

		fileCache = &FileCache{cache}
	})

	return fileCache
}

func (fc *FileCache) Lookup(key string) ([]byte, bool) {
	data, found := fc.cache.Get(key)
	if found {
		return data.([]byte), found
	} else {
		return nil, found
	}
}

func (fc *FileCache) Store(key string, data []byte) {
	fc.cache.Add(key, data)
}
