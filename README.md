goftp
=====

Golang FTP library with Walk support.


## Features

* AUTH TLS support
* Walk 

## Sample
```
package main

import (
    "github.com/dutchcoders/goftp"
    "crypto/tls"
)

func main() {
    var err error
    var ftp *goftp.FTP

    if ftp, err = goftp.Connect("ftp.server.com:21", true); err != nil {
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
    if curpath, err = ftp.Pwd("/"); err != nil {
        panic(err)
    }

    fmt.Printf("Current path: %s", curpath)

    var files []string
    if files, err = ftp.List(""); err != nil {
        panic(err)
    }

    fmt.Println(files)

    if file, err := os.Open("/tmp/test.txt"); err!=nil {
        panic(err)
    }

    if err := ftp.Stor("/test.txt", file); err!=nil {
        panic(err)
    }

    err = ftp.Walk("/", func(path string, info os.FileMode, err error) error {
        w := &bytes.Buffer{}

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
```

## Contributions

Contributions are welcome.

## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

## Copyright and license

Code and documentation copyright 2011-2014 Remco Verhoef.
Code released under [the MIT license](LICENSE).

