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

var WriteTests = []struct {
	in_writeData   []byte
	in_dummyData   []byte
	out_errorIsNil bool
}{
	{[]byte("01234567890"), []byte("OK\r\n"), true},
	{[]byte("01234567890"), []byte("NG\r\n"), false},
	{[]byte("01234567890"), []byte("HOGE\r\n"), false},
	{[]byte("01234567890"), []byte(""), false},
	{[]byte(""), []byte("OK\r\n"), true},
	{[]byte(""), []byte("NG\r\n"), false},
	{[]byte(""), []byte("HOGE\r\n"), false},
	{[]byte(""), []byte(""), false},
}

func TestWrite(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{serial}

	for i, tt := range WriteTests {
		serial.dummyData = tt.in_dummyData
		data := tt.in_writeData
		_, err := im.Write(data)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]Write(%v) => %v, want errorIsNil = %v",
				i, tt.in_writeData, err, tt.out_errorIsNil)
		}
	}
}
