package im920

import (
	"bufio"
	"bytes"
	"container/list"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
    "sync"
	"time"

	"github.com/tarm/serial"
)

type Id uint16
type Ch uint8
type Rssi uint8
type Node uint8
type Mode uint8

const (
	FAST_MODE Mode = iota + 1
	LONG_MODE
)

type Config struct {
	Name        string
	ReadTimeout time.Duration
}

type ReadInfo struct {
	FromNode Node
	FromId   Id
	FromRssi Rssi
}

type IM920 struct {
	s            io.ReadWriteCloser
    m            *sync.Mutex
	readTimeout  time.Duration
	lastReadInfo ReadInfo
	rcvedData    *list.List
	isBusyFunc   func() bool
}

const (
	defaultBps       = 19200
	maxTXDA          = 64
	maxReadSize      = 256
	waitBusyTimeout  = 500 * time.Millisecond
	waitBusyInterval = 10 * time.Millisecond
)

func Open(c *Config) (*IM920, error) {
	t := (1000 * 100 * 8 / defaultBps) * time.Millisecond
	sc := &serial.Config{Name: c.Name, Baud: defaultBps, ReadTimeout: t}
	s, err := serial.OpenPort(sc)
	if err != nil {
		return &IM920{}, fmt.Errorf("error: OpenPort failed: %s", err)
	}

	return &IM920{s: s, m: new(sync.Mutex), readTimeout: c.ReadTimeout, rcvedData: list.New()}, nil
}

func strToUint16(s string) (val uint16, err error) {
	if len(s)%2 == 1 {
		s = "0" + s
	}

	b, derr := hex.DecodeString(s)
	if derr != nil {
		err = fmt.Errorf("error: Decode failed: %s (%s)", derr, b)
		return
	}

	switch len(b) {
	case 0:
		err = fmt.Errorf("error: Decode failed: data not found (%s)", s)
	case 1:
		val = uint16(b[0])
	case 2:
		val = binary.BigEndian.Uint16(b)
	default:
		err = fmt.Errorf("error: Decode failed: invalid size (%v)", len(b))
	}

	return
}

func parseReadHeaders(s string) (info ReadInfo, err error) {
	headers := strings.Split(s, ",")
	if len(headers) < 3 {
		err = fmt.Errorf("error: Split headers failed: %s", s)
		return
	}

	node, derr := hex.DecodeString(headers[0])
	if derr != nil {
		err = fmt.Errorf("error: Decode Node failed (%s): %s", headers[0], derr)
		return
	}
	info.FromNode = Node(node[0])

	id, derr := strToUint16(headers[1])
	if err != nil {
		err = fmt.Errorf("error: Decode FromId failed (%s): %s", headers[1], derr)
		return
	}
	info.FromId = Id(id)

	rssi, derr := hex.DecodeString(headers[2])
	if err != nil {
		err = fmt.Errorf("error: Decode Rssi failed (%s): %s", headers[2], derr)
		return
	}
	info.FromRssi = Rssi(rssi[0])

	return
}

func (im *IM920) IsBusyFunc(f func() bool) {
	im.isBusyFunc = f

	return
}

func (im *IM920) waitNotBusy() bool {
	if im.isBusyFunc == nil {
		return true
	}

	timer := time.NewTimer(waitBusyTimeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return false
		default:
			if !im.isBusyFunc() {
				return true
			}
			time.Sleep(waitBusyInterval)
		}
	}

	return false
}

func (im *IM920) receive(p []byte) (readed int, err error) {
	timer := time.NewTimer(im.readTimeout)
	defer timer.Stop()

	readedInitialbyte := false

	for {
		select {
		case <-timer.C:
			if readed == 0 {
				err = fmt.Errorf("error: Read failed: no data")
			}
			return
		default:
			n, rerr := im.s.Read(p[readed : readed+1])
			if rerr != nil && rerr != io.EOF {
				err = fmt.Errorf("error: Read failed: %s", rerr)
				return
			}
			if n > 0 && !readedInitialbyte {
				readedInitialbyte = true
			}
			if n == 0 && readedInitialbyte {
				return
			}

			readed += n
			if readed >= len(p) {
				return
			}
		}
	}
}

