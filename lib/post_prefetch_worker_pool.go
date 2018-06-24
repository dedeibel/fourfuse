package fourfuse

import (
	"sync"
)

const prefetchQueueSize = 40
const numberOfWorkers = 6

var (
	prefetchWorkerPool     *PrefetchWorkerPool
	prefetchWorkerPoolOnce sync.Once
)

func GetPrefetchWorkerPool() *PrefetchWorkerPool {
	prefetchWorkerPoolOnce.Do(func() {
		prefetchWorkerPool = NewPrefetchWorkerPool()
	})

	return prefetchWorkerPool
}

type PrefetchWorkerPool struct {
	jobs           chan *RemoteFile
	active         sync.Map
	pause          bool
	pauseCondition *sync.Cond
}

func NewPrefetchWorkerPool() *PrefetchWorkerPool {
	workers := &PrefetchWorkerPool{
		jobs:           make(chan *RemoteFile, prefetchQueueSize),
		pause:          false,
		active:         sync.Map{},
		pauseCondition: sync.NewCond(&sync.Mutex{})}

	workers.startWorkers()

	return workers
}

func (wp *PrefetchWorkerPool) startWorkers() {
	for w := 0; w < numberOfWorkers; w++ {
		go worker(&wp.active, &wp.pause, wp.pauseCondition, w, wp.jobs)
	}
}

func worker(active *sync.Map, pause *bool, pauseCondition *sync.Cond, workerId int, jobs <-chan *RemoteFile) {
	for file := range jobs {
		LogDebugf("=> worker %d received job: '%s'", workerId, file.Slug())

		slept := false
		pauseCondition.L.Lock()
		for *pause {
			LogDebugf("-- worker %d sleeping", workerId)
			slept = true
			pauseCondition.Wait()
		}

		if slept {
			LogDebugf("-- worker %d resuming", workerId)
		}
		pauseCondition.L.Unlock()

		LogDebugf("=> worker %d starting job: '%s'", workerId, file.Slug())

		if !file.IsDownloaded() {
			file.Download(false)
		}

		LogDebugf("<= worker %d finished job: '%s'", workerId, file.Slug())
		active.Delete(file)
	}
}

func (wp *PrefetchWorkerPool) ScheduleDownload(file *RemoteFile) {
	LogDebugf("-> scheduling '%s'", file.Slug())
	if _, present := wp.active.LoadOrStore(file, true); !present {

		select {
		case wp.jobs <- file:
			// added to queue
		default:
			LogInfof("  skipping scheduling because of full queue: '%s'", file.Slug())
			// queue full, do not prefetch this one
		}
	}
	LogDebugf("<- scheduled '%s'", file.Slug())
}

func (wp *PrefetchWorkerPool) Pause() {
	LogDebug("... pause worker pool")

	wp.pauseCondition.L.Lock()
	wp.pause = true
	wp.pauseCondition.L.Unlock()
}

func (wp *PrefetchWorkerPool) Resume() {
	LogDebug(">>> resume worker pool")
	wp.pauseCondition.L.Lock()
	wp.pause = false
	wp.pauseCondition.L.Unlock()
	wp.pauseCondition.Broadcast()
}
