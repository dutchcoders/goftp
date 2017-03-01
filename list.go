package goftp

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
)

// List lists the path (or current directory)
func (ftp *FTP) List(path string) (files []string, err error) {
	if err = ftp.Type(TypeASCII); err != nil {
		return
	}

	var port int
	if port, err = ftp.Pasv(); err != nil {
		return
	}

	// check if MLSD works
	if err = ftp.send("MLSD %s", path); err != nil {
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
		// MLSD failed, lets try LIST
		if err = ftp.send("LIST %s", path); err != nil {
			return
		}

		if line, err = ftp.receiveNoDiscard(); err != nil {
			return
		}

		if !strings.HasPrefix(line, StatusFileOK) {
			// Really list is not working here
			err = errors.New(line)
			return
		}
	}

	reader := bufio.NewReader(pconn)

	for {
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		files = append(files, string(line))
	}
	// Must close for vsftp tlsed conenction otherwise does not receive connection
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
