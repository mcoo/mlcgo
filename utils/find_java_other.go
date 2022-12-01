//go:build !windows

package utils

import "errors"

func FindJavaPath() (j []string, e error) {
	return nil, errors.New("Err")
}
