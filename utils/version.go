package utils

import (
	"os"
	"path/filepath"
)

func GetLocalVersions(path string) (versions []string, e error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, v := range files {
		if v.IsDir() && PathExists(filepath.Join(path, v.Name(), v.Name()+".json")) {
			versions = append(versions, v.Name())
		}
	}
	return
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
