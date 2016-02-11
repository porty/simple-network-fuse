package client

import (
	"encoding/gob"
	"log"
	"net"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	snf "github.com/porty/simple-network-fuse"
	"golang.org/x/net/context"
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

type FS struct {
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func (fs FS) Root() (fs.Node, error) {
	name := ""
	req := snf.Request{
		Op:   snf.DIR_LIST,
		Name: &name,
	}
	err := fs.encoder.Encode(&req)
	if err != nil {
		log.Println("Failed to request file operation: " + err.Error())
		return nil, err
	}

	resp := snf.Response{}
	err = fs.decoder.Decode(&resp)
	if err != nil {
		log.Println("Failed to get response: " + err.Error())
		return nil, err
	}

	log.Printf("Response: %d", resp.ErrorCode)

	fileMap := make(map[string]snf.File)
	for _, f := range resp.Files {
		fileMap[f.Name] = f
	}
	return Dir{files: &fileMap}, nil
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

type Dir struct {
	files *map[string]snf.File
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
		return File{info: f}, nil
	}
	return nil, fuse.EEXIST
}

type File struct {
	info snf.File
}

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 0
	a.Mode = f.info.Mode
	a.Size = uint64(f.info.Size)
	a.Mtime = f.info.ModTime
	return nil
}
