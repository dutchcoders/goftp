package goftp

import (
	"os"
	"time"
)

// FileInfo implements os.FileInfo to support file metadata like size and modification time to be passed
// to WalkFunc
type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	typ     string
	modTime time.Time
}

// Name returns the filename
func (f *FileInfo) Name() string {
	return f.name
}

// Size returns the size of the file in bytes
func (f *FileInfo) Size() int64 {
	return f.size
}

// Mode implements os.FileInfo, used to indicate if file is a directory or regular file
func (f *FileInfo) Mode() os.FileMode {
	return f.mode
}

// ModTime returns the last modification time provided by FTP server if supported
func (f *FileInfo) ModTime() time.Time {
	return f.modTime
}

// IsDir is a shortcut for Mode().IsDir()
func (f *FileInfo) IsDir() bool {
	return f.Mode().IsDir()
}

// Sys returns nil, not supported for FTP listings
func (f *FileInfo) Sys() interface{} {
	return nil
}
