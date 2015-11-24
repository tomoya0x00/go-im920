package im920

import (
	"encoding/hex"
	"fmt"
	"github.com/tarm/serial"
	"strings"
	"time"
)

type Config struct {
	Name        string
	ReadTimeout time.Duration
}

type IM920 struct {
	s *serial.Port
}

const (
	defaultBps = 19200
	maxTXDA    = 64
	respSize   = 4
)

func Open(c *Config) (*IM920, error) {
	sc := &serial.Config{Name: c.Name, Baud: defaultBps, ReadTimeout: c.ReadTimeout}
	s, err := serial.OpenPort(sc)
	if err != nil {
		return &IM920{}, fmt.Errorf("Filed to open: %s", err)
	}

	return &IM920{s}, nil
}

// TODO: Improve error handling
func (im *IM920) Write(p []byte) (n int, err error) {
	command := "TXDA "

	b2w := len(p)
	if len(p) > maxTXDA {
		b2w = maxTXDA
	}
	param := strings.ToUpper(hex.EncodeToString(p[:b2w]))

	// TODO: BUSY WAIT
	im.s.Write([]byte(command))
	im.s.Write([]byte(param))
	im.s.Write([]byte("\r\n"))

	resp := make([]byte, respSize)
	readed, err := im.s.Read(resp)

	switch string(resp[:readed]) {
	case "OK\r\n":
		n = b2w
		err = nil
	case "NG\r\n":
		err = fmt.Errorf("NG response")
	default:
		err = fmt.Errorf("Unknown response")
	}

	return n, err
}

func (im *IM920) Close() error {
	return im.s.Close()
}
