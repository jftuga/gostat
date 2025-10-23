# gostat
display and set file timestamps

For more information on Birth Time, Modify Time, Access Time, and Create Time, please see the Wikipedia article on [MAC times](https://en.wikipedia.org/wiki/MAC_times).

Binaries for Windows, MacOS, and Linux are provided on the [Releases](https://github.com/jftuga/gostat/releases) page.

## Usage
```
Usage: gostat [OPTION]... [FILE]...
Display and set file time stamps

  -a string
    	set file access time, format: YYYYMMDD.HHMMSS
  -b string
    	set both access and modify time, format: YYYYMMDD.HHMMSS
  -m string
    	set file modify time, format: YYYYMMDD.HHMMSS
  -v	show program version and then exit
```

## Example - display times
```
PS C:\github.com\jftuga\gostat> .\gostat.exe go.*

name  : go.mod
size  : 84
btime : 2021-03-29 10:31:20.7843427 -0400 EDT
ctime : 2021-03-29 10:31:28.8303228 -0400 EDT
mtime : 2021-03-29 10:31:28.8303228 -0400 EDT
atime : 2021-03-29 10:31:28.8303228 -0400 EDT

name  : go.sum
size  : 171
btime : 2021-03-29 10:31:28.8347262 -0400 EDT
ctime : 2021-03-29 10:31:28.8347262 -0400 EDT
mtime : 2021-03-29 10:31:28.8347262 -0400 EDT
atime : 2021-03-29 10:31:28.8347262 -0400 EDT
```

## Example - change access time
```
PS C:\github.com\jftuga\gostat> .\gostat.exe -a 20210329.090807 .\README.md
name  : .\README.md
size  : 43
btime : 2021-03-29 08:16:26.7001842 -0400 EDT
ctime : 2021-03-29 10:55:21.2762476 -0400 EDT
mtime : 2021-03-29 08:16:26.7001842 -0400 EDT
atime : 2021-03-29 09:08:07 -0400 EDT
```

## Development

### Building
```bash
go build -ldflags="-s -w"
```

### Testing with Coverage
```bash
go test -v -cover
```
