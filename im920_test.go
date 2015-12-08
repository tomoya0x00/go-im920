package im920

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

type fakeSerial struct {
	dummyData  []byte
	writedData []byte
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
	serial.writedData = p
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
	{
		"HOGE", "HUGA", []byte("\r\n"),
		[]byte("\r\n"), true,
	},
	{
		"HOGE", "HUGA", []byte("00,06E5,B5:0A\r\nOK\r\n"),
		[]byte("OK\r\n"), true,
	},
	{
		"HOGE", "HUGA", []byte("00,06E5,B5:0A\r\n00,06E5,B5:0B\r\nOK\r\n"),
		[]byte("OK\r\n"), true,
	},
	{
		"HOGE", "HUGA", []byte("00,06E5,B5:0A\r\n00,06E5,B5:0B\r\nOK"),
		[]byte("OK"), true,
	},
	{
		"HOGE", "HUGA", []byte("00,06E5,B5:0B\r\nNG\r\n"),
		[]byte("NG\r\n"), false,
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

var IssueCommandNormalTests = []struct {
	in_cmd         string
	in_param       string
	in_dummyData   []byte
	out_errorIsNil bool
}{
	{
		"HOGE", "HUGA", []byte("OK\r\n"),
		true,
	},
	{
		"HOGE", "HUGA", []byte("NG\r\n"),
		false,
	},
	{
		"HOGE", "HUGA", []byte("FUGA\r\n"),
		false,
	},
	{
		"HOGE", "HUGA", []byte(""),
		false,
	},
	{
		"HOGE", "HUGA", []byte("\r\n"),
		false,
	},
}

func TestIssueNormalCommand(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range IssueCommandNormalTests {
		serial.dummyData = tt.in_dummyData
		err := im.IssueCommandNormal(tt.in_cmd, tt.in_param)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]IssueCommandNormal(%v, %v) => %v, want errorIsNil = %v",
				i, tt.in_cmd, tt.in_param, err, tt.out_errorIsNil)
		}
	}
}

var IssueCommandRespStrTests = []struct {
	in_cmd         string
	in_param       string
	in_dummyData   []byte
	out            string
	out_errorIsNil bool
}{
	{
		"HOGE", "HUGA", []byte("OK\r\n"),
		"OK", true,
	},
	{
		"HOGE", "HUGA", []byte("NG\r\n"),
		"NG", false,
	},
	{
		"HOGE", "HUGA", []byte("FUGA\r\n"),
		"FUGA", true,
	},
	{
		"HOGE", "HUGA", []byte("FUGA\r\nFUGA\r\n"),
		"FUGA\r\nFUGA", true,
	},
	{
		"HOGE", "HUGA", []byte(""),
		"", false,
	},
	{
		"HOGE", "HUGA", []byte("\r\n"),
		"", true,
	},
}

func TestIssueCommandRespStr(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range IssueCommandRespStrTests {
		serial.dummyData = tt.in_dummyData
		resp, err := im.IssueCommandRespStr(tt.in_cmd, tt.in_param)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]IssueCommandRespStr(%v, %v) => %v, want errorIsNil = %v",
				i, tt.in_cmd, tt.in_param, err, tt.out_errorIsNil)
		}
		if resp != tt.out {
			t.Errorf("[%d]IssueCommandRespStr() => %v, want data = %v",
				i, resp, tt.out)
		}
	}
}

var IssueCommandRespNumTests = []struct {
	in_cmd         string
	in_param       string
	in_dummyData   []byte
	out            uint16
	out_errorIsNil bool
}{
	{
		"HOGE", "HUGA", []byte("0\r\n"),
		0x0000, true,
	},
	{
		"HOGE", "HUGA", []byte("1\r\n"),
		0x0001, true,
	},
	{
		"HOGE", "HUGA", []byte("10\r\n"),
		0x0010, true,
	},
	{
		"HOGE", "HUGA", []byte("101\r\n"),
		0x0101, true,
	},
	{
		"HOGE", "HUGA", []byte("1010\r\n"),
		0x1010, true,
	},
	{
		"HOGE", "HUGA", []byte("01010\r\n"),
		0, false,
	},
	{
		"HOGE", "HUGA", []byte("1"),
		0, false,
	},
	{
		"HOGE", "HUGA", []byte(""),
		0, false,
	},
	{
		"HOGE", "HUGA", []byte("\r\n"),
		0, false,
	},
}

