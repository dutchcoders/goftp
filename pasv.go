package goftp

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Pasv enables passive data connection and returns port number
func (ftp *FTP) Pasv() (port int, err error) {
	doneChan := make(chan int, 1)
	go func() {
		defer func() {
			doneChan <- 1
		}()
		var line string
		if line, err = ftp.cmd("227", "PASV"); err != nil {
			return
		}
		re := regexp.MustCompile(`\((.*)\)`)
		res := re.FindAllStringSubmatch(line, -1)
		if len(res) == 0 || len(res[0]) < 2 {
			err = errors.New("PasvBadAnswer")
			return
		}
		s := strings.Split(res[0][1], ",")
		if len(s) < 2 {
			err = errors.New("PasvBadAnswer")
			return
		}
		l1, _ := strconv.Atoi(s[len(s)-2])
		l2, _ := strconv.Atoi(s[len(s)-1])

		port = l1<<8 + l2

		return
	}()

	select {
	case _ = <-doneChan:

	case <-time.After(time.Second * 10):
		err = errors.New("PasvTimeout")
		ftp.Close()
	}

	return
}
