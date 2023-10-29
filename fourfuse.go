package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"

	fourfuse "github.com/dedeibel/fourfuse/lib"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
}

func main() {
	fourfuse.InitializeLogger()
	fourfuse.UseSystemLocale()

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)

	c, err := fuse.Mount(
		mountpoint,
		fuse.AsyncRead(),
		fuse.FSName("4get"),
		fuse.Subtype("4get"))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	registerCleanupHandler(mountpoint)

	err = fs.Serve(c, fourfuse.NewFs(fourfuse.LoadBoards()))
	if err != nil {
		log.Fatal(err)
	}
}

func cleanupFusemount(path string) {
	if err := fuse.Unmount(path); err != nil {
		log.Fatal(err)
	}
}

func registerCleanupHandler(path string) {
	sigtermChannel := make(chan os.Signal, 1)
	signal.Notify(sigtermChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigtermChannel
		cleanupFusemount(path)
		os.Exit(0)
	}()
}
