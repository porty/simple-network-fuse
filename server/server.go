package server

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"path/filepath"
	"syscall"

	snf "github.com/porty/simple-network-fuse"
)

var OKStruct = snf.Response{
	ErrorCode: 0,
	Files:     []*snf.File{},
}

var BlankString = ""

func ServeTCP(bindAddress string, fs FileSystem) error {
	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		return err
	}
	return ServeListener(listener, fs)
}

func ServeListener(listener net.Listener, fs FileSystem) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			log.Println("Client connected")
			if err := ServeStream(conn, fs); err != nil {
				log.Println("Error while serving connection: " + err.Error())
			} else {
				log.Println("Client disconnected")
			}
		}(conn)
	}
}

func ServeStream(stream io.ReadWriter, fs FileSystem) error {
	r := new(snf.Request)
	for {
		if err := readRequest(stream, r); err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		if err := handleRequest(stream, r, fs); err != nil {
			return err
		}
	}
}

func readRequest(reader io.Reader, request *snf.Request) error {
	dec := gob.NewDecoder(reader)
	return dec.Decode(request)
}

func handleRequest(w io.Writer, r *snf.Request, fs FileSystem) error {
		if err := fs.CreateFile(r.Name); err != nil {
			return writeError(w, err)
	if r.Op == snf.FileCreate {
		}
		return writeOK(w)
		if err := fs.Delete(r.Name); err != nil {
			return writeError(w, err)
	} else if r.Op == snf.FileUnlink {
		}
		return writeOK(w)
		files, err := fs.List(r.Name)
	} else if r.Op == snf.DirList {
		if err != nil {
			return writeError(w, err)
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
