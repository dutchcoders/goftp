package goftp

import (
	"errors"
	"net"
	"strings"
)

// Retr retrieves file from remote host at path, using retrFn to read from the remote file.
func (ftp *FTP) Retr(path string, retrFn RetrFunc) (s string, err error) {
	if err = ftp.Type(TypeImage); err != nil {
		return
	}

	var port int
	if port, err = ftp.Pasv(); err != nil {
		return
	}

	if err = ftp.send("RETR %s", path); err != nil {
		return
	}

	var pconn net.Conn
	if pconn, err = ftp.newConnection(port); err != nil {
		return
	}
	defer pconn.Close()

	var line string
	if line, err = ftp.receiveNoDiscard(); err != nil {
		return
	}

	if !strings.HasPrefix(line, StatusFileOK) {
		err = errors.New(line)
		return
	}

	if err = retrFn(pconn); err != nil {
		return
	}

	pconn.Close()

	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, StatusClosingDataConnection) {
		err = errors.New(line)
		return
	}

	return
}
