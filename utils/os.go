package utils

import (
	"github.com/shirou/gopsutil/v3/host"
)

func GetOSInfo() (os, version, arch string, err error) {
	var info *host.InfoStat
	info, err = host.Info()
	if err != nil {
		return
	}
	os = info.OS
	if os == "darwin" {
		os = "osx"
	}
	version = info.PlatformVersion
	arch = info.KernelArch
	return
}
