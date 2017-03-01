package goftp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

// RePwdPath is the default expression for matching files in the current working directory
var RePwdPath = regexp.MustCompile(`\"(.*)\"`)

// FTP is a session for File Transfer Protocol
type FTP struct {
	conn net.Conn

	addr string

	debug     bool
	tlsconfig *tls.Config

	reader *bufio.Reader
	writer *bufio.Writer
}

type (
	// WalkFunc is called on each path in a Walk. Errors are filtered through WalkFunc
	WalkFunc func(path string, info os.FileMode, err error) error

	// RetrFunc is passed to Retr and is the handler for the stream received for a given path
	RetrFunc func(r io.Reader) error
)

func parseLine(line string) (perm string, t string, filename string) {
	for _, v := range strings.Split(line, ";") {
		v2 := strings.Split(v, "=")

		switch v2[0] {
		case "perm":
			perm = v2[1]
		case "type":
			t = v2[1]
		default:
			filename = v[1 : len(v)-2]
		}
	}
	return
}

// private function to send command and compare return code with expects
func (ftp *FTP) cmd(expects string, command string, args ...interface{}) (line string, err error) {
	if err = ftp.send(command, args...); err != nil {
		return
	}

	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, expects) {
		err = errors.New(line)
		return
	}

	return
}

func (ftp *FTP) receiveLine() (string, error) {
	line, err := ftp.reader.ReadString('\n')

	if ftp.debug {
		log.Printf("< %s", line)
	}

	return line, err
}

func (ftp *FTP) receive() (string, error) {
	line, err := ftp.receiveLine()

	if err != nil {
		return line, err
	}

	if (len(line) >= 4) && (line[3] == '-') {
		//Multiline response
		closingCode := line[:3] + " "
		for {
			str, err := ftp.receiveLine()
			line = line + str
			if err != nil {
				return line, err
			}
			if len(str) < 4 {
				if ftp.debug {
					log.Println("Uncorrectly terminated response")
				}
				break
			} else {
				if str[:4] == closingCode {
					break
				}
			}
		}
	}
	ftp.ReadAndDiscard()
	//fmt.Println(line)
	return line, err
}

func (ftp *FTP) receiveNoDiscard() (string, error) {
	line, err := ftp.receiveLine()

	if err != nil {
		return line, err
	}

	if (len(line) >= 4) && (line[3] == '-') {
		//Multiline response
		closingCode := line[:3] + " "
		for {
			str, err := ftp.receiveLine()
			line = line + str
			if err != nil {
				return line, err
			}
			if len(str) < 4 {
				if ftp.debug {
					log.Println("Uncorrectly terminated response")
				}
				break
			} else {
				if str[:4] == closingCode {
					break
				}
			}
		}
	}
	//ftp.ReadAndDiscard()
	//fmt.Println(line)
	return line, err
}

func (ftp *FTP) send(command string, arguments ...interface{}) error {
	if ftp.debug {
		log.Printf("> %s", fmt.Sprintf(command, arguments...))
	}

	command = fmt.Sprintf(command, arguments...)
	command += "\r\n"

	if _, err := ftp.writer.WriteString(command); err != nil {
		return err
	}

	if err := ftp.writer.Flush(); err != nil {
		return err
	}

	return nil
}

// open new data connection
func (ftp *FTP) newConnection(port int) (conn net.Conn, err error) {
	addr := fmt.Sprintf("%s:%d", strings.Split(ftp.addr, ":")[0], port)

	if ftp.debug {
		log.Printf("Connecting to %s\n", addr)
	}

	if conn, err = net.Dial("tcp", addr); err != nil {
		return
	}

	if ftp.tlsconfig != nil {
		conn = tls.Client(conn, ftp.tlsconfig)
	}

	return
}

// System types from Syst
var (
	SystemTypeUnixL8    = "UNIX Type: L8"
	SystemTypeWindowsNT = "Windows_NT"
)

var reSystStatus = map[string]*regexp.Regexp{
	SystemTypeUnixL8:    regexp.MustCompile(""),
	SystemTypeWindowsNT: regexp.MustCompile(""),
}

/*func GetFilesList(path string) (files []string, err error) {

}*/

/*


// login on server with strange login behavior
func (ftp *FTP) SmartLogin(username string, password string) (err error) {
	var code int
	// Maybe the server has some useless words to say. Make him talk
	code, _ = ftp.RawCmd("NOOP")

	if code == 220 || code == 530 {
		// Maybe with another Noop the server will ask us to login?
		code, _ = ftp.RawCmd("NOOP")
		if code == 530 {
			// ok, let's login
			code, _ = ftp.RawCmd("USER %s", username)
			code, _ = ftp.RawCmd("NOOP")
			if code == 331 {
				// user accepted, password required
				code, _ = ftp.RawCmd("PASS %s", password)
				code, _ = ftp.RawCmd("PASS %s", password)
				if code == 230 {
					code, _ = ftp.RawCmd("NOOP")
					return
				}
			}
		}

	}
	// Nothing strange... let's try a normal login
	return ftp.Login(username, password)
}

*/
