package im920

import (
	"bytes"
	"testing"
	"time"
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
	im := &IM920{s: serial, readTimeout: 1 * time.Second}

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

var ReadTests = []struct {
	in             []byte
	out_data       []byte
	out_n          int
	out_errorIsNil bool
}{
	{
		[]byte("00,06E5,B5:0A,1F,76,00,00,00,00,00\r\n"),
		[]byte{0x0A, 0x1F, 0x76, 0x00, 0x00, 0x00, 0x00, 0x00}, 8, true,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76\r\n"),
		[]byte{0x0A, 0x1F, 0x76}, 3, true,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76"),
		[]byte{}, 0, false,
	},
	{
		[]byte("00,06E5,B5:\r\n"),
		[]byte{}, 0, false,
	},
	{
		[]byte("00,06E5,B5\r\n"),
		[]byte{}, 0, false,
	},
	{
		[]byte(""),
		[]byte{}, 0, false,
	},
}

func TestRead(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 1 * time.Second}

	buf := make([]byte, maxReadSize)

	for i, tt := range ReadTests {
		serial.dummyData = tt.in
		n, err := im.Read(buf)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]Read() => %v, want errorIsNil = %v",
				i, err, tt.out_errorIsNil)
		}
		if !bytes.Equal(buf[:n], tt.out_data) {
			t.Errorf("[%d]Read() => %v, want data = %v",
				i, buf, tt.out_data)
		}

		if n != tt.out_n {
			t.Errorf("[%d]Read() => %v, want n = %v",
				i, n, tt.out_n)
		}
	}
}
