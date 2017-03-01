package goftp

import (
	"errors"
	"strings"
)

// Dele deletes path on remote host
func (ftp *FTP) Dele(path string) (err error) {
	if err = ftp.send("DELE %s", path); err != nil {
		return
	}

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, StatusActionOK) {
		return errors.New(line)
	}

	return
}
