package goftp

import (
	"errors"
	"io"
	"net"
	"strings"
)

// Stor uploads file to remote host path, from r
func (ftp *FTP) Stor(path string, r io.Reader) (err error) {
	if err = ftp.Type(TypeImage); err != nil {
		return
	}

	var port int
	if port, err = ftp.Pasv(); err != nil {
		return
	}

	if err = ftp.send("STOR %s", path); err != nil {
		return
	}

	var pconn net.Conn
	if pconn, err = ftp.newConnection(port); err != nil {
		return
	}
	defer pconn.Close()

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, StatusFileOK) {
		err = errors.New(line)
		return
	}

	if _, err = io.Copy(pconn, r); err != nil {
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
