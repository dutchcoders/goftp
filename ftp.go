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

	supportedfeatures uint32
}

const (
	MLSD = 1
	NLST = 2
	EPLF = 4
	BAD  = 1 >> 32 // this server only supports LIST
//	STUBBORN = -2
)

func (ftp *FTP) Close() {
	ftp.conn.Close()
}

type WalkFunc func(path string, info os.FileMode, err error) error
type RetrFunc func(r io.Reader) error
type ErrorHandlerFunc func(pwd string, errorCode int, errorStr string, shouldBeSkippable bool) (skippable bool, err error)

// walks recursively through path and call walkfunc for each file.
// - links are ignored.
// - the optional parameter deepLimit controls the max level of recursion.
// - recursion stops only if errHandler returns false ( or if it's not defined )
// - Directories are traversed in pre-order
func (ftp *FTP) WalkCustom(path string, walkFn WalkFunc, errHandler ErrorHandlerFunc, deepLimit ...int) (err error) {
	deep := -1
	if len(deepLimit) > 0 {
		deep = deepLimit[0]
	}

	if ftp.debug {
		log.Printf("Walking: '%s'\n", path)
		if len(deepLimit) > 0 {
			log.Printf("Deep limit is: '%d'\n", deepLimit[0])
		}

	}

	files, dirs, _, err := ftp.GetFilesList(path)

	if err != nil {
		if ftp.debug {
			log.Println("1 " + err.Error())
			log.Println("1--> " + path)
		}
		return
	}

	for _, subpath := range files {
		err = walkFn(subpath, os.FileMode(0), nil)
		if err != nil {
			if ftp.debug {
				log.Println("2 " + err.Error())
				log.Println("2--> " + subpath)
			}
			if errHandler == nil {
				return
			}
			code, _ := strconv.Atoi(err.Error()[:3])
			flag, e := errHandler(subpath, code, err.Error(), code == 540)
			if !flag || e != nil {
				return
			}
		}
	}
	for _, subpath := range dirs {
		if deep > 0 {
			err = ftp.WalkCustom(subpath, walkFn, errHandler, deep-1)
		} else if deep < 0 {
			err = ftp.WalkCustom(subpath, walkFn, errHandler)
		} else if deep == 0 {
			log.Println("Deep limit reached")
		}
		if err != nil {
			log.Println("3 " + err.Error())
			log.Println("3--> " + subpath)
			if errHandler == nil {
				return
			}
			code, _ := strconv.Atoi(err.Error()[:3])
			flag, e := errHandler(subpath, code, err.Error(), code == 550)
			if !flag || e != nil {
				return
			}
		}
	}

	return
}

// walks recursively through path and call walkfunc for each file.
// - links are ignored.
// - the optional parameter deepLimit controls the max level of recursion.
// - recursion stops on first error anyways.
// - Directories are traversed in pre-order
func (ftp *FTP) Walk(path string, walkFn WalkFunc, deepLimit ...int) (err error) {
	deep := -1
	if len(deepLimit) > 0 {
		deep = deepLimit[0]
	}
	return ftp.WalkCustom(path, walkFn, nil, deep)
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

// Open a passive connection with pasv, send a raw command, retrieve the response, close the connection, return the response
func (ftp *FTP) RawPassiveCmd(command string) (code int, response []string) {
	var port int
	var pconn net.Conn
	var line string
	var err error
	//var msg string

	// get the port
	if port, err = ftp.Pasv(); err != nil {
		return -1, nil
	}

	//send the request
	err = ftp.send(command)
	if err != nil {
		return -2, nil
	}

	//open a connection to retrieve data
	if pconn, err = ftp.newConnection(port); err != nil {
		return -3, nil
	}

	//retrieve response
	if line, err = ftp.receive(); err != nil {
		return -4, nil
	}
	code, err = strconv.Atoi(line[:3])
	if err != nil {
		return -5, nil
	}
	if code > 299 {
		return code, nil
	}

	reader := bufio.NewReader(pconn)

	for {
		line, err = reader.ReadString('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			pconn.Close()
			return -5, nil
		}
		response = append(response, string(line))
	}

	pconn.Close()

	if line, err = ftp.receive(); err != nil {
		return -6, nil
	}

	if code > 299 {
		return code, nil
	}

	return code, response
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
func (ftp *FTP) readAndDiscard() (int, error) {
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
	ftp.readAndDiscard()
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
	object := &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: false, supportedfeatures: 0}
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

	object := &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: true, supportedfeatures: 0}
	line, _ = object.receive()

	log.Print(line)

	return object, nil
}

// list the path (or current directory). return raw listing, do not parse it.
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
