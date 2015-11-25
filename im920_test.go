package im920

import (
	"testing"
)

type fakeSerial struct {
	dummyData []byte
}

func newFakeSerial() *fakeSerial {
	return &fakeSerial{}
}

func (serial *fakeSerial) Read(p []byte) (n int, err error) {
	if len(serial.dummyData) == 0 {
		return 0, nil
	}

	n = copy(p, serial.dummyData)
	err = nil

	if n < len(serial.dummyData) {
		serial.dummyData = serial.dummyData[n:len(serial.dummyData)]
	} else {
		serial.dummyData = make([]byte, 0)
	}

	return
}

func (serial *fakeSerial) Close() error {
	return nil
}

func (serial *fakeSerial) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestWrite(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{serial}

	serial.dummyData = []byte("OK\r\n")

	data := []byte("0123456789012345678901234567890123456789012345678901234567890123456789")
	_, err := im.Write(data)
	if err != nil {
		t.Errorf("Failed to write: %s", err)
	}
}
