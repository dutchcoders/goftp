package goftp

// ReadAndDiscard reads all the buffered bytes and returns the number of bytes
// that cleared from the buffer
func (ftp *FTP) ReadAndDiscard() (int, error) {
	var i int
	bufferSize := ftp.reader.Buffered()
	for i = 0; i < bufferSize; i++ {
		if _, err := ftp.reader.ReadByte(); err != nil {
			return i, err
		}
	}
	return i, nil
}
