package server

import (
	"os"
	"time"
)

type Response struct {
	ErrorCode uint32
	Files     []File
}

const (
	TYPE_FILE uint8 = iota
	TYPE_DIR
	TYPE_SOCKET
	TYPE_PIPE
)

type File struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
}

func FromFileInfo(fi os.FileInfo) File {
	return File{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
	}
}
