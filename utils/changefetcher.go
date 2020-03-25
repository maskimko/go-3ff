package utils

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

//TODO Use viper.GetBool here
var Debug bool = false
var fileMap map[string][]byte = make(map[string][]byte, 0)
var CacheHits int = 0
var lock sync.Mutex
var Logger *log.Logger

func GetStringAtPos(start, end int, filename string) (string, error) {
	if end < start {
		return "", errors.New("end position cannot be smaller than start position")
	}
	if start == end {
		return "", nil
	}
	if fd, ok := fileMap[filename]; ok {
		if Debug {
			Logger.Printf("Using cached file %s", filename)
		}
		if start > len(fd) {
			if Debug {
				Logger.Printf("Start %d is after the file length %d", start, len(fd))
			}
			return "", errors.New(fmt.Sprintf("Start %d is after the file length %d", start, len(fd)))
		}
		if end > len(fd)-1 {
			if Debug {
				Logger.Printf("End %d is after the file length %d", end, len(fd))
			}
			return "", errors.New(fmt.Sprintf("End %d is after the file length %d", end, len(fd)))
		}
		data := fd[start:end]
		CacheHits = CacheHits + 1
		return string(data), nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	lock.Lock()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot cache file %s Error: %s", filename, err)
		}
	}
	fileMap[filename] = bytes
	lock.Unlock()
	buf := make([]byte, end-start)
	n, err := file.ReadAt(buf, int64(start))
	s := string(buf)
	if err != nil {
		if err == io.EOF {
			if Debug {
				if n < end-start {
					Logger.Printf("Reached the end of the file %s before expected end of string \"%s\"", filename, s)
				} else {
					Logger.Printf("Reached the end of the file %s while fetching string \"%s\"", filename, s)
				}
			}
		} else {
			if Debug {
				Logger.Printf("Cannot read string at given offset range %d-%d from file %s. Error: %s", start, end, filename, err)
			}
			return "", err
		}
	}
	return s, nil
}
func GetStringFromRange(r hcl.Range) (string, error) {
	return GetStringAtPos(r.Start.Byte, r.End.Byte, r.Filename)
}

func GetStringFromHclSyntaxExpression(e hclsyntax.Expression) string {
	r := e.Range()
	val, err := GetStringFromRange(r)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot get string value of expression by range. File %s, (%d:%d - %d:%d)",
				r.Filename, r.Start.Line, r.Start.Column,
				r.End.Line, r.End.Column)
		}
		val = "unavailable"
	}
	return val
}

func GetChangeLogString(orig, modif hcl.Range) (string, error) {
	origString, err := GetStringFromRange(orig)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot fetch original string from file %s", orig.Filename)
		}
		return "", err
	}

	modifiedString, err := GetStringFromRange(modif)
	if err != nil {
		if Debug {
			Logger.Printf("Cannot fetch modified string from file %s", orig.Filename)
		}
		return "", err
	}
	arrow := color.YellowString("->")
	ls := fmt.Sprintf("%s %s %s", origString, arrow, modifiedString)
	return ls, nil
}
