package server

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	snf "github.com/porty/simple-network-fuse"
)

type FileSystem interface {
	CreateFile(name string) error
	Delete(name string) error
	List(name string) ([]*snf.File, error)
}

type RealFileSystem struct {
	basedir string
}

var _ FileSystem = &RealFileSystem{}

func NewRealFileSystem(dir string) (RealFileSystem, error) {
	dir = filepath.Clean(dir)
	fi, err := os.Stat(dir)
	if err != nil {
		return RealFileSystem{}, err
	}
	if !fi.IsDir() {
		return RealFileSystem{}, errors.New("Path specified was not a directory")
	}

	return RealFileSystem{
		basedir: dir,
	}, nil
}

func (fs *RealFileSystem) CreateFile(name string) error {
	f, err := os.Create(fs.JoinPath(name))
	if err != nil {
		return err
	}
	return f.Close()
}

func (fs *RealFileSystem) Delete(name string) error {
	return os.Remove(fs.JoinPath(name))
}

func (fs *RealFileSystem) List(name string) ([]*snf.File, error) {
	var path string
	if name == "" {
		path = fs.basedir
	} else {
		path = fs.JoinPath(name)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	files := make([]*snf.File, 0, len(fis))
	for _, fi := range fis {
		files = append(files, snf.FromFileInfo(fi))
	}
	return files, nil
}

func (fs *RealFileSystem) JoinPath(name string) string {
	return filepath.Join(fs.basedir, filepath.Clean(name))
}
