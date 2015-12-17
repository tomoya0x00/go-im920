package main

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"time"

	"github.com/tomoya0x00/go-im920"
)

// TODO: BusyCheck
func main() {
	serialName := "/dev/ttyAMA0" // for Raspberry Pi
	if runtime.GOOS == "windows" {
		serialName = "COM4"
	}

	c := &im920.Config{Name: serialName, ReadTimeout: 1000 * time.Millisecond}
	im, err := im920.Open(c)
	if err != nil {
		log.Fatalf("Failed to open: %s\n", err)
	}
	defer im.Close()

	id, err := im.GetId()
	if err != nil {
		fmt.Printf("Failed to GetId: %s\n", err)
	} else {
		fmt.Printf("ID: %v\n", id)
	}

	rids, err := im.GetAllRcvId()
	if err != nil {
		fmt.Printf("Failed to GetAllRcvId: %s\n", err)
	} else {
		fmt.Printf("RIDS: %v\n", rids)
	}

	rch := NewReceiver(im)
	b := make([]byte, 1)
	b[0] = 0

	for {
		select {
		case r := <-rch:
			fmt.Printf("Read %v\n", r)
			fmt.Printf("ReadInfo %v\n", im.LastReadInfo())
		default:
			fmt.Printf("Write 0x%02x\n", b[0])
			_, err = im.Write(b)
			if err != nil {
				fmt.Printf("Failed to write: %s\n", err)
			}

			time.Sleep(2000 * time.Millisecond)
			b[0]++
		}
	}
}

func NewReceiver(im *im920.IM920) <-chan []byte {
	ch := make(chan []byte, 100)
	go receiver(im, ch)

	return ch
}

func receiver(im *im920.IM920, ch chan<- []byte) {
	buf := make([]byte, 256)
	for {
		time.Sleep(100 * time.Millisecond)
		n, err := im.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Printf("Read error:%s\n", err)
			continue
		}
		if n > 0 {
			ch <- buf[:n]
		}
	}
}
