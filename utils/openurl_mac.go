//go:build darwin

package utils

import (
	"os/exec"
	"syscall"
)

func OpenUrl(url string) {
	exec.Command(`open`, url).Start()
}
