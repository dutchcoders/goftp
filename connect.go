package goftp

import (
	"crypto/tls"
	"bufio"
	"log"
	"net"
)

// ConnectTLS (FTP over TLS).
func ConnectTLS(addr string, config *tls.Config, debug bool) (*FTP, error) {
	var err error
	var conn net.Conn

	if conn, err = net.Dial("tcp", addr); err != nil {
		return nil, err
	}
	conn = tls.Client(conn, config)

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	// reader.ReadString('\n')
	object := &FTP{conn: conn, addr: addr, reader: reader, writer: writer, tlsconfig: config, debug: debug}
	object.receive()

	return object, nil
}

// Connect to server at addr (format "host:port"). debug is OFF
func Connect(addr string) (*FTP, error) {
	var err error
	var conn net.Conn

	if conn, err = net.Dial("tcp", addr); err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	// reader.ReadString('\n')
	object := &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: false}
	object.receive()

	return object, nil
}

// ConnectDbg to server at addr (format "host:port"). debug is ON
func ConnectDbg(addr string) (*FTP, error) {
	var err error
	var conn net.Conn

	if conn, err = net.Dial("tcp", addr); err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	var line string

	object := &FTP{conn: conn, addr: addr, reader: reader, writer: writer, debug: true}
	line, _ = object.receive()

	log.Print(line)

	return object, nil
}
