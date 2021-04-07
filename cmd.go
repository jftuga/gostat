/*
gostat.go
-John Taylor
Mar-29-2021

Display and set file time stamps
LICENSE: MIT License

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/djherbis/times.v1"
)

const pgmName string = "gostat"
const pgmDesc string = "Display and set file time stamps"
const pgmURL string = "https://github.com/jftuga/gostat"
const pgmLicense = "https://github.com/jftuga/gostat/blob/main/LICENSE"
const pgmVersion string = "1.0.1"

// expandGlobs - expand file wildcards into a list of file names
func expandGlobs(args []string) []string {
	var allFiles []string
	for _, glob := range args {
		globbed, err := filepath.Glob(glob)
		if err != nil{
			log.Printf("Glob Error: %s\n", err)
			continue
		}
		for _, file := range globbed {
			allFiles = append(allFiles, file)
		}
	}
	return allFiles
}
// Format - add thousands commas to an integer
// https://stackoverflow.com/a/31046325/452281
func Format(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}
// getFileTimes - return a small map containing time metadata for a single file
func getFileTimes(file string) map [string]time.Time {
	fileTimes := make(map [string]time.Time)
	t, err := times.Stat(file)
	if err != nil {
		log.Printf("getFileTimes Error: %s\n", err.Error())
		return fileTimes
	}
	fileTimes["a"] = t.AccessTime()
	fileTimes["m"] = t.ModTime()

	if t.HasChangeTime() {
		fileTimes["c"] = t.ChangeTime()
	}
	if t.HasBirthTime() {
		fileTimes["b"] = t.BirthTime()
	}
	return fileTimes
}

// showFileTimes - output file name, size; birth, create, modify, and access times
func showFileTimes(args []string) int {
	var fi os.FileInfo
	var err error
	count := 0
	for _, file := range expandGlobs(args) {
		fmt.Printf("name  : %s\n", file)
		fi, err = os.Stat(file)
		if err != nil {
			log.Printf("Lstat Error: %s\n", err)
			continue
		}
		count += 1
		fmt.Printf("size  : %s\n", Format(fi.Size()))
		t := getFileTimes(file)
		if b, found := t["b"]; found {
			fmt.Printf("btime : %s\n", b)
		}
		if c, found := t["c"]; found {
			fmt.Printf("ctime : %s\n", c)
		}
		fmt.Printf("mtime : %s\n", t["m"])
		fmt.Printf("atime : %s\n", t["a"])

		fmt.Println()
	}
	return count
}

// convertStr - convert a string to an int
func convertStr(location string, s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Invalid %s: %s\n", location, s)
	}
	return i
}

// createDate - return a time.Time value when given a string in YYYYMMDD.HHMMSS format
func createDate(dt string) time.Time {
	year := convertStr("year", dt[0:4])
	month := convertStr("month", dt[4:6])
	day := convertStr("day", dt[6:8])
	hour := convertStr("hour", dt[9:11])
	minute := convertStr("minute", dt[11:13])
	second := convertStr("second", dt[13:15])

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Now().Location())
}

// setFileTime - update a timestamps for a group of files
// op should equal: (a)ccess, (m)odify, (b)oth
func setFileTime(args []string, dt, op string) {
	var err error

	for _, file := range expandGlobs(args) {
		currentTimes := getFileTimes(file)
		if "m" == op {
			err = os.Chtimes(file, currentTimes["a"], createDate(dt))
		} else if "a" == op {
			fmt.Println(createDate(dt))
			err = os.Chtimes(file, createDate(dt), currentTimes["m"])
		}  else if "b" == op {
			dateTime := createDate(dt)
			err = os.Chtimes(file, dateTime, dateTime)
		} else {
			log.Fatalf("Invalid op: %s\n", op)
		}
		if err != nil {
			log.Printf("os.Chtimes Error: %s\n", err.Error())
			continue
		}
		showFileTimes([]string {file})
	}
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [FILE]...\n", pgmName)
	fmt.Fprintf(os.Stderr, "%s\n\n", pgmDesc)
	flag.PrintDefaults()
}

func showVersion() {
	fmt.Fprintf(os.Stderr, "%s\n", pgmName)
	fmt.Fprintf(os.Stderr, "%s\n", pgmDesc)
	fmt.Fprintf(os.Stderr, "version: %s\n", pgmVersion)
	fmt.Fprintf(os.Stderr, "homepage: %s\n", pgmURL)
	fmt.Fprintf(os.Stderr, "license: %s\n\n", pgmLicense)
}

func main() {
	argsVersion := flag.Bool("v", false, "show program version and then exit")
	argsAccess := flag.String("a", "", "set file access time, format: YYYYMMDD.HHMMSS")
	argsModify := flag.String("m", "", "set file modify time, format: YYYYMMDD.HHMMSS")
	argsBoth := flag.String("b", "", "set both access and modify time, format: YYYYMMDD.HHMMSS")
	flag.Usage = showUsage
	flag.Parse()

	if *argsVersion {
		showVersion()
		os.Exit(0)
	}

	args := flag.Args()
	if 0 == len(args) {
		showUsage()
		os.Exit(1)
	}

	wantChange := 0
	op := ""
	newTime := ""
	if len(*argsAccess) > 0 {
		wantChange += 1
		op = "a"
		newTime = *argsAccess
	}
	if len(*argsModify) > 0 {
		wantChange += 1
		op = "m"
		newTime = *argsModify
	}
	if len(*argsBoth) > 0 {
		wantChange += 1
		op = "b"
		newTime = *argsBoth
	}
	if wantChange > 1 {
		log.Fatalf("-a, -m, and -b are all mutually exclusive\n")
	}

	if wantChange > 0 {
		validDT := regexp.MustCompile(`20\d{2}\d{2}\d{2}.\d{2}\d{2}\d{2}$`)
		if validDT.MatchString(newTime) == false {
			log.Fatalf("Error: invalid time stamp: %s\nPlease use: YYYYMMDD.HHMMSS\n", newTime)
		}
		setFileTime(args, newTime, op)
		os.Exit(0)
	}

	count := showFileTimes(args)
	if count == 0 {
		log.Fatalf("Error: %s did not match any files\n", args)
	}
}
