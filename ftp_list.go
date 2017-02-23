package goftp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"regexp"
	"strings"
)

//getServerFeatures use FEAT to retrieve server features.
func (ftp *FTP) getServerFeatures() (features uint32) {
	features = 0
	code, str := ftp.RawCmd("FEAT")
	if code < 0 || code > 299 {
		return BAD
	}

	lines := strings.Split(str, "\n")

	for _, f := range lines {
		// Only MLSD/MLST is supported
		if (features&MLSD == 0) && strings.Contains(f, "MLST") { // not a bug, see command reference.
			features = features | MLSD

			// Check of NLST and EPLF are disabled since not yet supported

			/*} else if (features&NLST == 0) && strings.Contains(f, "NLST") {
				features = features | NLST
			} else if (features&EPLF == 0) && strings.Contains(f, "EPLF") {
				features = features | EPLF*/
		}
	}

	return features
}

//GetFilesList list the path (or current directory) and parse it.
//Return an array with the files, one with the directories and one with the links
func (ftp *FTP) GetFilesList(path string) (files []string, directories []string, links []string, err error) {
	if ftp.supportedfeatures == 0 {
		ftp.supportedfeatures = ftp.getServerFeatures()
	}
	if len(path) < 2 {
		path = "./"
	}
	if ftp.supportedfeatures&MLSD > 0 {
		//fmt.Println("Using MLSD")
		code, response := ftp.RawPassiveCmd("MLSD " + path)

		if code == 550 {
			return nil, nil, nil, errors.New("550 Requested action not taken. File unavailable")
		}
		if code < 0 || code > 299 {
			return nil, nil, nil, fmt.Errorf("%v MLSD did not work ", code)
		}
		return ftp.parseMLSD(response, path)
	} else if ftp.supportedfeatures&EPLF > 0 {
		//fmt.Println("Using EPLF")
		code, response := ftp.RawPassiveCmd("EPLF " + path)

		if code == 550 {
			return nil, nil, nil, errors.New("550 Requested action not taken. File unavailable")
		}
		if code < 0 || code > 299 {
			return nil, nil, nil, fmt.Errorf("%v EPLF did not work ", code)
		}
		return ftp.parseEPLF(response, path)
	} else if ftp.supportedfeatures&NLST > 0 {
		//fmt.Println("Using NLST")
		code, response := ftp.RawPassiveCmd("NLST " + path)

		if code == 550 {
			return nil, nil, nil, errors.New("550 Requested action not taken. File unavailable")
		}
		if code < 0 || code > 299 {
			return nil, nil, nil, fmt.Errorf("%v NLST did not work ", code)
		}
		return ftp.parseNLST(response, path)
	}
	//fmt.Println("Using LIST")
	code, response := ftp.RawPassiveCmd("LIST " + path)
	//fmt.Println(response)
	if code == 550 {
		return nil, nil, nil, errors.New("550 Requested action not taken. File unavailable")
	}
	if code < 0 || code > 299 {
		return nil, nil, nil, errors.New("LIST did not work")
	}
	return ftp.parseUnixLIST(response, path)
}

//parseMLSD parse the response of a MLSD
//Return an array with the files, one with the directories and one with the links
func (ftp *FTP) parseMLSD(data []string, basePath string) (files []string, directories []string, links []string, err error) {
	for _, line := range data {
		_, t, subpath := parseLineMLST(line)

		switch t {
		case "dir":
			if subpath == "." {
			} else if subpath == ".." {
			} else {
				directories = append(directories, path.Join(basePath, subpath))
			}
		case "file":
			files = append(files, path.Join(basePath, subpath))
		case "OS.unix=symlink":
			links = append(links, path.Join(basePath, subpath))
		}

	}
	return files, directories, links, err
}

//parseLineMLST Parse a single MLST line
func parseLineMLST(line string) (perm string, t string, filename string) {
	for _, v := range strings.Split(line, ";") {
		v2 := strings.Split(v, "=")

		switch v2[0] {
		case "perm":
			perm = v2[1]
		case "type":
			t = v2[1]
			if t == "OS.unix" {
				t = t + "=" + v2[2]
				//fmt.Println("Found a link -> " + t)
			}
		default:
			filename = v[1 : len(v)-2]
		}
	}

	return
}

//parseEPLF parse the response of a EPLF
//Return an array with the files, one with the directories and one with the links
func (ftp *FTP) parseEPLF(data []string, basePath string) (files []string, directories []string, links []string, err error) {
	return nil, nil, nil, errors.New("Not implemented! (EPLF)")
}

//parseNLST parse the response of a NLST
//Return an array with the files, one with the directories and one with the links
func (ftp *FTP) parseNLST(data []string, basePath string) (files []string, directories []string, links []string, err error) {
	return nil, nil, nil, errors.New("Not implemented! (NLST)")
}

//parseUnixLIST parse the response of a LIST
//Return an array with the files, one with the directories and one with the links
func (ftp *FTP) parseUnixLIST(data []string, basePath string) (files []string, directories []string, links []string, err error) {
	var pattern = regexp.MustCompile(`([-ld])` + //dir,link flags -dbclps
		`([-rwxs]+)\s+` + //permissions
		`\d+\s+` + //items count
		`([\d\w\-]+)\s+` + //owner
		`(\(\?\)|[\d\w\-]+)\s+` + //group
		`(\d+)\s*` + //size
		`(\w+\s*\d+\s*\d+\:*\d*)\s+` + //date|time
		`(\S+)` /*name*/)
	for _, line := range data {
		match := pattern.FindStringSubmatch(line)
		/*for i, val := range match {
			fmt.Printf("entry %d:%s\n", i, val)
		}*/
		if len(match) > 6 {
			if match[1] == "-" { // a file
				files = append(files, path.Join(basePath, match[7]))
			} else if match[1] == "d" { // a directory
				directories = append(directories, path.Join(basePath, match[7]))
			} else if match[1] == "l" { // a link
				token := strings.Trim(strings.Split(line, "->")[1], " ")
				links = append(links, token)
			}
		}
	}

	sum := len(files) + len(directories) + len(links)
	if sum > 0 {
		return files, directories, links, nil
	}
	//empty folder
	if ftp.debug {
		log.Printf("Empty folder")
	}
	return nil, nil, nil, nil
}

//splitLines extract single lines from a buffered reader
func (ftp *FTP) splitLines(reader *bufio.Reader) (files []string, err error) {
	var line string
	for {
		line, err = reader.ReadString('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		files = append(files, string(line))
	}
	return files, nil
}
