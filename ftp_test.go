package goftp

import "testing"

import "fmt"

var goodServer string
var badServer string

func init() {
	goodServer = "bo.mirror.garr.it:21"
	//badServer = "ftp.packardbell.com:21"
	badServer = "ftp.musicbrainz.org:21"
}

func standard(host string) (msg string) {
	var err error
	var connection *FTP

	if connection, err = Connect(host); err != nil {
		return "Can't connect"
	}
	if err = connection.Login("anonymous", "anonymous"); err != nil {
		return "Can't login"
	}
	connection.ReadAndDiscard()
	if _, err = connection.List(""); err != nil {
		return "Can't list"
	}
	connection.Close()
	return ""
}

func smart(host string) (msg string) {
	var err error
	var connection *FTP

	if connection, err = Connect(host); err != nil {
		return "Can't connect"
	}
	if err = connection.SmartLogin("anonymous", "anonymous"); err != nil {
		return "Can't login"
	}
	connection.ReadAndDiscard()
	if _, err = connection.List(""); err != nil {
		return "Can't list"
	}
	connection.Close()
	return ""
}

func TestLogin00(t *testing.T) {
	fmt.Printf("---	Standard login on good server\n")
	str := standard(goodServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestLogin01(t *testing.T) {
	fmt.Printf("---	Standard login on bad server\n")
	str := standard(badServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestLogin02(t *testing.T) {
	fmt.Printf("---	Smart login on good server\n")
	str := smart(goodServer)
	if len(str) > 0 {
		t.Error(str)
	}
}

func TestLogin03(t *testing.T) {
	fmt.Printf("---	Smart login on bad server\n")
	str := smart(badServer)
	if len(str) > 0 {
		t.Error(str)
	}

}
