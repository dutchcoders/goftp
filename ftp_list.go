package goftp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func (ftp *FTP) updateFeat() (features uint32) {
	fmt.Println("Updating features list")
	code, str := ftp.RawCmd("FEAT")
	fmt.Println(str)
	if code < 0 || code > 299 {
		return 0
	}
	return 1
}

func (ftp *FTP) GetFilesList(path string) (files []string, directories []string, err error) {
	ftp.updateFeat()
	return
}

// list the path (or current directory). return raw listing, do not parse it.
func (ftp *FTP) List(path string) (files []string, err error) {
	mlsd := true
	nlst := true
	eplf := true
	list := true
	//
	if err = ftp.Type("A"); err != nil {
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

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "150") {
		mlsd = false
		// MLSD failed, lets try LIST
		if err = ftp.send("LIST %s", path); err != nil {
			return
		}

		if line, err = ftp.receive(); err != nil {
			return
		}

		if !strings.HasPrefix(line, "150") {
			list = false
			// Really list is not working here
			err = errors.New(line)
			return
		}
	}

	if ftp.debug {
		if !mlsd {
			log.Printf("MLSD not supported")
			if !nlst {
				log.Printf("NLST not supported")
				if !eplf {
					log.Printf("EPLF not supported")
					if !list {
						log.Printf("LIST not supported (this should not appen!)")
					}
				}
			}
		}
	}

	reader := bufio.NewReader(pconn)

	files, err = ftp.splitLines(reader)

	if err != nil {
		return nil, err
	}

	pconn.Close()

	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "226") {
		err = errors.New(line)
		return
	}

	return
}

func (ftp *FTP) splitLines(reader *bufio.Reader) (files []string, err error) {
	var line string
	for {
		line, err = reader.ReadString('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		files = append(files, string(line))
	}
	return files, nil
}
