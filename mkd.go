package goftp

// Mkd makes a directory on the remote host
func (ftp *FTP) Mkd(path string) error {
	_, err := ftp.cmd(StatusPathCreated, "MKD %s", path)
	return err
}
