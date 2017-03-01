package goftp

import (
	"errors"
	"strings"
)

// Syst returns the system type of the remote host
func (ftp *FTP) Syst() (line string, err error) {
	if err := ftp.send("SYST"); err != nil {
		return "", err
	}
	if line, err = ftp.receive(); err != nil {
		return
	}
	if !strings.HasPrefix(line, StatusSystemType) {
		err = errors.New(line)
		return
	}

	return strings.SplitN(strings.TrimSpace(line), " ", 2)[1], nil
}
