package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/porty/simple-network-fuse/server"
)

func TestSimpleMount(t *testing.T) {
	log.Print("Starting test yay")
	mountdir, err := ioutil.TempDir(os.TempDir(), "snf")
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer func() {
		if err := os.Remove(mountdir); err != nil {
			log.Printf("Failed to remove temp dir \"%s\": %s", mountdir, err.Error())
		}
	}()
	log.Printf("Created mount dir at %s", mountdir)
	fs, err := server.NewRealFileSystem(mountdir)
	if err != nil {
		t.Error(err.Error())
		return
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err.Error())
		return
	}
	log.Printf("Listening at %s", ln.Addr())

	serverResultChan := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			serverResultChan <- err
			return
		}
		ln.Close()
		log.Print("Accepted a client :)")
		serverResultChan <- server.ServeStream(conn, &fs)
	}()

	log.Print("About to mount...")
	mountResultChan := make(chan error, 1)
	go func() {
		mountResultChan <- Mount(ln.Addr().String(), mountdir)
	}()

	mounted := false
	for started := time.Now(); (time.Now().Sub(started) <= (3 * time.Second)) && !mounted; {
		mounted, err = isMounted(mountdir)
		if err != nil {
			t.Errorf("Failed attempting to figure out if dir was mounted or not: %s", err.Error())
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !mounted {
		t.Error("Given up waiting for directory to mount")
		return
	}

	log.Print("Mounted. Unmounting")

	if err != nil {
		t.Errorf("Expecting no error from Mount(), received: %s", err.Error())
		return
	}

	cmd := exec.Command("umount", mountdir)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err = cmd.Run(); err != nil {
		t.Errorf("Expected no error when unmounting, received: %s", err.Error())
		return
	}
	if err = <-serverResultChan; err != nil {
		t.Errorf("Received error from server.ServeConnection(): %s", err.Error())
		return
	}
	log.Print("Done")
}

func isMounted(mountdir string) (bool, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("mount | grep %s | wc -l", mountdir))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Printf("BTW, whilst attempting to find if the dir was mounted the output was: %s", out.String())
		return false, err
	}

	mounts, err := strconv.Atoi(strings.TrimSpace(out.String()))
	if err != nil {
		return false, err
	}

	return (mounts == 1), nil
}
