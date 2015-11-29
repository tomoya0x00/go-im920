package im920

import (
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

type IM920 struct {
	s           io.ReadWriteCloser
	readTimeout time.Duration
}

const (
	defaultBps  = 19200
	maxTXDA     = 64
	respSize    = 4
	maxReadSize = 256
)

func Open(c *Config) (*IM920, error) {
    t := (1000 * 100 * 8 / defaultBps) * time.Millisecond
	sc := &serial.Config{Name: c.Name, Baud: defaultBps, ReadTimeout: t}
	s, err := serial.OpenPort(sc)
	if err != nil {
		return &IM920{}, fmt.Errorf("Filed to open: %s", err)
	}

	return &IM920{s: s, readTimeout: c.ReadTimeout}, nil
}

func (im *IM920) receive(p []byte) (readed int, err error) {
	timer := time.NewTimer(im.readTimeout)
    defer timer.Stop()

    readedInitialbyte := false
    
	for {
		select {
		case <-timer.C:
			if readed == 0 {
				err = fmt.Errorf("Failed to read: no data")
			}
			return
		default:
            n, rerr := im.s.Read(p[readed:readed+1])
			if rerr != nil {
				err = fmt.Errorf("Failed to read : %s", rerr)
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

func (im *IM920) IssueCommand(cmd, param string) error {
	// TODO: BUSY WAIT
	s := []string{cmd, param, "\r\n"}
	_, err := im.s.Write([]byte(strings.Join(s, " ")))
	if err != nil {
		return fmt.Errorf("Failed to write: %s", err)
	}

	// TODO: BUSY WAIT
	resp := make([]byte, respSize)
	readed, err := im.receive(resp)

	switch string(resp[:readed]) {
	case "OK\r\n":
		err = nil
	case "NG\r\n":
		err = fmt.Errorf("NG response")
	default:
		err = fmt.Errorf("Unknown response: %v", resp[:readed])
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

	readed, err := im.receive(buf)
	if err != nil {
		return 0, fmt.Errorf("Failed to read: %s", err)
	}
	if readed == 0 {
		return 0, fmt.Errorf("Failed to read: no data")
	}

	s := string(buf)

	dataStart := strings.Index(s, ":")
	if dataStart < 0 {
		return 0, fmt.Errorf("Failed to find the start of data : %s", s)
	}

	dataEnd := strings.Index(s, "\r\n")
	if dataEnd < 0 {
		return 0, fmt.Errorf("Failed to find the end of data : %s", s)
	} 

	dataStr := strings.Replace(s[dataStart+1:dataEnd], ",", "", -1)
	read, err := hex.DecodeString(dataStr)
	if err != nil {
		return 0, fmt.Errorf("Failed to decode : %s", err)
	}
	if len(read) == 0 {
		return 0, fmt.Errorf("Failed to decode : no data (%s)", s)
	}

	n = copy(p, read)

	return
}

func (im *IM920) Close() error {
	return im.s.Close()
}
