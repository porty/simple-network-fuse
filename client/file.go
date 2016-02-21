package client

import (
	"bazil.org/fuse"
	snf "github.com/porty/simple-network-fuse"
	"golang.org/x/net/context"
)

type File struct {
	info *snf.File
}

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 0
	a.Mode = f.info.Mode
	a.Size = uint64(f.info.Size)
	a.Mtime = f.info.ModTime
	return nil
}
