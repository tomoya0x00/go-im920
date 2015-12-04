package im920_test

import (
	"fmt"
	"github.com/tomoya0x00/go-im920"
	"time"
)

func Example() {
	// This test will not be run, it has no "Output:" comment.
	c := &im920.Config{Name: "COM4", ReadTimeout: 1 * time.Second}
	im, err := im920.Open(c)
	if err != nil {
		fmt.Printf("Failed to open: %s\n", err)
		return
	}
	defer im.Close()

	id, err := im.GetId()
	if err != nil {
		fmt.Printf("Failed to GetId: %s\n", err)
		return
	}
	fmt.Printf("ID: %v\n", id)

	/*
		    err = im.AddRcvId(0x0001)
			if err != nil {
				fmt.Printf("Failed to AddRcvId: %s\n", err)
				return
			}
	*/

	version, err := im.IssueCommandRespStr("RDVR", "")
	if err != nil {
		fmt.Printf("Failed to RDVR: %s\n", err)
		return
	}
	fmt.Printf("VERSION: %v\n", version)

	/*
		    err = im.DeleteAllRcvId()
			if err != nil {
				fmt.Printf("Failed to DeleteAllRcvId: %s\n", err)
				return
			}
	*/

	rids, err := im.GetAllRcvId()
	if err != nil {
		fmt.Printf("Failed to GetAllRcvId: %s\n", err)
		return
	}
	fmt.Printf("RIDS: %v\n", rids)

	data := []byte("0123456789")
	_, err = im.Write(data)
	if err != nil {
		fmt.Printf("Failed to write: %s\n", err)
		return
	}
}
