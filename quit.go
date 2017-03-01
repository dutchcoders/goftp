package goftp

// Quit sends quit to the server and close the connection. No need to Close after this.
func (ftp *FTP) Quit() (err error) {
	if _, err := ftp.cmd(StatusConnectionClosing, "QUIT"); err != nil {
		return err
	}

	ftp.conn.Close()
	ftp.conn = nil

	return nil
}
