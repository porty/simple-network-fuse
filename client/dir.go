package client

import (
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	snf "github.com/porty/simple-network-fuse"
	"golang.org/x/net/context"
)

type Dir struct {
	conn  connection
	files *map[string]*snf.File
	path  string
}

func DirFromPath(conn connection, path string) (*Dir, error) {
	req := snf.Request{
		Op:   snf.DirList,
		Name: path,
	}
	err := conn.enc.Encode(&req)
	if err != nil {
		log.Println("Failed to request file operation: " + err.Error())
		return nil, err
	}

	resp := snf.Response{}
	err = conn.dec.Decode(&resp)
	if err != nil {
		log.Println("Failed to get response: " + err.Error())
		return nil, err
	}

	fileMap := make(map[string]*snf.File)
	for _, f := range resp.Files {
		fileMap[f.Name] = f
	}
	return &Dir{
		conn:  conn,
		files: &fileMap,
		path:  path,
	}, nil
}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
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
		if f.Mode.IsDir() {
			return DirFromPath(d.conn, pathMunge(d.path, name))
		}
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

func pathMunge(basePath, name string) string {
	return filepath.Join(filepath.Clean(basePath), filepath.Clean(name))
}
