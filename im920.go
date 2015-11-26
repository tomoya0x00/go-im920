package im920

import (
	"encoding/hex"
	"fmt"
	"github.com/tarm/serial"
	"io"
	"strings"
	"time"
)

type Config struct {
	Name        string
	ReadTimeout time.Duration
}

type IM920 struct {
	s io.ReadWriteCloser
}

const (
	defaultBps  = 19200
	maxTXDA     = 64
	respSize    = 4
	maxReadSize = 256
)

func Open(c *Config) (*IM920, error) {
	sc := &serial.Config{Name: c.Name, Baud: defaultBps, ReadTimeout: c.ReadTimeout}
	s, err := serial.OpenPort(sc)
	if err != nil {
		return &IM920{}, fmt.Errorf("Filed to open: %s", err)
	}

	return &IM920{s}, nil
}

func (im *IM920) IssueCommand(cmd, param string) error {
	// TODO: BUSY WAIT
	s := []string{cmd, param, "\r\n"}
	_, err := im.s.Write([]byte(strings.Join(s, " ")))
	if err != nil {
		return fmt.Errorf("Failed to write: %s", err)
	}

	// TODO: BUSY WAIT
	resp := make([]byte, respSize)
	readed, err := im.s.Read(resp)
	if err != nil {
		return fmt.Errorf("Failed to read: %s", err)
	}

	switch string(resp[:readed]) {
	case "OK\r\n":
		err = nil
	case "NG\r\n":
		err = fmt.Errorf("NG response")
	default:
		err = fmt.Errorf("Unknown response")
	}

	return err
}

func (im *IM920) Write(p []byte) (n int, err error) {
	cmd := "TXDA"
	b2w := len(p)
	if len(p) > maxTXDA {
		b2w = maxTXDA
	}
	param := strings.ToUpper(hex.EncodeToString(p[:b2w]))

	err = im.IssueCommand(cmd, param)
	if err != nil {
		n = 0
	} else {
		n = b2w
	}

	return
}

func (im *IM920) Read(p []byte) (n int, err error) {
	buf := make([]byte, maxReadSize)

	readed, err := im.s.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("Failed to read: %s", err)
	}
	if readed == 0 {
		return 0, fmt.Errorf("Failed to read: no data")
	}

	s := string(buf)

	dataStart := strings.Index(s, ":")
	if dataStart < 0 {
		return 0, fmt.Errorf("Failed to find the start of data")
	}

	dataEnd := strings.Index(s, "\r\n")
	if dataEnd < 0 {
		return 0, fmt.Errorf("Failed to find the end of data")
	}

	dataStr := strings.Replace(s[dataStart+1:dataEnd], ",", "", -1)
	read, err := hex.DecodeString(dataStr)
	if err != nil {
		return 0, fmt.Errorf("Failed to decode : %s", err)
	}
	if len(read) == 0 {
		return 0, fmt.Errorf("Failed to decode: no data")
	}

	n = copy(p, read)

	return
}

func (im *IM920) Close() error {
	return im.s.Close()
}
