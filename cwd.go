package goftp

// Cwd changes current working directory on remote host to path
func (ftp *FTP) Cwd(path string) (err error) {
	_, err = ftp.cmd(StatusActionOK, "CWD %s", path)
	return
}
