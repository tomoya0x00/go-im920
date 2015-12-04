package im920

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type Config struct {
	Name        string
	ReadTimeout time.Duration
}

type ReadInfo struct {
	Node   uint8
	FromId uint16
	Rssi   uint8
}

type IM920 struct {
	s            io.ReadWriteCloser
	readTimeout  time.Duration
	lastReadInfo ReadInfo
}

const (
	defaultBps  = 19200
	maxTXDA     = 64
	maxReadSize = 256
)

type Id uint16

func Open(c *Config) (*IM920, error) {
	t := (1000 * 100 * 8 / defaultBps) * time.Millisecond
	sc := &serial.Config{Name: c.Name, Baud: defaultBps, ReadTimeout: t}
	s, err := serial.OpenPort(sc)
	if err != nil {
		return &IM920{}, fmt.Errorf("error: OpenPort failed: %s", err)
	}

	return &IM920{s: s, readTimeout: c.ReadTimeout}, nil
}

func strToUint16(s string) (val uint16, err error) {
	b, err := hex.DecodeString(s)
	if err != nil {
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
	info.Node = uint8(node[0])

	info.FromId, derr = strToUint16(headers[1])
	if err != nil {
		err = fmt.Errorf("error: Decode FromId failed (%s): %s", headers[1], derr)
		return
	}

	rssi, derr := hex.DecodeString(headers[2])
	if err != nil {
		err = fmt.Errorf("error: Decode Rssi failed (%s): %s", headers[2], derr)
		return
	}
	info.Rssi = uint8(rssi[0])

	return
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
			if rerr != nil {
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

func (im *IM920) IssueCommand(cmd, param string) (resp []byte, err error) {
	// TODO: BUSY WAIT
	s := []string{cmd, param, "\r\n"}
	_, werr := im.s.Write([]byte(strings.Join(s, " ")))
	if werr != nil {
		err = fmt.Errorf("error: Write failed: %s", werr)
		return
	}

	// TODO: BUSY WAIT
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

	if len(resp) == 0 {
		err = fmt.Errorf("error: receive failed: no data")
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
	buf := make([]byte, maxReadSize)

	readed, err := im.receive(buf)
	if err != nil {
		return 0, fmt.Errorf("error: Read failed: %s", err)
	}
	if readed == 0 {
		return 0, fmt.Errorf("error: Read failed: no data")
	}

	strs := strings.Split(string(buf), ":")
	if len(strs) < 2 {
		return 0, fmt.Errorf("error: Split header and data failed (%s)", string(buf))
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

func (im *IM920) Close() error {
	return im.s.Close()
}
