//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

func OpenUrl(url string) {
	cmd := exec.Command(`cmd`, `/c`, `start`, url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}