func TestIssueCommandRespNum(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range IssueCommandRespNumTests {
		serial.dummyData = tt.in_dummyData
		resp, err := im.IssueCommandRespNum(tt.in_cmd, tt.in_param)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]IssueCommandRespNum(%v, %v) => %v, want errorIsNil = %v",
				i, tt.in_cmd, tt.in_param, err, tt.out_errorIsNil)
		}
		if resp != tt.out {
			t.Errorf("[%d]IssueCommandRespNum() => %v, want data = %v",
				i, resp, tt.out)
		}
	}
}

var IssueCommandRespNumsTests = []struct {
	in_cmd         string
	in_param       string
	in_dummyData   []byte
	out            []uint16
	out_errorIsNil bool
}{
	{
		"HOGE", "HUGA", []byte("0\r\n"),
		[]uint16{0x0000}, true,
	},
	{
		"HOGE", "HUGA", []byte("1\r\n"),
		[]uint16{0x0001}, true,
	},
	{
		"HOGE", "HUGA", []byte("10\r\n"),
		[]uint16{0x0010}, true,
	},
	{
		"HOGE", "HUGA", []byte("101\r\n"),
		[]uint16{0x0101}, true,
	},
	{
		"HOGE", "HUGA", []byte("1010\r\n"),
		[]uint16{0x1010}, true,
	},
	{
		"HOGE", "HUGA", []byte("1010\r\n0101\r\n"),
		[]uint16{0x1010, 0x0101}, true,
	},
	{
		"HOGE", "HUGA", []byte("01010\r\n"),
		nil, false,
	},
	{
		"HOGE", "HUGA", []byte("1"),
		nil, false,
	},
	{
		"HOGE", "HUGA", []byte(""),
		nil, false,
	},
	{
		"HOGE", "HUGA", []byte("\r\n"),
		nil, true,
	},
}

func TestIssueCommandRespNums(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range IssueCommandRespNumsTests {
		serial.dummyData = tt.in_dummyData
		resp, err := im.IssueCommandRespNums(tt.in_cmd, tt.in_param)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]IssueCommandRespNums(%v, %v) => %v, want errorIsNil = %v",
				i, tt.in_cmd, tt.in_param, err, tt.out_errorIsNil)
		}
		if !reflect.DeepEqual(resp, tt.out) {
			t.Errorf("[%d]IssueCommandRespNums() => %v, want data = %v",
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
	out_readInfo   ReadInfo
	out_errorIsNil bool
}{
	{
		[]byte("00,06E5,B5:0A,1F,76"),
		[]byte{},
		ReadInfo{},
		false,
	},
	{
		[]byte("00,06E5,B5:\r\n"),
		[]byte{},
		ReadInfo{},
		false,
	},
	{
		[]byte("00,06E5,B5\r\n"),
		[]byte{},
		ReadInfo{},
		false,
	},
	{
		[]byte(""),
		[]byte{},
		ReadInfo{},
		false,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76,00,00,00,00,00\r\n"),
		[]byte{0x0A, 0x1F, 0x76, 0x00, 0x00, 0x00, 0x00, 0x00},
		ReadInfo{FromNode: 0x00, FromId: 0x06E5, FromRssi: 0xb5},
		true,
	},
	{
		[]byte("00,06E5,B5:0A,1F,76\r\n"),
		[]byte{0x0A, 0x1F, 0x76},
		ReadInfo{FromNode: 0x00, FromId: 0x06E5, FromRssi: 0xb5},
		true,
	},
	{
		[]byte("00,06E5,B5:0A\r\n"),
		[]byte{0x0A},
		ReadInfo{FromNode: 0x00, FromId: 0x06E5, FromRssi: 0xb5},
		true,
	},
	{
		[]byte("01,06E6,B6:0A\r\n"),
		[]byte{0x0A},
		ReadInfo{FromNode: 0x01, FromId: 0x06E6, FromRssi: 0xb6},
		true,
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
		info := im.LastReadInfo()
		if info != tt.out_readInfo {
			t.Errorf("[%d]LastReadInfo() => %v, want data = %v",
				i, info, tt.out_readInfo)
		}
	}
}

var GetIdTests = []struct {
	in_dummyData   []byte
	out            Id
	out_errorIsNil bool
}{
	{[]byte("0000\r\n"), 0x0000, true},
	{[]byte("0001\r\n"), 0x0001, true},
	{[]byte("ffff\r\n"), 0xffff, true},
	{[]byte("0000"), 0, false},
	{[]byte(""), 0, false},
	{[]byte("\r\n"), 0, false},
}

func TestGetId(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range GetIdTests {
		serial.dummyData = tt.in_dummyData
		id, err := im.GetId()
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]GetId() => %v, want errorIsNil = %v",
				i, err, tt.out_errorIsNil)
		}
		if id != tt.out {
			t.Errorf("[%d]GetId() => %v, want data = %v",
				i, id, tt.out)
		}
	}
}

