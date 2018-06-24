package fourfuse

const numberOfPrefetchEntries = 8

type PostPrefetch struct {
	imageList     *ImageList
	previousIndex int
}

func NewPostPrefetch(imageList *ImageList) *PostPrefetch {
	return &PostPrefetch{
		imageList:     imageList,
		previousIndex: -1}
}

func (pp *PostPrefetch) onAccess(remoteFile *RemoteFile) {
	index, found := pp.imageList.GetIndex(remoteFile)

	switch {
	case !found:
		return
	case index == pp.previousIndex:
		return
	case (index + 1) == pp.previousIndex:
		LogDebugf("detected movement up (current index %d), prefetching\n", index)
		pp.prefetch(pp.imageList.GetNEntriesBefore(index, numberOfPrefetchEntries))
	case (index - 1) == pp.previousIndex:
		LogDebugf("detected movement down (current index %d), prefetching\n", index)
		pp.prefetch(pp.imageList.GetNEntriesAfter(index, numberOfPrefetchEntries))
	}

	pp.previousIndex = index
}

func (pp *PostPrefetch) prefetch(files []*RemoteFile) {
	for _, file := range files {
		if !file.IsDownloaded() {
			GetPrefetchWorkerPool().ScheduleDownload(file)
		}
	}
}
