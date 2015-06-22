// ftpTree
package main

import (
	"fmt"
	"github.com/VincenzoLaSpesa/goftp"
	"os"
	"strings"
	"time"
)

var t rune
var b rune
var l rune
var lastdeep int

func main() {
	t = '├'
	b = '─'
	l = '└'
	lastdeep = -1
	fmt.Println(walk("bo.mirror.garr.it:21"))
}

func walk(host string) (msg string) {

	var err error
	var connection *goftp.FTP
	deep := 5

	if connection, err = goftp.Connect(host); err != nil {
		return "Can't connect ->" + err.Error()
	}
	if err = connection.Login("anonymous", "anonymous"); err != nil {
		return "Can't login ->" + err.Error()
	}

	fmt.Println(host)

	err = connection.Walk("/", func(path string, info os.FileMode, err error) error {

		I := strings.Count(path, "/") - 1
		lindex := strings.LastIndex(path, "/")

		if lastdeep != I { //abbiamo cambiato directory
			for i := 1; i < I; i++ {
				fmt.Print("    ")
			}
			fmt.Print("├───")
			fmt.Println(path[:lindex])
		}

		for i := 0; i < I; i++ {
			fmt.Print("    ")
		}
		fmt.Print("├───")
		nomefile := path[1+lindex:]
		fmt.Println(nomefile)
		time.Sleep(200 * time.Millisecond)
		lastdeep = I
		return nil

	}, deep)
	if err != nil {
		return "Can't walk ->" + err.Error()
	}
	connection.Close()
	return ""
}
