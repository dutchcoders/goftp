package goftp

import (
	"errors"
	"strings"
)

// Stat gets the status of path from the remote host
func (ftp *FTP) Stat(path string) ([]string, error) {
	if err := ftp.send("STAT %s", path); err != nil {
		return nil, err
	}

	stat, err := ftp.receive()
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(stat, StatusFileStatus) &&
		!strings.HasPrefix(stat, StatusDirectoryStatus) &&
		!strings.HasPrefix(stat, StatusSystemStatus) {
		return nil, errors.New(stat)
	}
	if strings.HasPrefix(stat, StatusSystemStatus) {
		return strings.Split(stat, "\n"), nil
	}
	lines := []string{}
	for _, line := range strings.Split(stat, "\n") {
		if strings.HasPrefix(line, StatusFileStatus) {
			continue
		}
		//fmt.Printf("%v\n", re.FindAllStringSubmatch(line, -1))
		lines = append(lines, strings.TrimSpace(line))

	}
	// TODO(vbatts) parse this line for SystemTypeWindowsNT
	//"213-status of /remfdata/all.zip:\r\n    09-12-15  04:07AM             37192705 all.zip\r\n213 End of status.\r\n"

	// and this for SystemTypeUnixL8
	// "-rw-r--r--   22 4015     4015        17976 Jun 10  1994 COPYING"
	// "drwxr-xr-x    6 4015     4015         4096 Aug 21 17:25 kernels"
	return lines, nil
}
