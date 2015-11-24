package im920

import (
	"testing"
	"time"
)

const (
	serialName = "COM4"
)

func TestOpen(t *testing.T) {
	c := &Config{Name: serialName, ReadTimeout: time.Second * 5}
	im, err := Open(c)
	if err != nil {
		t.Errorf("Failed to open: %s", err)
		return
	}
	defer im.Close()
}

func TestWrite(t *testing.T) {
	c := &Config{Name: serialName, ReadTimeout: time.Second * 5}
	im, err := Open(c)
	if err != nil {
		t.Errorf("Failed to open: %s", err)
		return
	}
	defer im.Close()

	data := []byte("0123456789012345678901234567890123456789012345678901234567890123456789")
	_, err = im.Write(data)
	if err != nil {
		t.Errorf("Failed to write: %s", err)
	}
}
