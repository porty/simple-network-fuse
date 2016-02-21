package client

import (
	"encoding/gob"
	"log"

	"bazil.org/fuse/fs"
	snf "github.com/porty/simple-network-fuse"
)

type FS struct {
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func (fs FS) Root() (fs.Node, error) {
	req := snf.Request{
		Op:   snf.DIR_LIST,
		Name: "",
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

	fileMap := make(map[string]*snf.File)
	for _, f := range resp.Files {
		fileMap[f.Name] = f
	}
	return Dir{files: &fileMap}, nil
}
