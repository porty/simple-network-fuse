package client

import (
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	snf "github.com/porty/simple-network-fuse"
	"golang.org/x/net/context"
)

type Dir struct {
	files *map[string]*snf.File
}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Dir::ReadDirAll()")
	files := make([]fuse.Dirent, 0, len(*d.files))

	for _, f := range *d.files {
		files = append(files, fuse.Dirent{
			Inode: 0,
			Name:  f.Name,
			Type:  fileModeToDirentType(f.Mode),
		})
	}
	return files, nil
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if f, ok := (*d.files)[name]; ok {
		return File{info: f}, nil
	}
	return nil, fuse.ENOENT
}

func fileModeToDirentType(fm os.FileMode) fuse.DirentType {
	if (fm & os.ModeNamedPipe) != 0 {
		return fuse.DT_FIFO
	}
	if (fm & os.ModeCharDevice) != 0 {
		return fuse.DT_Char
	}
	if (fm & os.ModeDir) != 0 {
		return fuse.DT_Dir
	}
	if (fm & os.ModeDevice) != 0 {
		// ??
		return fuse.DT_Block
	}
	if (fm & os.ModeSymlink) != 0 {
		return fuse.DT_Link
	}
	if (fm & os.ModeSocket) != 0 {
		return fuse.DT_Socket
	}
	return fuse.DT_File
}
