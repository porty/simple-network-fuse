package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/porty/simple-network-fuse/client"
)

func main() {
	if exitCode := intmain(); exitCode != 0 {
		os.Exit(exitCode)
	}
}

func intmain() int {
	host := flag.String("host", "", "The host address, e.g. host.lan:4925")
	mountpoint := flag.String("mountpoint", "", "Where to mount to, e.g. ~/over-here")

	flag.Parse()

	if *host == "" || *mountpoint == "" {
		flag.PrintDefaults()
		return 1
	}

	if err := client.Mount(*host, *mountpoint); err != nil {
		fmt.Println("Failed to start client: " + err.Error())
		return 1
	}

	return 0
}
