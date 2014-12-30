package goftp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type FTP struct {
	conn net.Conn

	addr string

	debug bool

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

func (ftp *FTP) Walk(path string, walkFn WalkFunc) (err error) {
	/*
		if err = walkFn(path, os.ModeDir, nil); err != nil {
			if err == filepath.SkipDir {
				return nil
			}
		}
	*/
	if ftp.debug {
		fmt.Printf("Walking: '%s'\n", path)
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
		default:
		}
	}

	return
}

func (ftp *FTP) CWD(path string) (err error) {
	if err = ftp.send("CWD %s", path); err != nil {
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

func (ftp *FTP) Type(t string) (err error) {
	if err = ftp.send("TYPE %s", t); err != nil {
		return
	}

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "200") {
		return errors.New(line)
	}

	return
}

func (ftp *FTP) receive() (line string, err error) {
	line, err = ftp.reader.ReadString('\n')
	if ftp.debug {
		fmt.Printf("< %s\n", line)
	}

	return
}

func (ftp *FTP) send(command string, arguments ...interface{}) (err error) {
	command = fmt.Sprintf(command, arguments...)
	command += "\r\n"

	if ftp.debug {
		fmt.Printf("> %s", command)
	}

	if _, err = ftp.writer.WriteString(command); err != nil {
		return
	}

	err = ftp.writer.Flush()

	return
}

func (ftp *FTP) Pasv() (port int, err error) {
	if err = ftp.send("PASV"); err != nil {
		return
	}

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "227") {
		err = errors.New(line)
		return
	}

	re, err := regexp.Compile(`\((.*)\)`)

	res := re.FindAllStringSubmatch(line, -1)

	s := strings.Split(res[0][1], ",")

	l1, _ := strconv.Atoi(s[len(s)-2])
	l2, _ := strconv.Atoi(s[len(s)-1])

	port = l1*256 + l2

	return
}

func (ftp *FTP) Retr(path string, f RetrFunc) (s string, err error) {
	if err = ftp.Type("I"); err != nil {
		return
	}

	var port int
	if port, err = ftp.Pasv(); err != nil {
		return
	}

	// _, err = ftp.writer.WriteString(fmt.Sprintf("LIST %s\r\n", path))
	// check for features LIST / MLSD
	if err = ftp.send("RETR %s", path); err != nil {
		return
	}

	var pconn net.Conn

	if ftp.debug {
		fmt.Println(fmt.Sprintf("Connecting to %s:%d", ftp.addr, port))
	}

	if pconn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ftp.addr, port)); err != nil {
		return "", err
	}

	var line string
	if line, err = ftp.receive(); err != nil {
		return
	}

	if !strings.HasPrefix(line, "150") {
		err = errors.New(line)
		return
	}

	if err = f(pconn); err != nil {
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

	if ftp.debug {
		fmt.Println(fmt.Sprintf("Connecting to %s:%d", ftp.addr, port))
	}

	if pconn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ftp.addr, port)); err != nil {
		return nil, err
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
	_, err = ftp.writer.WriteString(fmt.Sprintf("USER %s\r\n", username))
	ftp.writer.Flush()

	var line string

	line, err = ftp.reader.ReadString('\n')
	if !strings.HasPrefix(line, "331") {
		return errors.New(line)
	}

	_, err = ftp.writer.WriteString(fmt.Sprintf("PASS %s\r\n", password))
	ftp.writer.Flush()

	line, err = ftp.reader.ReadString('\n')
	if !strings.HasPrefix(line, "230") {
		return errors.New(line)
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

	return &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: false}, nil
}
