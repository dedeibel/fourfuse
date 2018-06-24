package fourfuse

import (
	"sort"

	"bazil.org/fuse"
)

type listEntry struct {
	file  *RemoteFile
	index int
}

func (le *listEntry) Slug() string {
	return le.file.Slug()
}

type ImageList struct {
	images       map[string]*listEntry
	order        []*listEntry
	postPrefetch *PostPrefetch
}

func NewImageList() *ImageList {
	imageList := &ImageList{
		images: make(map[string]*listEntry),
		order:  make([]*listEntry, 0)}

	imageList.postPrefetch = NewPostPrefetch(imageList)
	return imageList
}

func (il *ImageList) Copy() *ImageList {
	imageList := NewImageList()
	for _, listEntry := range il.order {
		imageList.Add(listEntry.file)
	}
	return imageList
}

func newListEntry(file *RemoteFile, index int) *listEntry {
	return &listEntry{file, index}
}

func (il *ImageList) Add(files ...*RemoteFile) {
	for _, file := range files {
		if file != nil {
			file.SetAccessCallback(il.postPrefetch)
			entry := newListEntry(file, il.nextIndex())
			il.images[file.Slug()] = entry
			il.order = append(il.order, entry)
		}
	}
}

func (il *ImageList) nextIndex() int {
	return len(il.order)
}

func (il *ImageList) Get(slug string) (file *RemoteFile, present bool) {
	if entry, present := il.images[slug]; present {
		return entry.file, present
	} else {
		return nil, false
	}
}

func (il *ImageList) GetIndex(file *RemoteFile) (index int, found bool) {
	entry, found := il.images[file.Slug()]
	if found {
		index = entry.index
	}

	return index, found
}

func (il *ImageList) GetContentDirents() []fuse.Dirent {
	var entries = make([]fuse.Dirent, 0, len(il.order))
	for _, entry := range il.order {
		entries = append(entries, entry.file.GetDirent())
	}
	return entries
}

func (il *ImageList) reindex() {
	for index, entry := range il.order {
		entry.index = index
	}
}

func (il *ImageList) SortByLocale() {
	sort.Sort(FilenameByLocale(il.order))
	il.reindex()
}

func (il *ImageList) GetNEntriesBefore(index int, amount int) []*RemoteFile {
	results := make([]*RemoteFile, 0, amount)
	for i := Max(index-1, 0); i >= Max(0, index-amount); i -= 1 {
		results = append(results, il.order[i].file)
	}

	return results
}

func (il *ImageList) GetNEntriesAfter(index int, amount int) []*RemoteFile {
	max := len(il.order) - 1
	results := make([]*RemoteFile, 0, amount)
	for i := Min(index+1, max); i <= Min(index+amount, max); i += 1 {
		results = append(results, il.order[i].file)
	}

	return results
}
