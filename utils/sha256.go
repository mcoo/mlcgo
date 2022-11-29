package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

func Sha256(filePath string) (result string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
