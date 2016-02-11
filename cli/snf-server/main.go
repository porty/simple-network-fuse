package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/porty/simple-network-fuse/server"
)

func main() {
	if exitCode := intmain(); exitCode != 0 {
		os.Exit(exitCode)
	}
}

func intmain() int {
	bind := flag.String("bind", "", "Address to bind to, e.g. localhost:5956")
	path := flag.String("path", "", "Path to expose, e.g. /mnt/files")

	flag.Parse()

	fmt.Printf("The bind is %s and the path is %s\n", *bind, *path)

	if *bind == "" || *path == "" {
		flag.PrintDefaults()
		return 1
	}

	if err := server.ServeTCP(*bind, *path); err != nil {
		fmt.Println("Failed to start server: " + err.Error())
		return 1
	}
	return 0
}
