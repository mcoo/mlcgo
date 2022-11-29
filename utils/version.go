package utils

import (
	"io/ioutil"
	"mlcgo/model"
	"os"
	"path/filepath"
)

func GetAllVersion(path string) (versions []model.Version, e error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, v := range files {
		if v.IsDir() && PathExists(filepath.Join(path, v.Name(), v.Name()+".json")) {
			versions = append(versions, model.Version{Name: v.Name()})
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
