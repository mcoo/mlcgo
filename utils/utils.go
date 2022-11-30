package utils

import (
	"errors"
	"mlcgo/model"

	"github.com/imroc/req/v3"
)

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
