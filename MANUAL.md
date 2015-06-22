# goftp documentation #

**Table of Contents**  (*generated with [DocToc](http://doctoc.herokuapp.com/)*)

- [Public functions](#)
		- [type FTP](#)
		- [func  Connect](#)
		- [func  ConnectDbg](#)
		- [func (*FTP) AuthTLS](#)
		- [func (*FTP) Close](#)
		- [func (*FTP) Cwd](#)
		- [func (*FTP) Dele](#)
		- [func (*FTP) List](#)
		- [func (*FTP) Login](#)
		- [func (*FTP) Mkd](#)
		- [func (*FTP) Noop](#)
		- [func (*FTP) Pasv](#)
		- [func (*FTP) Pwd](#)
		- [func (*FTP) Quit](#)
		- [func (*FTP) RawCmd](#)
		- [func (*FTP) ReadAndDiscard](#)
		- [func (*FTP) Rename](#)
		- [func (*FTP) Retr](#)
		- [func (*FTP) Stor](#)
		- [func (*FTP) Type](#)
		- [func (*FTP) Walk](#)
		- [type RetrFunc](#)
		- [type WalkFunc](#)
- [Notes on the different list commands for ftp](#)
	- [How do I know which commands the target server supports?](#)
	- [MLSD](#)
	- [NLST](#)
	- [EPLF](#)
	- [LIST](#)
- [Sample Code](#)

## Public functions ##

#### type FTP

```go
type FTP struct {
}
```


#### func  Connect

```go
func Connect(addr string) (*FTP, error)
```

connect to server, debug is OFF

#### func  ConnectDbg

```go
func ConnectDbg(addr string) (*FTP, error)
```

connect to server, debug is ON

#### func (*FTP) AuthTLS

```go
func (ftp *FTP) AuthTLS(config tls.Config) error
```

secures the ftp connection by using TLS

#### func (*FTP) Close

```go
func (ftp *FTP) Close()
```

#### func (*FTP) Cwd

```go
func (ftp *FTP) Cwd(path string) (err error)
```

change current path

#### func (*FTP) Dele

```go
func (ftp *FTP) Dele(path string) (err error)
```

delete file

#### func (*FTP) List

```go
func (ftp *FTP) List(path string) (files []string, err error)
```

list the path (or current directory)

#### func (*FTP) Login

```go
func (ftp *FTP) Login(username string, password string) (err error)
```

login to the server

#### func (*FTP) Mkd

```go
func (ftp *FTP) Mkd(path string) error
```

make directory

#### func (*FTP) Noop

```go
func (ftp *FTP) Noop() (err error)
```

will send a NOOP (no operation) to the server

#### func (*FTP) Pasv

```go
func (ftp *FTP) Pasv() (port int, err error)
```

enables passive data connection and returns port number

#### func (*FTP) Pwd

```go
func (ftp *FTP) Pwd() (path string, err error)
```

get current path

#### func (*FTP) Quit

```go
func (ftp *FTP) Quit() (err error)
```

send quit to the server and close the connection

#### func (*FTP) RawCmd

```go
func (ftp *FTP) RawCmd(command string, args ...interface{}) (code int, line string)
```

Send raw commands, return response as string and response code as int

#### func (*FTP) ReadAndDiscard

```go
func (ftp *FTP) ReadAndDiscard() (int, error)
```

read all the buffered bytes and return

#### func (*FTP) Rename

```go
func (ftp *FTP) Rename(from string, to string) (err error)
```

rename file

#### func (*FTP) Retr

```go
func (ftp *FTP) Retr(path string, retrFn RetrFunc) (s string, err error)
```

retrieves file

#### func (*FTP) Stor

```go
func (ftp *FTP) Stor(path string, r io.Reader) (err error)
```

upload file

#### func (*FTP) Type

```go
func (ftp *FTP) Type(t string) error
```

change transfer type

#### func (*FTP) Walk

```go
func (ftp *FTP) Walk(path string, walkFn WalkFunc) (err error)
```

walks recursively through path and call walkfunc for each file

#### type RetrFunc

```go
type RetrFunc func(r io.Reader) error
```


#### type WalkFunc

```go
type WalkFunc func(path string, info os.FileMode, err error) error
```

---

# Appendix #

This section contains information about the state of the lib development and the ftp protocol.
It will provide a starting point for extending the lib with new features.

## Notes on the different list commands for ftp ##

This is a small review of how FTP listing commands actually work in practice.
A more detailed description of the FTP protocols can be found in the following links:


1. [http://cr.yp.to/ftp.html](http://cr.yp.to/ftp.html "http://cr.yp.to/ftp.html") (D. J. Bernstein's FTP Reference )
2. [https://tools.ietf.org/pdf/rfc959.pdf](https://tools.ietf.org/pdf/rfc959.pdf "https://tools.ietf.org/pdf/rfc959.pdf") (RFC959)
3. [http://www.nsftools.com/tips/RawFTP.htm](http://www.nsftools.com/tips/RawFTP.htm) (NSF Tools - List of raw FTP commands)

### How do I know which commands the target server supports? ###

The first thing to do is trying the FEAT ftp command.

[https://tools.ietf.org/html/rfc2389#section-3.2](https://tools.ietf.org/html/rfc2389#section-3.2 "https://tools.ietf.org/html/rfc2389#section-3.2")

While not absolutely necessary, FEAT will help reduce unnecessary traffic between the user-PI and server PI as more extensions may be introduced in the future.  

If no mechanism existed for this, a user-FTP process would have to try each
extension in turn resulting in a series of exchanges between the user-PI and server-PI.  Apart from being possibly wasteful, this procedure may not always be possible, as issuing of a command just to determine if it is supported or not may have some effect that is not desired.

FEAT it will output like this:

    211-Extensions supported:
      MLST size*;create;modify*;perm;media-type
      SIZE
      COMPRESSION
      MDTM
    211 END


### MLSD ###
**This format is supported in goftp**

It's the preferred way to obtain directory listing in machine readable format.
It'll output one line per file, formatted like this:
( this is a ProFTPD response )

	modify=20081125111318;perm=adfr;size=24;type=OS.unix=symlink;unique=FE0BU200059DA;UNIX.group=0;UNIX.mode=0777;UNIX.owner=0; linux

or like this

	size=0;type=dir;perm=el;modify=20020409191530; bin
	size=3919312;type=file;perm=r;modify=20000310140400; bar.txt
	size=6686176;type=file;perm=r;modify=20001215181000; baz.txt
	size=3820092;type=file;perm=r;modify=20000310140300; foo.txt
	size=27439;type=file;perm=r;modify=20020923151312; foo.zip

The order of fields is not fixed, but it's very easy to parse each line to a dictionary

### NLST ###

This format is **NOT yet** supported in goftp

Returns a list of filenames in the given directory (defaulting to the current directory), with no other information. 
The difference between LIST and NLST is that NLST returns a compressed form of the directory, showing only the name of each file, while LIST returns the entire directory.

The NLST format consists of a sequence of abbreviated pathnames. Each pathname is terminated by \015\012, without regard to the current binary flag. If an abbreviated pathname starts with a slash, it represents the pathname obtained by replacing each \000 by \012. If an abbreviated pathname does not start with a slash, it represents the pathname obtained by concatenating

1.     the pathname of the directory;
1.     a slash, if the pathname of the directory does not end with a slash; and
1.     the abbreviated pathname, with each \000 replaced by \012. 

For example, if a directory /pub produces 

	foo\015\012bar\015\012 

under NLST, it refers to the pathnames /pub/foo and /pub/bar. 

### EPLF ###

This format is **NOT yet** supported in goftp

EPLF stands for "Easily Parsed LIST Format", it was designed in 1996 by D. J. Bernstein

EPLF was designed to

1.     reliably communicate the information needed by clients;
2.     make client and server implementation as easy as possible; and
3.     be readable to humans, when readability does not complicate implementations. 

Output will be formatted like this:

     +i8388621.48594,m825718503,r,s280, djb.html
     +i8388621.50690,m824255907,/, 514
     +i8388621.48598,m824253270,r,s612, 514.html
 

An EPLF response to LIST is a series of lines, each line specifying one file. Each line contains

1.     a plus sign (\053);
2.     a series of facts about the file;
3.     a tab (\011);
4.     an abbreviated pathname; and
5.     \015\012. 
(The terminating \015\012 does not depend on the binary flag.)

 Each fact is zero or more bytes of information, terminated by a comma and not containing any tabs. Facts may appear in any order. Each fact appears at most once.

Facts have the general format xy, where x is one of the following strings:

1.     r: If this file's pathname is supplied as a RETR parameter, the RETR may succeed. The server is required to use an empty y. The server must supply this fact unless it is confident that (because of file type problems, permission problems, etc.) there's no point in the client sending RETR. Mirroring clients can save time by issuing RETR requests only for files where this fact is supplied. The presence of r does not guarantee that RETR will succeed: for example, the file may be removed or renamed, or the RETR may suffer a temporary failure.
1.     /: If this file's pathname is supplied as a CWD parameter, the CWD may succeed. The server is required to use an empty y. As with r, the server must supply this fact unless it is confident that there's no point in the client sending CWD. Indexing clients can save time by issuing CWD requests only for files where this fact is supplied. The presence of / does not guarantee that CWD will succeed.
1.     s: The size of this file is y. The server is required to provide a sequence of one or more ASCII digits in y, specifying a number. If the file is retrieved as a binary file and is not modified, it will contain exactly y bytes. This fact is optional; it should not be supplied for files that can never be retrieved, or for files whose size is constantly changing. Clients can use this fact to preallocate space.
1.     m: This file was last modified at y. The server is required to provide a sequence of one or more ASCII digits in y, specifying a number of seconds, real time, since the UNIX epoch at the beginning of 1970 GMT. This fact is optional; it should not be supplied by servers that do not know the time in GMT, and it should not be supplied for files that have been modified more recently than one minute ago. (It also cannot be supplied for files last modified before 1970.) Mirroring clients can save time by skipping files whose modification time has not changed since the previous mirror.
1.     i: This file has identifier y. If two files on the same FTP server (not necessarily in the same LIST response) have the same identifiers then they have the same contents: a RETR of each file will produce the same results, for example, and a CWD to each file will produce the same results in a subsequent RETR or LIST. (Under UNIX, for example, the server could use dev.ino as an identifier, where dev and ino are the device number and inode number of the file as returned by stat(). Note that lstat() is not a good idea for FTP directory listings.) Indexing clients can use this fact to avoid searching the same directory twice; mirroring clients can use this fact to avoid retrieving the same file twice. This fact is optional, but high-quality servers will always supply it at least for directories so that indexing programs can avoid CWD loops.
1.     up: The client may use SITE CHMOD to change the UNIX permission bits of this file. The server must provide three ASCII digits in y, in octal, showing the current permission bits. 


Modification times are expressed as second counters rather than calendar dates and times, for example, because second counters are much easier to generate and parse, making it more likely that browsers will display times in the viewer's time zone and native language. 


References: [http://cr.yp.to/ftp/list/eplf.html](http://cr.yp.to/ftp/list/eplf.html)


### LIST ###

__Only__ standard Unix LS is supported in goftp 

List is the original way to do listing on ftp.
It's intended for human readable format.
The output format is not standard.
It should be similar to something like this:

	-rw-r--r--   1 aokur    (?)            17 Jan  5  2010 fcheck.js

or this

	-rw-r--r--   1 0        0            2407 Jun 12 04:01 dim
	-rw-r--r--   1 0        0             522 Jun 12 05:30 enti-garr-v6.txt
	-rw-r--r--   1 0        0        1468105470 Jun 12 04:01 ls-lR
	-rw-r--r--   1 0        0        245428953 Jun 12 04:01 ls-lR.gz
	-rw-r--r--   1 0        0        15775864 Jun 12 04:04 ls-lR.patch.gz
	-rw-r--r--   1 0        0              22 Jun 12 04:04 ls-lR.times
	lrwxrwxrwx   1 0        0              11 Nov 25  2008 mirrors -> pub/mirrors
	drwxr-xr-x   7 0        0            4096 May 20  2013 pub
	-rw-r--r--   1 0        0            8623 Jun 12 05:40 six2four-garr.txt
	-rw-r--r--   1 0        0             460 Jun 12 05:35 teredo-garr.txt
	-rw-r--r--   1 0        0              12 May 19 11:38 timezone
	-rw-r--r--   1 0        0              61 Jun 12 05:00 v4v6
	-rw-r--r--   1 0        0             166 Jul 24  2008 welcome.msg

or this

	lrwxrwxrwx    1 0        0              19 Apr 11  2009 debian -> ./pub/debian/debian
	lrwxrwxrwx    1 0        0              20 Apr 11  2009 debian-cd -> ./pub/debian-cdimage
	lrwxrwxrwx    1 0        0              20 Apr 11  2009 debian-cdimage -> ./pub/debian-cdimage
	drwxr-xr-x    6 0        0            4096 Jun 08 17:09 pub
	-rw-r--r--    1 0        0             819 Feb 03  2009 welcome.msg

The LIST format varies widely from server to server. The most common format is /bin/ls format, which is difficult to parse with even moderate reliability. This poses a serious problem for clients that need more information than names.

A standard response in /bin/ls format will line contains


1. - for a regular file or d for a directory;
2. the literal string rw-r--r-- 1 owner group for a regular file, or rwxr-xr-x 1 owner group for a directory;
3. the file size in decimal right-justified in a 13-byte field;
4. a three-letter month name, first letter capitalized;
5. a day number right-justified in a 3-byte field;
6. a space and a 2-digit hour number;
7. a colon and a 2-digit minute number;
8. a space and the abbreviated pathname of the file. 

So a regular expression for parsing this format could be:

	([-ld])([-rwx]+)\s{2,}\d+\s(\d|\w+)\s{2,}(\(\?\)|\d+|\w+)\s{2,}(\d+)\s*(\w+\s*\d+\s*\d+\:*\d*)\s(\S+)

that, applied on the sample above, will extract tokens like this:

	entry 0:lrwxrwxrwx    1 0        0              19 Apr 11  2009 debian
	entry 1:l
	entry 2:rwxrwxrwx
	entry 3:0
	entry 4:0
	entry 5:19
	entry 6:Apr 11  2009
	entry 7:debian
.

	entry 0:lrwxrwxrwx    1 0        0              20 Apr 11  2009 debian-cd
	entry 1:l
	entry 2:rwxrwxrwx
	entry 3:0
	entry 4:0
	entry 5:20
	entry 6:Apr 11  2009
	entry 7:debian-cd
.
	 
	entry 0:lrwxrwxrwx    1 0        0              20 Apr 11  2009 debian-cdimage
	entry 1:l
	entry 2:rwxrwxrwx
	entry 3:0
	entry 4:0
	entry 5:20
	entry 6:Apr 11  2009
	entry 7:debian-cdimage
.
	 
	entry 0:drwxr-xr-x    6 0        0            4096 Jun 08 17:09 pub
	entry 1:d
	entry 2:rwxr-xr-x
	entry 3:0
	entry 4:0
	entry 5:4096
	entry 6:Jun 08 17:09
	entry 7:pub
.
	 
	entry 0:-rw-r--r--    1 0        0             819 Feb 03  2009 welcome.msg
	entry 1:-
	entry 2:rw-r--r--
	entry 3:0
	entry 4:0
	entry 5:819
	entry 6:Feb 03  2009
	entry 7:welcome.msg

goftp will only recognize output that match this format.


References: [http://cr.yp.to/ftp/list/binls.html](http://cr.yp.to/ftp/list/binls.html)

---

## Sample Code

```go
package main
		
		import (
		    "github.com/dutchcoders/goftp"
		    "crypto/tls"
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
