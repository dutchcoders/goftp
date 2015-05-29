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

// send quit to the server and close the connection
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

// Send raw commands, return response as string and response code as int
func (ftp *FTP) RawCmd(command string, args ...interface{}) (code int, line string) {
	if ftp.debug {
		log.Printf("Raw-> %s\n", fmt.Sprintf(command, args...), code)
	}

	code = -1
	var err error
	if err = ftp.send(command, args...); err != nil {
		return code, ""
	}
	if line, err = ftp.receive(); err != nil {
		return code, ""
	}
	code, err = strconv.Atoi(line[:3])
	if ftp.debug {
		log.Printf("Raw<-	<- %d \n", code)
	}
	return code, line
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

// rename file
func (ftp *FTP) Rename(from string, to string) (err error) {
	if _, err = ftp.cmd("350", "RNFR %s", from); err != nil {
		return
	}

	if _, err = ftp.cmd("250", "RNTO %s", to); err != nil {
		return
	}

	return
}

// make directory
func (ftp *FTP) Mkd(path string) error {
	_, err := ftp.cmd("257", "MKD %s", path)
	return err
}

// get current path
func (ftp *FTP) Pwd() (path string, err error) {
	var line string
	if line, err = ftp.cmd("257", "PWD"); err != nil {
		return
	}

	res := REGEX_PWD_PATH.FindAllStringSubmatch(line[4:], -1)

	path = res[0][1]
	return
}

// change current path
func (ftp *FTP) Cwd(path string) (err error) {
	_, err = ftp.cmd("250", "CWD %s", path)
	return
}

// delete file
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

// secures the ftp connection by using TLS
func (ftp *FTP) AuthTLS(config tls.Config) error {
	if _, err := ftp.cmd("234", "AUTH TLS"); err != nil {
		return err
	}

	// wrap tls on existing connection
	ftp.tlsconfig = &config

	ftp.conn = tls.Client(ftp.conn, &config)
	ftp.writer = bufio.NewWriter(ftp.conn)
	ftp.reader = bufio.NewReader(ftp.conn)

	if _, err := ftp.cmd("200", "PBSZ 0"); err != nil {
		return err
	}

	if _, err := ftp.cmd("200", "PROT P"); err != nil {
		return err
	}

	return nil
}

// read all the buffered bytes and return
func (ftp *FTP) ReadAndDiscard() (int, error) {
	var i int
	var err error
	buffer_size := ftp.reader.Buffered()
	for i = 0; i < buffer_size; i++ {
		if _, err = ftp.reader.ReadByte(); err != nil {
			return i, err
		}
	}
	return i, err
}

// change transfer type
func (ftp *FTP) Type(t string) error {
	_, err := ftp.cmd("200", "TYPE %s", t)
	return err
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

// upload file
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

// retrieves file
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

/*func GetFilesList(path string) (files []string, err error) {

}*/

// list the path (or current directory)
func (ftp *FTP) List(path string) (files []string, err error) {
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
		// MLSD failed, lets try LIST
		if err = ftp.send("LIST %s", path); err != nil {
			return
		}

		if line, err = ftp.receive(); err != nil {
			return
		}

		if !strings.HasPrefix(line, "150") {
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

// login to the server
func (ftp *FTP) Login(username string, password string) (err error) {
	if _, err = ftp.cmd("331", "USER %s", username); err != nil {
		if strings.HasPrefix(err.Error(), "230") {
			// Ok, probably anonymous server
			// but login was fine, so return no error
			err = nil
		} else {
			return
		}
	}

	if _, err = ftp.cmd("230", "PASS %s", password); err != nil {
		return
	}

	return
}

// connect to server, debug is OFF
func Connect(addr string) (*FTP, error) {
	var err error
	var conn net.Conn

	if conn, err = net.Dial("tcp", addr); err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	//reader.ReadString('\n')
	object := &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: false}
	object.receive()

	return object, nil
}

// connect to server, debug is ON
func ConnectDbg(addr string) (*FTP, error) {
	var err error
	var conn net.Conn

	if conn, err = net.Dial("tcp", addr); err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	var line string

	object := &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: false}
	line, _ = object.receive()

	log.Print(line)

	return object, nil
}
