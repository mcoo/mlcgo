package resolver

import (
	"io/ioutil"
	"log"
	"mlcgo/downloader"
	"mlcgo/model"
	"mlcgo/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

func ResolverClient(client []byte, isDemo, isCustomResolution bool, minecraftPath string) (clientInfo *model.ClientInfo, err error) {
	clientInfo = &model.ClientInfo{}
	g := gjson.ParseBytes(client)
	if inherits := g.Get("inheritsFrom"); inherits.Exists() {
		// forge
		cinfo1, err := resolverClient(&g, isDemo, isCustomResolution)
		if err != nil {
			return nil, err
		}
		origin, err := ioutil.ReadFile(filepath.Join(minecraftPath, "versions", inherits.String(), inherits.String()+".json"))
		if err != nil {
			return nil, err
		}
		g1 := gjson.ParseBytes(origin)
		cinfo2, err := resolverClient(&g1, isDemo, isCustomResolution)
		clientInfo.MainClass = cinfo1.MainClass
		clientInfo.AssetIndex = cinfo2.AssetIndex
		clientInfo.GameArguments = append(cinfo2.GameArguments, cinfo1.GameArguments...)
		clientInfo.JavaVersion = cinfo2.JavaVersion
		clientInfo.Libraries = append(cinfo2.Libraries, cinfo1.Libraries...)
		clientInfo.Id = cinfo2.Id
		clientInfo.JvmArguments = append(cinfo2.JvmArguments, cinfo1.JvmArguments...)
		clientInfo.Logging = cinfo2.Logging
		clientInfo.VersionType = cinfo1.VersionType

	} else {
		return resolverClient(&g, isDemo, isCustomResolution)
	}
	return
}

func resolverClient(g *gjson.Result, isDemo, isCustomResolution bool) (clientInfo *model.ClientInfo, err error) {
	clientInfo = &model.ClientInfo{}
	// game 游戏参数列表
	gameArgs := g.Get("arguments.game")
	for _, v := range gameArgs.Array() {
		if v.Type == gjson.String {
			clientInfo.GameArguments = append(clientInfo.GameArguments, v.String())
		}
		if v.Type == gjson.JSON {
			if !v.Get("rules").Exists() {
				continue
			}
			if action, err := RulesIsAction(v.Get("rules"), isDemo, isCustomResolution); action && err == nil {
				v.Get("value").ForEach(func(key, value gjson.Result) bool {
					s := value.String()
					if m := strings.Index(s, "="); strings.Contains(s, " ") && m != -1 {
						s = s[0:m+1] + "\"" + s[m+1:] + "\""

					}
					clientInfo.GameArguments = append(clientInfo.GameArguments, s)
					return true
				})
			} else {
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	// JVM参数列表
	jvmArgs := g.Get("arguments.jvm")
	for _, v := range jvmArgs.Array() {
		if v.Type == gjson.String {
			clientInfo.JvmArguments = append(clientInfo.JvmArguments, v.String())
		}
		if v.Type == gjson.JSON {
			if !v.Get("rules").Exists() {
				continue
			}
			if action, err := RulesIsAction(v.Get("rules"), isDemo, isCustomResolution); action && err == nil {
				v.Get("value").ForEach(func(key, value gjson.Result) bool {
					s := value.String()
					clientInfo.JvmArguments = append(clientInfo.JvmArguments, s)
					return true
				})
			} else {
				if err != nil {
					log.Println(err)
				}
			}

		}
	}
	// Libraries 分析
	os, _, _, _ := utils.GetOSInfo()
	libs := g.Get("libraries")
	for _, v := range libs.Array() {
		if !v.Get("rules").Exists() {
			native := v.Get("natives").Exists()
			if native {
				nativePath := "downloads.classifiers." + v.Get("natives."+os).String()
				path := v.Get("downloads.artifact.path").String()
				if path == "" {
					m := strings.Split(v.Get("name").String(), ":")
					if len(m) == 3 {
						path = filepath.Join(strings.ReplaceAll(m[0], ".", "/"), m[1], m[2], m[1]+"-"+m[2]+".jar")
					}
				}
				clientInfo.Libraries = append(clientInfo.Libraries, model.Library{
					Name:       v.Get("name").String(),
					Sha1:       v.Get("downloads.artifact.sha1").String(),
					Path:       path,
					Url:        v.Get("downloads.artifact.url").String(),
					Size:       v.Get("downloads.artifact.size").Int(),
					Native:     native,
					NativeURL:  v.Get(nativePath + ".url").String(),
					NativeSha1: v.Get(nativePath + ".sha1").String(),
					NativeSize: v.Get(nativePath + ".size").Int(),
					NativePath: v.Get(nativePath + ".path").String(),
				})

			} else {
				path := v.Get("downloads.artifact.path").String()
				if path == "" {
					m := strings.Split(v.Get("name").String(), ":")
					if len(m) == 3 {
						path = filepath.Join(strings.ReplaceAll(m[0], ".", "/"), m[1], m[2], m[1]+"-"+m[2]+".jar")
					}
				}
				clientInfo.Libraries = append(clientInfo.Libraries, model.Library{
					Name: v.Get("name").String(),
					Sha1: v.Get("downloads.artifact.sha1").String(),
					Path: path,
					Url:  v.Get("downloads.artifact.url").String(),
					Size: v.Get("downloads.artifact.size").Int(),
				})
			}

			continue
		}
		if action, err := RulesIsAction(v.Get("rules"), isDemo, isCustomResolution); action && err == nil {
			native := v.Get("natives").Exists()
			if native {
				nativePath := "downloads.classifiers." + v.Get("natives."+os).String()
				path := v.Get("downloads.artifact.path").String()
				if path == "" {
					m := strings.Split(v.Get("name").String(), ":")
					if len(m) == 3 {
						path = filepath.Join(strings.ReplaceAll(m[0], ".", "/"), m[1], m[2], m[1]+"-"+m[2]+".jar")
					}
				}
				clientInfo.Libraries = append(clientInfo.Libraries, model.Library{
					Name:       v.Get("name").String(),
					Sha1:       v.Get("downloads.artifact.sha1").String(),
					Path:       path,
					Url:        v.Get("downloads.artifact.url").String(),
					Size:       v.Get("downloads.artifact.size").Int(),
					Native:     native,
					NativeURL:  v.Get(nativePath + ".url").String(),
					NativeSha1: v.Get(nativePath + ".sha1").String(),
					NativeSize: v.Get(nativePath + ".size").Int(),
					NativePath: v.Get(nativePath + ".path").String(),
				})

			} else {
				path := v.Get("downloads.artifact.path").String()
				if path == "" {
					m := strings.Split(v.Get("name").String(), ":")
					if len(m) == 3 {
						path = filepath.Join(strings.ReplaceAll(m[0], ".", "/"), m[1], m[2], m[1]+"-"+m[2]+".jar")
					}
				}
				clientInfo.Libraries = append(clientInfo.Libraries, model.Library{
					Name: v.Get("name").String(),
					Sha1: v.Get("downloads.artifact.sha1").String(),
					Path: path,
					Url:  v.Get("downloads.artifact.url").String(),
					Size: v.Get("downloads.artifact.size").Int(),
				})
			}
		} else {
			if err != nil {
				log.Println(err)
			}
		}
	}
	// 资源文件
	if assetIndex := g.Get("assetIndex"); assetIndex.Exists() {
		clientInfo.AssetIndex.ID = assetIndex.Get("id").String()
		clientInfo.AssetIndex.Sha1 = assetIndex.Get("sha1").String()
		clientInfo.AssetIndex.Size = int(assetIndex.Get("size").Int())
		clientInfo.AssetIndex.TotalSize = int(assetIndex.Get("totalSize").Int())
		clientInfo.AssetIndex.URL = assetIndex.Get("url").String()
	}
	// logging
	if logging := g.Get("logging.client"); logging.Exists() {
		clientInfo.Logging.Argument = logging.Get("argument").String()
		clientInfo.Logging.Type = logging.Get("type").String()
		clientInfo.Logging.File.ID = logging.Get("file.id").String()
		clientInfo.Logging.File.Sha1 = logging.Get("file.sha1").String()
		clientInfo.Logging.File.Size = int(logging.Get("file.size").Int())
		clientInfo.Logging.File.URL = logging.Get("file.url").String()
	}
	// JAVA Version
	if javaVersion := g.Get("javaVersion"); javaVersion.Exists() {
		clientInfo.JavaVersion.Component = javaVersion.Get("component").String()
		clientInfo.JavaVersion.MajorVersion = int(javaVersion.Get("majorVersion").Int())
	}
	// main class
	clientInfo.MainClass = g.Get("mainClass").String()
	// version type
	clientInfo.VersionType = g.Get("type").String()
	// id
	clientInfo.Id = g.Get("id").String()
	return
}

func ResolverAssets(clientInfo *model.ClientInfo, MinecraftPath string, mirror model.DownloadMirrorSource) (assets []model.Asset, err error) {
	assetsPath := filepath.Join(MinecraftPath, `assets`)
	os.MkdirAll(assetsPath, 0755)
	var jsonByte []byte
	if jsonPath := filepath.Join(assetsPath, "indexes", clientInfo.AssetIndex.ID+".json"); utils.PathExists(jsonPath) {
		sha1, err := utils.SHA1File(jsonPath)
		if err == nil && strings.EqualFold(sha1, clientInfo.AssetIndex.Sha1) {
			jsonByte, _ = ioutil.ReadFile(jsonPath)
		}
	}
	if jsonByte == nil {
		resp, err := req.R().Get(downloader.ReplaceDownloadUrl(clientInfo.AssetIndex.URL, mirror))
		if err != nil {
			return nil, err
		}
		jsonByte = resp.Bytes()
		os.MkdirAll(filepath.Join(assetsPath, "indexes"), 0755)
		err = ioutil.WriteFile(filepath.Join(assetsPath, "indexes", clientInfo.AssetIndex.ID+".json"), jsonByte, 0777)
		if err != nil {
			return nil, err
		}
	}
	// 资源分析
	gjson.GetBytes(jsonByte, "objects").ForEach(func(key, value gjson.Result) bool {
		asset := model.Asset{}
		if hash := value.Get("hash").String(); len(hash) > 4 {
			asset.Hash = hash
			asset.Path = filepath.Join(assetsPath, "objects", hash[0:2], hash)
			asset.Url = "https://resources.download.minecraft.net/" + hash[0:2] + "/" + hash
			asset.Size = int(value.Get("size").Int())
			assets = append(assets, asset)
		}
		return true
	})
	return

}
func RuleMatch(rule gjson.Result, isDemo, isCustomResolution bool) (match bool, err error) {
	os, osVersion, arch, err := utils.GetOSInfo()
	if err != nil {
		return false, err
	}
	match = true
	if v := rule.Get("features"); v.Exists() {
		if m := v.Get("has_custom_resolution"); m.Exists() {
			if m.Bool() != isCustomResolution {
				match = false
				return
			}
		}
		if m := v.Get("is_demo_user"); m.Exists() {
			if m.Bool() != isDemo {
				match = false
				return
			}
		}
	}

	if rule.Get("os").Exists() {
		if os != rule.Get("os.name").String() {
			match = false
			return
		} else {
			if v := rule.Get("os.version"); v.Exists() {
				rxp, err := regexp.Compile(v.String())
				if err != nil {
					return match, err
				}
				if !rxp.MatchString(osVersion) {
					match = false
					return match, err
				}
			}
			if v := rule.Get("os.arch"); v.Exists() {
				if v.String() != arch {
					match = false
					return
				}
			}
		}
	}
	return
}
func RulesIsAction(rules gjson.Result, isDemo, isCustomResolution bool) (action bool, err error) {
	if !rules.Exists() {
		return true, nil
	}
	action = true
	for _, rule := range rules.Array() {
		// disallow 如果能匹配上 直接返回RuleIsAction
		if rule.Get("action").String() == "disallow" {
			if match, err := RuleMatch(rule, isDemo, isCustomResolution); match {
				return false, nil
			} else {
				if err != nil {
					return false, err
				}
			}
		}
		if rule.Get("action").String() == "allow" {
			if match, err := RuleMatch(rule, isDemo, isCustomResolution); !match && err == nil {
				action = false
			} else {
				if err != nil {
					return false, err
				}
			}
		}
	}
	return action, nil
}