func (im *IM920) getResponse(p []byte) (readed int, err error) {
	rcv := make([]byte, maxReadSize)
	rcved, rerr := im.receive(rcv)
	if rerr != nil {
		err = fmt.Errorf("error: receive failed: %s", rerr)
		return
	}
	if rcved == 0 {
		err = fmt.Errorf("error: receive failed: no data")
		return
	}

	strs := strings.Split(string(rcv[:rcved]), "\r\n")

	for i, v := range strs {
		if !strings.Contains(v, ":") {
			resp := strings.Join(strs[i:], "\r\n")
			readed = copy(p, resp)
			break
		}

		if len(v) > 0 {
			if i+1 < len(strs) {
				v += "\r\n"
			}
			im.rcvedData.PushBack(v)
		}
	}

	return
}

func (im *IM920) IssueCommand(cmd, param string) (resp []byte, err error) {
    im.m.Lock()
    defer im.m.Unlock()
    
	if !im.waitNotBusy() {
		err = fmt.Errorf("error: BusyWait failed")
		return
	}

	_, werr := im.s.Write([]byte(cmd + " " + param + "\r\n"))
	if werr != nil {
		err = fmt.Errorf("error: Write failed: %s", werr)
		return
	}

	rcv := make([]byte, maxReadSize)
	rcved, rerr := im.getResponse(rcv)
	if rerr != nil {
		err = fmt.Errorf("error: getResponse failed: %s", rerr)
		return
	}
	if rcved == 0 {
		err = fmt.Errorf("error: getResponse failed: no data")
		return
	}

	if bytes.Equal(rcv[:rcved], []byte("NG\r\n")) {
		err = fmt.Errorf("NG response")
	}

	return rcv[:rcved], err
}

func (im *IM920) IssueCommandNormal(cmd, param string) error {
	resp, err := im.IssueCommand(cmd, param)
	if err != nil {
		return err
	}

	if !bytes.Equal(resp, []byte("OK\r\n")) {
		return fmt.Errorf("error: Unknown response (%v)", resp)
	}

	return nil
}

func (im *IM920) IssueCommandRespStr(cmd, param string) (resp string, err error) {
	rcv, err := im.IssueCommand(cmd, param)
	if err != nil {
		resp = strings.Replace(string(rcv), "\r\n", "", -1)
		return
	}

	dataEnd := strings.LastIndex(string(rcv), "\r\n")
	if strings.Index(string(rcv), "\r\n") < 0 {
		err = fmt.Errorf("error: not found the end of data (%v)", rcv)
		return
	}

	resp = string(rcv[:dataEnd])

	return
}

func (im *IM920) IssueCommandRespNum(cmd, param string) (resp uint16, err error) {
	rcv, err := im.IssueCommandRespStr(cmd, param)
	if err != nil {
		return
	}

	resp, err = strToUint16(rcv)

	return
}

func (im *IM920) IssueCommandRespNums(cmd, param string) (resp []uint16, err error) {
	rcv, err := im.IssueCommandRespStr(cmd, param)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader([]byte(rcv)))

	for scanner.Scan() {
		val, terr := strToUint16(scanner.Text())
		if terr != nil {
			err = terr
			return
		}
		resp = append(resp, val)
	}

	return
}

func (im *IM920) Write(p []byte) (n int, err error) {
	cmd := "TXDA"
	b2w := len(p)
	if len(p) > maxTXDA {
		b2w = maxTXDA
	}
	param := strings.ToUpper(hex.EncodeToString(p[:b2w]))

	err = im.IssueCommandNormal(cmd, param)
	if err != nil {
		n = 0
	} else {
		n = b2w
	}

	return
}

func (im *IM920) Read(p []byte) (n int, err error) {
    im.m.Lock()
    defer im.m.Unlock()
    
	str := ""
	if im.rcvedData.Len() > 0 {
		e := im.rcvedData.Front()
		str = e.Value.(string)
		im.rcvedData.Remove(e)
	} else {
		buf := make([]byte, maxReadSize)
		readed, err := im.receive(buf)
		if err != nil {
			return 0, fmt.Errorf("error: Read failed: %s", err)
		}
		if readed == 0 {
			return 0, fmt.Errorf("error: Read failed: no data")
		}
		str = string(buf)
	}

	strs := strings.Split(str, ":")
	if len(strs) < 2 {
		return 0, fmt.Errorf("error: Split header and data failed (%s)", str)
	}

	info, err := parseReadHeaders(strs[0])
	if err != nil {
		return 0, fmt.Errorf("error: parseReadHeaders failed: %s", err)
	}

	dataEnd := strings.Index(strs[1], "\r\n")
	if dataEnd < 0 {
		return 0, fmt.Errorf("error: not found the end of data (%s)", strs[1])
	}

	dataStr := strings.Replace(strs[1][:dataEnd], ",", "", -1)
	data, err := hex.DecodeString(dataStr)
	if err != nil {
		return 0, fmt.Errorf("error: Decode failed (%s): %s", dataStr, err)
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("error: Decode failed: no data")
	}

	n = copy(p, data)
	im.lastReadInfo = info

	return
}

