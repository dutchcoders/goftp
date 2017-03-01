package goftp

// Rmd remove directory
func (ftp *FTP) Rmd(path string) (err error) {
	_, err = ftp.cmd(StatusActionOK, "RMD %s", path)
	return
}
