package goftp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (ftp *FTP) getServerFeatures() (features uint32) {
	features = 0
	code, str := ftp.RawCmd("FEAT")
	//fmt.Println(str)
	if code < 0 || code > 299 {
		return BAD
	}

	lines := strings.Split(str, "\n")

	for _, f := range lines {
		if (features&MLSD == 0) && strings.Contains(f, "MLST") { // not a bug.
			features = features | MLSD
		} else if (features&NLST == 0) && strings.Contains(f, "NLST") {
			features = features | NLST
		} else if (features&EPLF == 0) && strings.Contains(f, "EPLF") {
			features = features | EPLF
		}
	}

	return features
}

// list the path (or current directory) and parse it. Return an array with the files and one with the directories
func (ftp *FTP) GetFilesList(path string) (files []string, directories []string, err error) {
	if ftp.supportedfeatures == 0 {
		ftp.supportedfeatures = ftp.getServerFeatures()
	}
	if ftp.supportedfeatures&MLSD > 0 {
		fmt.Println("Using MLSD")
		code, response := ftp.RawPassiveCmd("MLSD " + path)
		if code < 0 || code > 299 {
			return nil, nil, errors.New("MLSD did not work")
		}
		return ftp.parseMLSD(response, path)
	} else if ftp.supportedfeatures&EPLF > 0 {
		fmt.Println("Using EPLF")
		code, response := ftp.RawPassiveCmd("EPLF " + path)
		if code < 0 || code > 299 {
			return nil, nil, errors.New("EPLF did not work")
		}
		return ftp.parseEPLF(response, path)
	} else if ftp.supportedfeatures&NLST > 0 {
		fmt.Println("Using NLST")
		code, response := ftp.RawPassiveCmd("NLST " + path)
		if code < 0 || code > 299 {
			return nil, nil, errors.New("NLST did not work")
		}
		return ftp.parseNLST(response, path)
	} else {
		fmt.Println("Using LIST")
		code, response := ftp.RawPassiveCmd("LIST " + path)
		if code < 0 || code > 299 {
			return nil, nil, errors.New("LIST did not work")
		}
		return ftp.parseUnixLIST(response, path)
	}
}

func (ftp *FTP) parseMLSD(data []string, basePath string) (files []string, directories []string, err error) {
	var i:=0
	for _, line := range data {
		_, t, subpath := parseLine_MLST(line)

		switch t {
		case "dir":
			if subpath == "." {
				i++
			} else if subpath == ".." {
				i++
			} else {
				directories = append(directories, basePath+subpath+"/")
			}
		case "file":
			files = append(files, basePath+subpath)
		}
	}
	if i=0 { // no "." and no ".." ? This should not happen
		errors.New("The path seems not valid")
	}
	return files, directories, err
}

//

func (ftp *FTP) parseEPLF(data []string, basePath string) (files []string, directories []string, err error) {
	fmt.Errorf("Not implemented!\n")
	return nil, nil, errors.New("Not implemented!")
}

func (ftp *FTP) parseNLST(data []string, basePath string) (files []string, directories []string, err error) {
	fmt.Errorf("Not implemented!\n")
	return nil, nil, errors.New("Not implemented!")
}

func (ftp *FTP) parseUnixLIST(data []string, basePath string) (files []string, directories []string, err error) {
	/*Stolen straight from the ASF's commons Java FTP LIST parser library.
	  http://svn.apache.org/repos/asf/commons/proper/net/trunk/src/java/org/apache/commons/net/ftp/


	  REGEXP = %r{
	    ([pbcdlfmpSs-])
	    (((r|-)(w|-)([xsStTL-]))((r|-)(w|-)([xsStTL-]))((r|-)(w|-)([xsStTL-])))\+?\s+
	    (?:(\d+)\s+)?
	    (\S+)\s+
	    (?:(\S+(?:\s\S+)*)\s+)?
	    (?:\d+,\s+)?
	    (\d+)\s+
	    ((?:\d+[-/]\d+[-/]\d+)|(?:\S+\s+\S+))\s+
	    (\d+(?::\d+)?)\s+
	    (\S*)(\s*.*)
	  }x
	*/

	return nil, nil, errors.New("Not implemented!")
}

//
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

func parseLine_MLST(line string) (perm string, t string, filename string) {
	for _, v := range strings.Split(line, ";") {
		v2 := strings.Split(v, "=")

		switch v2[0] {
		case "perm":
			perm = v2[1]
		case "type":
			t = v2[1]
		default:
			filename = v[1 : len(v)-2]
		}
	}

	return
}
