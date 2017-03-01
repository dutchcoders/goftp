package goftp

// Rename file on the remote host
func (ftp *FTP) Rename(from string, to string) (err error) {
	if _, err = ftp.cmd(StatusActionPending, "RNFR %s", from); err != nil {
		return
	}

	if _, err = ftp.cmd(StatusActionOK, "RNTO %s", to); err != nil {
		return
	}

	return
}
