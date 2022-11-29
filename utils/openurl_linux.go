//go:build linux

package utils

import (
	"os/exec"
	"syscall"
)

func OpenUrl(url string) {
	exec.Command(`xdg-open`, url).Start()
}
