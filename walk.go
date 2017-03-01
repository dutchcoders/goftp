package goftp

import (
	"log"
	"os"
)

// Walk walks recursively through path and call walkfunc for each file
func (ftp *FTP) Walk(path string, walkFn WalkFunc) (err error) {
	/*
		if err = walkFn(path, os.ModeDir, nil); err != nil {
			if err == filepath.SkipDir {
				return nil
			}
		}
	*/
	if ftp.debug {
		log.Printf("Walking: '%s'\n", path)
	}

	var lines []string

	if lines, err = ftp.List(path); err != nil {
		return
	}

	for _, line := range lines {
		_, t, subpath := parseLine(line)

		switch t {
		case "dir":
			if subpath == "." {
			} else if subpath == ".." {
			} else {
				if err = ftp.Walk(path+subpath+"/", walkFn); err != nil {
					return
				}
			}
		case "file":
			if err = walkFn(path+subpath, os.FileMode(0), nil); err != nil {
				return
			}
		}
	}

	return
}