/*
var AddRcvIdTests = []struct {
    in             Id
	in_dummyData   []byte
	out_writedData []byte
	out_errorIsNil bool
}{
	{
        0x0001, []byte("OK\r\nOK\r\nOK\r\n"),
        []byte("SRID 0001 \r\n"), true,
    },
	{
        0xffff, []byte("OK\r\nOK\r\nOK\r\n"),
        []byte("SRID ffff \r\n"), true,
    },
    {
        0x0001, []byte("NG\r\n"),
        nil, false,
    },
    {
        0x0001, []byte("OK\r\nNG\r\n"),
        []byte("SRID 0001 \r\n"), false,
    },
    {
        0x0001, []byte("OK\r\nOK\r\nNG\r\n"),
        []byte("SRID 0001 \r\n"), false,
    },
}

func TestAddRcvId(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range AddRcvIdTests {
		serial.dummyData = tt.in_dummyData
		err := im.AddRcvId(tt.in)
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]AddRcvId() => %v, want errorIsNil = %v",
				i, err, tt.out_errorIsNil)
		}
		if !bytes.Equal(serial.writedData, tt.out_writedData) {
			t.Errorf("[%d]AddRcvId() => %v, want data = %v",
				i, serial.writedData, tt.out_writedData)
		}
	}
}
*/

var GetAllRcvIdTests = []struct {
	in_dummyData   []byte
	out            []Id
	out_errorIsNil bool
}{
	{
		[]byte("0000\r\n"),
		[]Id{0x0000}, true,
	},
	{
		[]byte("0000\r\n0001\r\n"),
		[]Id{0x0000, 0x0001}, true,
	},
	{
		[]byte("0000\r\n0001\r\n0002\r\n"),
		[]Id{0x0000, 0x0001, 0x0002}, true,
	},
	{
		[]byte("0000"),
		nil, false,
	},
	{
		[]byte(""),
		nil, false,
	},
	{
		[]byte("\r\n"),
		nil, true,
	},
}

func TestGetAllRcvId(t *testing.T) {
	serial := newFakeSerial()
	im := &IM920{s: serial, readTimeout: 100 * time.Millisecond}

	for i, tt := range GetAllRcvIdTests {
		serial.dummyData = tt.in_dummyData
		ids, err := im.GetAllRcvId()
		if (tt.out_errorIsNil && (err != nil)) ||
			(!tt.out_errorIsNil && (err == nil)) {
			t.Errorf("[%d]GetId() => %v, want errorIsNil = %v",
				i, err, tt.out_errorIsNil)
		}
		if !reflect.DeepEqual(ids, tt.out) {
			t.Errorf("[%d]GetId() => %v, want data = %v",
				i, ids, tt.out)
		}
	}
}
