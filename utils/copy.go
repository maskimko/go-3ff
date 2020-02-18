package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//This function should be used to accelerate git clone to the separate directory.
//Actually it is better just copy .tf files and then compare trees
//isGit option omits .git directory and non-*.tf files
func Copy(src, dst string, isGit bool) error {
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		_, err := copyFile(src, dst)
		if err != nil {
			return err
		}
	} else {
		di, err := os.Stat(dst)
		if err != nil {
			return err
		}
		if !di.IsDir() {
			return fmt.Errorf("cannot copy directory %s to file %s", src, dst)
		}
		err = filepath.Walk(src, getWalkFunc(src, dst, isGit))
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	if err != nil {
		return nBytes, err
	}
	err = os.Chmod(dst, sourceFileStat.Mode())
	return nBytes, err
}

func getWalkFunc(src, dst string, isGit bool) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		dest := strings.ReplaceAll(path, src, dst)
		if info.IsDir() {
			if strings.HasSuffix(path, ".git") {
				if isGit {
					return filepath.SkipDir
				}
			} else {
				err := os.Mkdir(dest, info.Mode())
				if err != nil {
					return err
				}
			}
		} else {
			if isGit {
				if !strings.HasSuffix(path, ".tf") {
					return nil
				}
			}
			_, err := copyFile(path, dest)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
