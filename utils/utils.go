package utils

import (
	"errors"
	mlclog "mlcgo/log"
	"mlcgo/model"
	"strings"

	"github.com/imroc/req/v3"
)

var log = mlclog.Log

func IfThen[T any](cond bool, trueVal T, falseVal T) T {
	if cond {
		return trueVal
	} else {
		return falseVal
	}
}

func GetAllMinecraftVersion() (versions *model.McVersions, err error) {
	resp, err := req.R().Get("https://piston-meta.mojang.com/mc/game/version_manifest.json")
	if err != nil {
		return nil, err
	}
	versions = &model.McVersions{}
	err = resp.UnmarshalJson(versions)
	return versions, err
}

func GetLatestMinecraftVersion(versions *model.McVersions, isPrerelease bool) (model.Version, error) {
	var latest string
	if isPrerelease {
		latest = versions.Latest.Release
	} else {
		latest = versions.Latest.Snapshot
	}
	for _, v := range versions.Versions {
		if v.ID == latest {
			return v, nil
		}
	}
	return model.Version{}, errors.New("not found")
}

func ArtifactFrom(name string) (string, error) {
	pts := strings.Split(name, ":")
	if len(pts) < 3 {
		return "", errors.New("error length")
	}
	domain := pts[0]
	name = pts[1]
	ext := "jar"
	last := len(pts) - 1
	if idx := strings.Index(pts[last], "@"); idx != -1 {
		ext = pts[last][idx+1:]
		pts[last] = pts[last][:idx]
	}
	version := pts[2]
	file := name + "-" + version
	if len(pts) > 3 {
		file += "-" + pts[3]
	}
	file += "." + ext
	return strings.ReplaceAll(domain, ".", "/") + "/" + name + "/" + version + "/" + file, nil
}
