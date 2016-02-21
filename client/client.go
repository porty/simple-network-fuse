package client

import (
	"encoding/gob"
	"log"
	"net"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func Mount(host, mountpoint string) error {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Println("Connected")

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("simple-network-fuse"),
		fuse.Subtype("rofl"),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	err = fs.Serve(c, FS{
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
	})
	if err != nil {
		return err
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}
	return nil
}
