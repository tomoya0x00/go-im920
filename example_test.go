package im920_test

import (
	"fmt"
	"github.com/tomoya0x00/go-im920"
	"time"
)

func ExampleRead() {
	// This test will not be run, it has no "Output:" comment.
	c := &im920.Config{Name: "COM4", ReadTimeout: 1 * time.Second}
	im, err := im920.Open(c)
	if err != nil {
		fmt.Printf("Failed to open: %s\n", err)
		return
	}
	defer im.Close()

	for i := 0; i < 20; i++ {
		buf := make([]byte, 64)
		n, err := im.Read(buf)
		if err != nil {
			fmt.Printf("Failed to read: %s\n", err)
			continue
		}

		fmt.Printf("Read Data: %v\n", buf[:n])
	}
}
