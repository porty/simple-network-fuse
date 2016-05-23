package server

import (
	"bytes"
	"encoding/gob"
	"errors"
	"testing"

	snf "github.com/porty/simple-network-fuse"
)

type FakeFileSystem struct {
	CreateFileFunc func(name string) error
	DeleteFunc     func(name string) error
	ListFunc       func(name string) ([]*snf.File, error)
}

func (ffs *FakeFileSystem) CreateFile(name string) error {
	if ffs.CreateFileFunc != nil {
		return ffs.CreateFileFunc(name)
	}
	return nil
}

func (ffs *FakeFileSystem) Delete(name string) error {
	if ffs.DeleteFunc != nil {
		return ffs.DeleteFunc(name)
	}
	return nil
}

func (ffs *FakeFileSystem) List(name string) ([]*snf.File, error) {
	if ffs.ListFunc != nil {
		return ffs.ListFunc(name)
	}
	return nil, nil
}

func TestImmediateDisconnect(t *testing.T) {
	fs := FakeFileSystem{}
	var input bytes.Buffer
	var output bytes.Buffer
	s := New(&input, &output, &fs)
	err := s.Serve()
	if err != nil {
		t.Error("Expected no error, got: " + err.Error())
	}
	if output.Len() > 0 {
		t.Errorf("Expected no network traffic, got %d bytes", output.Len())
	}
}

func TestCreateFile(t *testing.T) {
	fileToCreate := "sup"
	fs := FakeFileSystem{}
	var input bytes.Buffer
	var output bytes.Buffer
	encoder := gob.NewEncoder(&input)
	decoder := gob.NewDecoder(&output)

	// test successful create file
	encoder.Encode(snf.Request{
		Op:   snf.FileCreate,
		Name: fileToCreate,
	})
	fs.CreateFileFunc = func(name string) error {
		if name != fileToCreate {
			t.Error("Expected '%s', received '%s'", fileToCreate, name)
		}
		return nil
	}
	s := New(&input, &output, &fs)
	err := s.Serve()
	if err != nil {
		t.Error("Expected no error, received: " + err.Error())
		return
	}
	var response snf.Response
	if err = decoder.Decode(&response); err != nil {
		t.Error("Expected no error, received: " + err.Error())
		return
	}
	if response.ErrorCode != 0 {
		t.Errorf("Expected error code 0, received %d", response.ErrorCode)
	}
	if len(response.Files) > 0 {
		t.Errorf("Expected to receive no files, received %d of them", len(response.Files))
	}

	// test failed create file
	input.Reset()
	output.Reset()
	encoder = gob.NewEncoder(&input)
	decoder = gob.NewDecoder(&output)
	encoder.Encode(snf.Request{
		Op:   snf.FileCreate,
		Name: fileToCreate,
	})
	fs.CreateFileFunc = func(name string) error {
		if name != fileToCreate {
			t.Error("Expected '%s', received '%s'", fileToCreate, name)
		}
		return errors.New("some error")
	}
	s = New(&input, &output, &fs)
	err = s.Serve()
	if err != nil {
		t.Error("Expected no error, received: " + err.Error())
		return
	}
	if err = decoder.Decode(&response); err != nil {
		t.Error("Expected no error, received: " + err.Error())
		return
	}
	if response.ErrorCode == 0 {
		t.Errorf("Expected error code other than 0, received %d", response.ErrorCode)
	}
	if len(response.Files) > 0 {
		t.Errorf("Expected to receive no files, received %d of them", len(response.Files))
	}
}
