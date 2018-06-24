# fourfuse

Readonly 4chan fuse filesystem in golang (on linux)

## Installation


```
go get github.com/dedeibel/fourfuse
```

Fuse must be installed.

```
apt install fuse

lsmod | grep fuse
fuse                  118784  1
```

Tested on debian stretch with go 1.10

## Usage

```
mkdir /tmp/mnt
fourfuse /tmp/mnt
```

```
### List boards
ls /tmp/mnt
3DCG
'Do It Yourself'
Music
'Shit 4chan Says'
…
```

```
### List threads posts
ls /tmp/mnt/Photography | head
1971605 Go pro Please post images that are JPG format smaller than
1971605 Go pro Please post images that are JPG format smaller than_t.jpg
3260531 Segue Thread
3260531 Segue Thread_t.jpg
3261816 So the goddamn sun came out for once and I tried to use th
3261816 So the goddamn sun came out for once and I tried to use th_t.jpg
…
```

```
### Find Photography gear threads
ls -d /tmp/mnt/Photography/*Gear*
'/tmp/mnt/Photography/3310621 gear Gear thread'        '/tmp/mnt/Photography/3313958 gear Gear thread'
'/tmp/mnt/Photography/3310621 gear Gear thread_t.jpg'  '/tmp/mnt/Photography/3313958 gear Gear thread_t.jpg'
# "*_t.jpg" is the thumbnails for each thread
```

```
### Show comment of one of the posts in the gear thread
cat /tmp/mnt/Photography/3310621\ gear\ Gear\ thread/3310621\ gear\ Gear\ thread/comment.txt
Last thread: >>3307144 

Read the sticky first!

Post anything gear related, cameras, lenses, filters, bags, tripods, other accessories (clothing, fancy straps, Leica) etc...
Post your question here, instead of starting a new thread about which lens to buy or what are the best beginner cameras.
…
```

```
### List all images in the gear thread
ls /tmp/mnt/Photography/3310621\ gear\ Gear\ thread/images
3310621 01.jpg
3310634 asd.jpg
3310636 xr-s.jpg
…
```

## Usage hints and limitations

* Images inside a of a directory are prefetched when moving sequentially,
  **ordered by name**. This can result in a seemless browsing experience.
* 4chan API's restriction of 1 request per second is honored, therefore
  listing directories and the amount of entries can be slow
* Accessing boards in a **file explorer** that lists directories might take long
  since they try to display the number of contained files. It might help to switch
  to an **icon view mode** before browsing.
* Threads and Posts are not updated you will have to kill and restart the app
  for that
* A lot of memory might be used since a lot of the data is cached

### Download images of a whole board

For images of the board Photography.

```
# Setup the fuse mount
fourfuse /tmp/mnt

# Download all files in the images/ directories - will take a while
rsync --recursive --human-readable --stats --progress --time-limit=30 --prune-empty-dirs --include="*/" --include="images/*.*" --exclude="*" /tmp/mnt/Photography /tmp/photos

# Remove empty directories
find . -type d -empty -delete
```

## Development hints

Enable more logging by enabling it in ``lib/logger.go``

Automatically build your code to get quick feedback. Combines well with
automatically saving editors, there is a plugin for vim ([vim-auto-save](https://github.com/907th/vim-auto-save)).

```
reflex -r '\.go$' -- bash -c 'echo -e "---\n\n\n" ; go build && echo "ok"'
```

* reflex file change watcher [github.com/cespare/reflex](https://github.com/cespare/reflex
)
```
# I also use goimports
goimports -w -d .
```

## Future ideas

* Update threads
    * After timeout – the api recommends min 1 minute
* Dynamically increase the prefetch amount if hit multiple times
* Provide thread text view in a single file
* Embed post comments in image exif data to get all infos via an image viewer
* Configure prefetch amount and workers etc. via cli parameters

## Dependencies

* [bazil/fuse](https://github.com/bazil/fuse) – thanks!
* [dedeibel/go-4chan-api](https://github.com/dedeibel/go-4chan-api)
    * a fork with fixes for [moshee/go-4chan-api](https://github.com/moshee/go-4chan-api) – thanks!

## References

* [4chan API doc](https://github.com/4chan/4chan-API)


