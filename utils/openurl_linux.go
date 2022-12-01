//go:build linux

package utils

import (
	"os/exec"
)

func OpenUrl(url string) {
	exec.Command(`xdg-open`, url).Start()
}
