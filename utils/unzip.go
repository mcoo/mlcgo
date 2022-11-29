package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnzipNative(zipPath string, target string) error {
	os.MkdirAll(target, 0755)
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, v := range reader.File {
		if v.FileInfo().IsDir() {
			continue
		}
		name := v.FileInfo().Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".dll" || ext == ".so" || ext == ".dylib" {
			path := filepath.Join(target, name)
			rc, err := v.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			w, err := os.Create(path)
			if err != nil {
				return err
			}
			defer w.Close()
			_, err = io.Copy(w, rc)
			if err != nil {
				return err
			}

		}
	}
	return nil
}
