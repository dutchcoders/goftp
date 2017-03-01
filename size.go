package goftp

import (
	"strconv"
)

// Size returns the size of a file.
func (ftp *FTP) Size(path string) (size int, err error) {
	line, err := ftp.cmd("213", "SIZE %s", path)

	if err != nil {
		return 0, err
	}

	return strconv.Atoi(line[4 : len(line)-2])
}
