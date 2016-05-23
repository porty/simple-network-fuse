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

type Server struct {
	encoder *gob.Encoder
	decoder *gob.Decoder
	fs      FileSystem
}

func New(r io.Reader, w io.Writer, fs FileSystem) *Server {
	return &Server{
		encoder: gob.NewEncoder(w),
		decoder: gob.NewDecoder(r),
		fs:      fs,
	}
}

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
			s := New(conn, conn, fs)
			if err := s.Serve(); err != nil {
				log.Println("Error while serving connection: " + err.Error())
			} else {
				log.Println("Client disconnected")
			}
		}(conn)
	}
}

func (s *Server) Serve() error {
	r := new(snf.Request)
	for {
		if err := s.readRequest(r); err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		if err := s.handleRequest(r); err != nil {
			return err
		}
	}
}

func (s *Server) readRequest(request *snf.Request) error {
	return s.decoder.Decode(request)
}

func (s *Server) handleRequest(r *snf.Request) error {
	if r.Op == snf.FileCreate {
		if err := s.fs.CreateFile(r.Name); err != nil {
			return s.writeError(err)
		}
		return s.writeOK()
	} else if r.Op == snf.FileUnlink {
		if err := s.fs.Delete(r.Name); err != nil {
			return s.writeError(err)
		}
		return s.writeOK()
	} else if r.Op == snf.DirList {
		files, err := s.fs.List(r.Name)
		if err != nil {
			return s.writeError(err)
		}
		resp := snf.Response{
			ErrorCode: 0,
			Files:     files,
		}
		return s.writeResponse(&resp)
	}
	return fmt.Errorf("Protocol error: unknown operation %d", r.Op)
}

func (s *Server) writeError(err error) error {
	response := snf.Response{
		ErrorCode: uint32(syscall.EIO),
	}
	return s.encoder.Encode(&response)
}

func (s *Server) writeOK() error {
	return s.encoder.Encode(&OKStruct)
}

func (s *Server) writeResponse(resp *snf.Response) error {
	return s.encoder.Encode(resp)
}

func pathMunge(basePath, name string) string {
	return filepath.Join(filepath.Clean(basePath), filepath.Clean(name))
}
