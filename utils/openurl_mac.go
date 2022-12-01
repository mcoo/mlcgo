//go:build darwin

package utils

import (
	"os/exec"
)

func OpenUrl(url string) {
	exec.Command(`open`, url).Start()
}
