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
	"strconv"
	"strings"
)

var REGEX_PWD_PATH *regexp.Regexp = regexp.MustCompile(`\"(.*)\"`)

type FTP struct {
	conn net.Conn

	addr string

	debug     bool
	tlsconfig *tls.Config

	reader *bufio.Reader
	writer *bufio.Writer
}

func (ftp *FTP) Close() {
	ftp.conn.Close()
}

type WalkFunc func(path string, info os.FileMode, err error) error
type RetrFunc func(r io.Reader) error

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

// walks recursively through path and call walkfunc for each file
func (ftp *FTP) Walk(path string, walkFn WalkFunc) (err error) {
	/*
		if err = walkFn(path, os.ModeDir, nil); err != nil {
			if err == filepath.SkipDir {
				return nil
			}
		}
	*/
	if ftp.debug {
		log.Printf("Walking: '%s'\n", path)
	}

	var lines []string

	if lines, err = ftp.List(path); err != nil {
		return
	}

	for _, line := range lines {
		_, t, subpath := parseLine(line)

		switch t {
		case "dir":
			if subpath == "." {
			} else if subpath == ".." {
			} else {
				if err = ftp.Walk(path+subpath+"/", walkFn); err != nil {
					return
				}
			}
		case "file":
			if err = walkFn(path+subpath, os.FileMode(0), nil); err != nil {
				return
			}
		}
	}

	return
}

func (ftp *FTP) Quit() (err error) {
	if _, err := ftp.cmd("221", "QUIT"); err != nil {
		return err
	}

	ftp.conn.Close()
	ftp.conn = nil

	return nil
}

// will send a NOOP (no operation) to the server
func (ftp *FTP) Noop() (err error) {
	_, err = ftp.cmd("200", "NOOP")
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

func (ftp *FTP) Rename(from string, to string) (err error) {
	if _, err = ftp.cmd("350", "RNFR %s", from); err != nil {
		return
	}

	if _, err = ftp.cmd("250", "RNTO %s", to); err != nil {
		return
	}

	return
}
func (ftp *FTP) Mkd(path string) (err error) {
	_, err = ftp.cmd("257", "MKD %s", path)
	return
}

func (ftp *FTP) Pwd() (path string, err error) {
	var line string
	if line, err = ftp.cmd("257", "PWD"); err != nil {
		return
	}

	res := REGEX_PWD_PATH.FindAllStringSubmatch(line[4:], -1)

	path = res[0][1]
	return
}

func (ftp *FTP) Cwd(path string) (err error) {
	_, err = ftp.cmd("250", "CWD %s", path)
	return
}

func (ftp *FTP) Dele(path string) (err error) {
	if err = ftp.send("DELE %s", path); err != nil {
		return
	}

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "250") {
		return errors.New(line)
	}

	return
}

func (ftp *FTP) AuthTLS(config tls.Config) (err error) {
	if _, err = ftp.cmd("234", "AUTH TLS"); err != nil {
		return
	}

	// wrap tls on existing connection
	ftp.tlsconfig = &config

	ftp.conn = tls.Client(ftp.conn, &config)
	ftp.writer = bufio.NewWriter(ftp.conn)
	ftp.reader = bufio.NewReader(ftp.conn)

	if _, err = ftp.cmd("200", "PROT P"); err != nil {
		return
	}

	return
}

func (ftp *FTP) Type(t string) (err error) {
	_, err = ftp.cmd("200", "TYPE %s", t)
	return
}

func (ftp *FTP) receive() (line string, err error) {
	line, err = ftp.reader.ReadString('\n')

	if ftp.debug {
		log.Printf("< %s", line)
	}

	return
}

func (ftp *FTP) send(command string, arguments ...interface{}) (err error) {
	if ftp.debug {
		log.Printf("> %s", fmt.Sprintf(command, arguments...))
	}

	command = fmt.Sprintf(command, arguments...)
	command += "\r\n"


	if _, err = ftp.writer.WriteString(command); err != nil {
		return
	}

	err = ftp.writer.Flush()

	return
}

// enables passive data connection and returns port number
func (ftp *FTP) Pasv() (port int, err error) {
	var line string
	if line, err = ftp.cmd("227", "PASV"); err != nil {
		return
	}

	re, err := regexp.Compile(`\((.*)\)`)

	res := re.FindAllStringSubmatch(line, -1)

	s := strings.Split(res[0][1], ",")

	l1, _ := strconv.Atoi(s[len(s)-2])
	l2, _ := strconv.Atoi(s[len(s)-1])

	port = l1<<8 + l2

	return
}

func (ftp *FTP) newConnection(port int) (conn net.Conn, err error) {
	if ftp.debug {
		log.Printf("Connecting to %s\n", addr)
	}

	if conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ftp.addr, port)); err != nil {
		return
	}

	if ftp.tlsconfig != nil {
		conn = tls.Client(conn, ftp.tlsconfig)
	}

	return
}

func (ftp *FTP) Stor(path string, r io.Reader) (err error) {
	if err = ftp.Type("I"); err != nil {
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

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "150") {
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

	if !strings.HasPrefix(line, "226") {
		err = errors.New(line)
		return
	}

	return

}

func (ftp *FTP) Retr(path string, retrFn RetrFunc) (s string, err error) {
	if err = ftp.Type("I"); err != nil {
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

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "150") {
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

	if !strings.HasPrefix(line, "226") {
		err = errors.New(line)
		return
	}

	return
}

func (ftp *FTP) List(path string) (files []string, err error) {
	if err = ftp.Type("A"); err != nil {
		return
	}

	var port int
	if port, err = ftp.Pasv(); err != nil {
		return
	}

	// _, err = ftp.writer.WriteString(fmt.Sprintf("LIST %s\r\n", path))
	// check for features LIST / MLSD
	if err = ftp.send("MLSD %s", path); err != nil {
		return
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
		err = errors.New(line)
		return
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

func (ftp *FTP) Login(username string, password string) (err error) {
	if _, err = ftp.cmd("331", "USER %s", username); err != nil {
		return
	}

	if _, err = ftp.cmd("230", "PASS %s", password); err != nil {
		return
	}

	return
}

func Connect(addr string) (*FTP, error) {
	var err error
	var conn net.Conn

	if conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", addr, 21)); err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	var line string

	line, err = reader.ReadString('\n')
	fmt.Print(line)

	return &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: true}, nil
}
