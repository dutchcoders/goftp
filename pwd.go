package goftp

// Pwd gets current path on the remote host
func (ftp *FTP) Pwd() (path string, err error) {
	var line string
	if line, err = ftp.cmd(StatusPathCreated, "PWD"); err != nil {
		return
	}

	res := RePwdPath.FindAllStringSubmatch(line[4:], -1)

	path = res[0][1]
	return
}
