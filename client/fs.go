package client

import "bazil.org/fuse/fs"

type FS struct {
	conn connection
}

func (fs *FS) Root() (fs.Node, error) {
	return DirFromPath(fs.conn, "/")
}
