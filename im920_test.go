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

var ReceiveTests = []struct {
	in_out         []byte
	out_errorIsNil bool
}{
	{
		[]byte("00,06E5,B5:0A,1F,76,00,00,00,00,00\r\n"), true,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76\r\n"), true,
	},
	{
		[]byte("00,06E5,B5:0A\r\n"), true,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76"), true,
	},
	{
		[]byte("00,06E5,B5:\r\n"), true,
	},
	{
		[]byte("00,06E5,B5\r\n"), true,
	},
	{
		[]byte(""), false,
	},
}

func TestReceive(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	buf := make([]byte, maxReadSize)

	for i, tt := range ReceiveTests {
		serial.dummyData = tt.in_out
		n, err := im.receive(buf)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]receive() => %v, want errorIsNil = %v",
				i, err, tt.out_errorIsNil)
		}
		if !bytes.Equal(buf[:n], tt.in_out) {
			t.Errorf("[%d]receive() => %v, want data = %v",
				i, buf, tt.in_out)
		}
	}
}

var IssueCommandTests = []struct {
	in_cmd         string
	in_param       string
	in_dummyData   []byte
	out            []byte
	out_errorIsNil bool
}{
	{
		"HOGE", "HUGA", []byte("OK\r\n"),
		[]byte("OK\r\n"), true,
	},
	{
		"HOGE", "HUGA", []byte("NG\r\n"),
		[]byte("NG\r\n"), false,
	},
	{
		"HOGE", "HUGA", []byte("FUGA\r\n"),
		[]byte("FUGA\r\n"), true,
	},
	{
		"HOGE", "HUGA", []byte(""),
		[]byte(""), false,
	},
}

func TestIssueCommand(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range IssueCommandTests {
		serial.dummyData = tt.in_dummyData
		resp, err := im.IssueCommand(tt.in_cmd, tt.in_param)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]IssueCommand(%v, %v) => %v, want errorIsNil = %v",
				i, tt.in_cmd, tt.in_param, err, tt.out_errorIsNil)
		}
		if !bytes.Equal(resp, tt.out) {
			t.Errorf("[%d]IssueCommand() => %v, want data = %v",
				i, resp, tt.out)
		}
	}
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
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

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
	out            []byte
	out_errorIsNil bool
}{
	{
		[]byte("00,06E5,B5:0A,1F,76,00,00,00,00,00\r\n"),
		[]byte{0x0A, 0x1F, 0x76, 0x00, 0x00, 0x00, 0x00, 0x00}, true,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76\r\n"),
		[]byte{0x0A, 0x1F, 0x76}, true,
	},
	{
		[]byte("00,06E5,B5:0A\r\n"),
		[]byte{0x0A}, true,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76"),
		[]byte{}, false,
	},
	{
		[]byte("00,06E5,B5:\r\n"),
		[]byte{}, false,
	},
	{
		[]byte("00,06E5,B5\r\n"),
		[]byte{}, false,
	},
	{
		[]byte(""),
		[]byte{}, false,
	},
}

func TestRead(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	buf := make([]byte, maxReadSize)

	for i, tt := range ReadTests {
		serial.dummyData = tt.in
		n, err := im.Read(buf)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]Read() => %v, want errorIsNil = %v",
				i, err, tt.out_errorIsNil)
		}
		if !bytes.Equal(buf[:n], tt.out) {
			t.Errorf("[%d]Read() => %v, want data = %v",
				i, buf, tt.out)
		}
	}
}
