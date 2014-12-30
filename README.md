goftp
=====

Golang FTP library with Walk support.


## Sample
```
    var err error
    var ftp *FTP

    if ftp, err = ftp.ConnectFTP("ftp.server.com"); err != nil {
        panic(err)
    }

    defer ftp.Close()

    if err = ftp.Login("username", "password"); err != nil {
        panic(err)
    }

    if err = ftp.CWD("/"); err != nil {
        panic(err)
    }

    var files []string
    if files, err = ftp.List(""); err != nil {
        panic(err)
    }

    fmt.Println(files)

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
```

