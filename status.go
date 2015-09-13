package goftp

const (
	StatusOK                = "200"
	StatusFileStatus        = "213"
	StatusConnectionClosing = "221"
	StatusActionOK          = "250"
	StatusPathCreated       = "257"
	StatusActionPending     = "350"
)

var statusText = map[string]string{
	StatusOK:                "Command okay",
	StatusFileStatus:        "File status",
	StatusConnectionClosing: "Service closing control connection",
	StatusActionOK:          "Requested file action okay, completed",
	StatusPathCreated:       "Pathname Created",
	StatusActionPending:     "Requested file action pending further information",
}

// StatusText returns a text for the FTP status code. It returns the empty
// string if the code is unknown.
func StatusText(code string) string {
	return statusText[code]
}
