goftp
=====

Golang FTP library with Walk support.

## Features

* AUTH TLS support
* Walk 

## Sample
		package main
		
		import (
		    "crypto/sha256"
		    "crypto/tls"
		    "fmt"
		    "io"
		    "os"

		    "github.com/dutchcoders/goftp"
		)
		
		func main() {
		    var err error
		    var ftp *goftp.FTP
		
		    // For debug messages: goftp.ConnectDbg("ftp.server.com:21")
		    if ftp, err = goftp.Connect("ftp.server.com:21"); err != nil {
		        panic(err)
		    }
		
		    defer ftp.Close()
		
		    config := tls.Config{
		        InsecureSkipVerify: true,
		        ClientAuth:         tls.RequestClientCert,
		    }
		
		    if err = ftp.AuthTLS(config); err != nil {
		        panic(err)
		    }
		
		    if err = ftp.Login("username", "password"); err != nil {
		        panic(err)
		    }
		
		    if err = ftp.Cwd("/"); err != nil {
		        panic(err)
		    }
		
		    var curpath string
		    if curpath, err = ftp.Pwd(); err != nil {
		        panic(err)
		    }
		
		    fmt.Printf("Current path: %s", curpath)
		
		    var files []string
		    if files, err = ftp.List(""); err != nil {
		        panic(err)
		    }
		
		    fmt.Println(files)
		
		    var file *os.File
		    if file, err = os.Open("/tmp/test.txt"); err != nil {
		        panic(err)
		    }
		
		    if err := ftp.Stor("/test.txt", file); err != nil {
		        panic(err)
		    }
		
		    err = ftp.Walk("/", func(path string, info os.FileMode, err error) error {
		        _, err = ftp.Retr(path, func(r io.Reader) error {
		            var hasher = sha256.New()
		            if _, err = io.Copy(hasher, r); err != nil {
		                return err
		            }
		
		            hash := fmt.Sprintf("%s %x", path, sha256.Sum256(nil))
		            fmt.Println(hash)
		
		            return err
		        })
		
		        return nil
		    })
		}

## Contributions

Contributions are welcome.

* Sourav Datta: for his work on the anonymous user login and multiline return status.
* Vincenzo La Spesa: for his work on resolving login issues with specific ftp servers


## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

## Copyright and license

Code and documentation copyright 2011-2014 Remco Verhoef.
Code released under [the MIT license](LICENSE).

