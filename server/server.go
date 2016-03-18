package server

import (
	"encoding/gob"
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
	Files:     []*snf.File{},
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
				log.Println("Error while serving connection: " + err.Error())
			} else {
				log.Println("Client disconnected")
			}
		}(conn)
	}
}

func ServeConnection(conn net.Conn, path string) error {
	r := new(snf.Request)
	for {
		if err := readRequest(conn, r); err != nil {
			if err != io.EOF {
				return err
			}
			return nil
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
		fullPath := pathMunge(path, r.Name)
		f, err := os.Create(fullPath)
		if err != nil {
			return writeError(w, err)
		}
		f.Close()
		return writeOK(w)
	} else if r.Op == snf.FILE_UNLINK {
		fullPath := pathMunge(path, r.Name)
		if err := os.Remove(fullPath); err != nil {
			return writeError(w, err)
		}
		return writeOK(w)
	} else if r.Op == snf.DIR_LIST {
		fullPath := pathMunge(path, r.Name)
		f, err := os.Open(fullPath)
		if err != nil {
			return writeError(w, err)
		}
		defer f.Close()

		fis, err := f.Readdir(0)
		if err != nil && err != io.EOF {
			return writeError(w, err)
		}
		files := make([]*snf.File, 0, len(fis))
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
