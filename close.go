package goftp

// Close ends the FTP connection
func (ftp *FTP) Close() error {
	return ftp.conn.Close()
}