func (im *IM920) LastReadInfo() ReadInfo {
	return im.lastReadInfo
}

func (im *IM920) GetId() (id Id, err error) {
	rcv, ierr := im.IssueCommandRespNum("RDID", "")
	if ierr != nil {
		err = fmt.Errorf("error: RDID failed: %s", ierr)
		return
	}

	id = Id(rcv)

	return
}

func (im *IM920) AddRcvId(id Id) (err error) {
	ierr := im.IssueCommandNormal("ENWR", "")
	if ierr != nil {
		err = fmt.Errorf("error: ENWR failed: %s", ierr)
		return
	}

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(id))
	ierr = im.IssueCommandNormal("SRID", hex.EncodeToString(b))
	if ierr != nil {
		err = fmt.Errorf("error: SRID failed: %s", ierr)
		return
	}

	ierr = im.IssueCommandNormal("DSWR", "")
	if ierr != nil {
		err = fmt.Errorf("error: DSWR failed: %s", ierr)
		return
	}

	return
}

func (im *IM920) GetAllRcvId() (ids []Id, err error) {
	rcv, ierr := im.IssueCommandRespNums("RRID", "")
	if ierr != nil {
		err = fmt.Errorf("error: RRID failed: %s", ierr)
		return
	}

	for _, v := range rcv {
		ids = append(ids, Id(v))
	}

	return
}

func (im *IM920) DeleteAllRcvId() error {
	ierr := im.IssueCommandNormal("ENWR", "")
	if ierr != nil {
		return fmt.Errorf("error: ENWR failed: %s", ierr)
	}

	ierr = im.IssueCommandNormal("ERID", "")
	if ierr != nil {
		return fmt.Errorf("error: ERID failed: %s", ierr)
	}

	ierr = im.IssueCommandNormal("DSWR", "")
	if ierr != nil {
		return fmt.Errorf("error: DSWR failed: %s", ierr)
	}

	return nil
}

func (im *IM920) SetCh(ch Ch, persist bool) (err error) {
	if persist {
		ierr := im.IssueCommandNormal("ENWR", "")
		if ierr != nil {
			err = fmt.Errorf("error: ENWR failed: %s", ierr)
			return
		}
	}

	b := make([]byte, 1)
	b[0] = byte(ch)
	ierr := im.IssueCommandNormal("STCH", hex.EncodeToString(b))
	if ierr != nil {
		err = fmt.Errorf("error: STCH failed: %s", ierr)
		return
	}

	if persist {
		ierr := im.IssueCommandNormal("DSWR", "")
		if ierr != nil {
			err = fmt.Errorf("error: DSWR failed: %s", ierr)
			return
		}
	}

	return
}

func (im *IM920) GetCh() (ch Ch, err error) {
	rcv, ierr := im.IssueCommandRespNum("RDCH", "")
	if ierr != nil {
		err = fmt.Errorf("error: RDCH failed: %s", ierr)
		return
	}

	ch = Ch(rcv)

	return
}

func (im *IM920) GetRssi() (rssi Rssi, err error) {
	rcv, ierr := im.IssueCommandRespNum("RDRS", "")
	if ierr != nil {
		err = fmt.Errorf("error: RDRS failed: %s", ierr)
		return
	}

	rssi = Rssi(rcv)

	return
}

func (im *IM920) SetCommMode(mode Mode, persist bool) (err error) {
	if persist {
		ierr := im.IssueCommandNormal("ENWR", "")
		if ierr != nil {
			err = fmt.Errorf("error: ENWR failed: %s", ierr)
			return
		}
	}

	b := make([]byte, 1)
	b[0] = byte(mode)
	ierr := im.IssueCommandNormal("STRT", hex.EncodeToString(b))
	if ierr != nil {
		err = fmt.Errorf("error: STRT failed: %s", ierr)
		return
	}

	if persist {
		ierr := im.IssueCommandNormal("DSWR", "")
		if ierr != nil {
			err = fmt.Errorf("error: DSWR failed: %s", ierr)
			return
		}
	}

	return
}

func (im *IM920) GetCommMode() (mode Mode, err error) {
	rcv, ierr := im.IssueCommandRespNum("RDRT", "")
	if ierr != nil {
		err = fmt.Errorf("error: RDRT failed: %s", ierr)
		return
	}

	mode = Mode(rcv)

	return
}

func (im *IM920) Close() error {
	return im.s.Close()
}
