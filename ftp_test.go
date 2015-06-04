package goftp

import "testing"
import "os"

import "fmt"

var goodServer string
var uglyServer string
var badServer string

func init() {
	//ProFTPD 1.3.5 Server (Debian)
	goodServer = "bo.mirror.garr.it:21"

	//Symantec EMEA FTP Server
	badServer = "ftp.packardbell.com:21"

	//Unknown server
	uglyServer = "ftp.musicbrainz.org:21"
}

func active(host string) (msg string) {
	var err error
	var connection *FTP

	if connection, err = Connect(host); err != nil {
		return "Can't connect ->" + err.Error()
	}
	if err = connection.Login("anonymous", "anonymous"); err != nil {
		return "Can't login ->" + err.Error()
	}
	code, response := connection.RawActiveCmd("LIST .")
	if code < 200 {
		return fmt.Sprintf("Can't list -> %d", code)
	}
	fmt.Println(response)
	connection.Close()
	return ""

}

func standard(host string) (msg string) {
	var err error
	var connection *FTP

	if connection, err = Connect(host); err != nil {
		return "Can't connect ->" + err.Error()
	}
	if err = connection.Login("anonymous", "anonymous"); err != nil {
		return "Can't login ->" + err.Error()
	}
	if _, err = connection.List(""); err != nil {
		return "Can't list ->" + err.Error()
	}
	connection.Close()
	return ""
}

func walk(host string) (msg string) {
	var err error
	var connection *FTP
	deep := 2

	if connection, err = Connect(host); err != nil {
		return "Can't connect ->" + err.Error()
	}
	if err = connection.Login("anonymous", "anonymous"); err != nil {
		return "Can't login ->" + err.Error()
	}

	err = connection.Walk("/", func(path string, info os.FileMode, err error) error {
		fmt.Println(path)
		return nil

	}, deep)
	if err != nil {
		return "Can't walk ->" + err.Error()
	}
	connection.Close()
	return ""

}

func TestLogin_good(t *testing.T) {
	str := standard(goodServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestLogin_bad(t *testing.T) {
	str := standard(badServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestLogin_ugly(t *testing.T) {
	str := standard(uglyServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

//

func TestWalk_good(t *testing.T) {
	str := walk(goodServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestWalk_bad(t *testing.T) {
	str := walk(badServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestWalk_ugly(t *testing.T) {
	str := walk(uglyServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestActiveCommand(t *testing.T) {
	str := active(uglyServer)
	if len(str) > 0 {
		t.Error(str)
	}
}
