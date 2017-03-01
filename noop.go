package goftp

// Noop will send a NOOP (no operation) to the server
func (ftp *FTP) Noop() (err error) {
	_, err = ftp.cmd(StatusOK, "NOOP")
	return
}
