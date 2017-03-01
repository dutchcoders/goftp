package goftp

// TypeCode for the representation types
type TypeCode string

const (
	// TypeASCII for ASCII
	TypeASCII = "A"
	// TypeEBCDIC for EBCDIC
	TypeEBCDIC = "E"
	// TypeImage for an Image
	TypeImage = "I"
	// TypeLocal for local byte size
	TypeLocal = "L"
)

// Type changes transfer type.
func (ftp *FTP) Type(t TypeCode) error {
	_, err := ftp.cmd(StatusOK, "TYPE %s", t)
	return err
}
