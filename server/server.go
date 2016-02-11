package server

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"syscall"

	snf "github.com/porty/simple-network-fuse"
)

var OKStruct = snf.Response{
	ErrorCode: 0,
	Files:     []snf.File{},
}

var BlankString = ""

func ServeTCP(bindAddress string, path string) error {
	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		return err
	}
	return ServeListener(listener, path)
}

func ServeListener(listener net.Listener, path string) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			log.Println("Client connected")
			if err := ServeConnection(conn, path); err != nil {
				if err == io.EOF {
					log.Println("Client disconnected")
				} else {
					log.Println("Error while serving connection: " + err.Error())
				}
			}
		}(conn)
	}
}

func ServeConnection(conn net.Conn, path string) error {
	r := new(snf.Request)
	for {
		if err := readRequest(conn, r); err != nil {
			return err
		}
		if err := handleRequest(conn, r, path); err != nil {
			return err
		}
	}
}

func readRequest(reader io.Reader, request *snf.Request) error {
	dec := gob.NewDecoder(reader)
	return dec.Decode(request)
}

func handleRequest(w io.Writer, r *snf.Request, path string) error {
	if r.Op == snf.FILE_CREATE {
		if r.Name == nil {
			return errors.New("Protocol error: Name field required for FILE_CREATE operation")
		}
		fullPath := pathMunge(path, *r.Name)
		f, err := os.Create(fullPath)
		if err != nil {
			return writeError(w, err)
		}
		f.Close()
		return writeOK(w)
	} else if r.Op == snf.FILE_UNLINK {
		if r.Name == nil {
			return errors.New("Protocol error: Name field required for FILE_UNLINK operation")
		}
		fullPath := pathMunge(path, *r.Name)
		if err := os.Remove(fullPath); err != nil {
			return writeError(w, err)
		}
		return writeOK(w)
	} else if r.Op == snf.DIR_LIST {
		if r.Name == nil {
			//return errors.New("Protocol error: Name field required for DIR_LIST operation")
			r.Name = &BlankString
		}
		fullPath := pathMunge(path, *r.Name)
		f, err := os.Open(fullPath)
		if err != nil {
			return writeError(w, err)
		}
		defer f.Close()

		fis, err := f.Readdir(0)
		if err != nil && err != io.EOF {
			return writeError(w, err)
		}
		files := make([]snf.File, 0, len(fis))
		for _, fi := range fis {
			files = append(files, snf.FromFileInfo(fi))
		}
		resp := snf.Response{
			ErrorCode: 0,
			Files:     files,
		}
		return writeResponse(w, &resp)
	} else {
		return fmt.Errorf("Protocol error: unknown operation %d", r.Op)
	}
}

func pathMunge(basePath, name string) string {
	return filepath.Join(filepath.Clean(basePath), filepath.Clean(name))
}

func writeError(w io.Writer, err error) error {
	enc := gob.NewEncoder(w)
	response := snf.Response{
		ErrorCode: uint32(syscall.EIO),
	}
	return enc.Encode(&response)
}

func writeOK(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(&OKStruct)
}

func writeResponse(w io.Writer, resp *snf.Response) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(resp)
}
