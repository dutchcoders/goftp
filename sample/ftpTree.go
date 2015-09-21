// ftpTree

/*
This sample shows a complete application of goftp.
It walks trough an ftp server and print the folder structure like
tree command of unix ( and msdos )
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/VincenzoLaSpesa/goftp"
)

var t rune
var b rune
var l rune
var lastdeep int
var lastDir string

//const defaultServer = "bo.mirror.garr.it:21"

func main() {
	t = '├'
	b = '─'
	l = '└'
	lastdeep = -1
	var server string

	if len(os.Args) < 2 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Insert a valid ftp url, like an.ftp.server:21")
		server, _ = reader.ReadString('\n')
		server = strings.TrimSpace(server)
		fmt.Println("Connecting to <", server, ">")
	} else {
		server = os.Args[1]
	}

	fmt.Println(walk(server))
}

func walk(host string) (msg string) {

	var err error
	var connection *goftp.FTP
	deep := 4

	if connection, err = goftp.Connect(host); err != nil {
		return "Can't connect ->" + err.Error()
	}
	if err = connection.Login("anonymous", "anonymous"); err != nil {
		return "Can't login ->" + err.Error()
	}

	fmt.Println(host)

	err = connection.WalkCustom("/", func(path string, info os.FileMode, err error) error {

		I := strings.Count(path, "/") - 1
		lindex := strings.LastIndex(path, "/")
		currentDir := path[:lindex]

		if lastdeep != I || lastDir != currentDir { //change of dir
			for i := 1; i < I; i++ {
				fmt.Print("|   ")
			}
			fmt.Print("├───")
			fmt.Println(currentDir)
		}

		for i := 0; i < I; i++ {
			fmt.Print("|   ")
		}
		fmt.Print("├───")
		nomefile := path[1+lindex:]
		fmt.Println(nomefile)
		// I don't wanna flood
		time.Sleep(200 * time.Millisecond)
		lastdeep = I
		lastDir = currentDir
		return nil

	},
		func(pwd string, errorCode int, errorStr string, shouldBeSkippable bool) (skippable bool, err error) {
			if errorCode == 550 {
				fmt.Println("Skipping <", pwd, ">")
				return true, nil
			} else {

				fmt.Println("Error on <", pwd, "> of type <", errorStr, ">. giving up.")
				return false, nil
			}
		},
		deep)
	if err != nil {
		fmt.Println("Error on ", err.Error())
		return "Can't walk ->" + err.Error()
	}
	connection.Close()
	return ""
}
