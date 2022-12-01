package mlcgo

import (
	"mlcgo/auth"
	"mlcgo/utils"
	"path/filepath"
	"regexp"
)

func (c *Core) argumentsReplace(arguments []string) []string {
	c.flagReplaceOnce.Do(func() {
		c.flagReplaceMap = make(map[string]func() string)
		c.flagReplaceMap["${auth_player_name}"] = func() string {
			return c.name
		}
		c.flagReplaceMap["${version_name}"] = func() string {
			return c.clientInfo.Id
		}
		c.flagReplaceMap["${game_directory}"] = func() string {
			if c.gameDir == "" {
				if c.versionIsolation {
					c.gameDir = filepath.Join(c.minecraftPath, `versions`, c.version)
				} else {
					c.gameDir = c.minecraftPath
				}

			}
			log.Debugln("Game Dir:", c.gameDir)
			return c.gameDir
		}
		c.flagReplaceMap["${assets_root}"] = func() string {
			return filepath.Join(c.minecraftPath, `assets`)
		}
		c.flagReplaceMap["${assets_index_name}"] = func() string {
			return c.clientInfo.AssetIndex.ID
		}
		c.flagReplaceMap["${version_type}"] = func() string {
			return "MLCGO"
		}
		c.flagReplaceMap["${launcher_version}"] = func() string {
			return Version + "/MLCGO"
		}
		c.flagReplaceMap["${launcher_name}"] = func() string {
			return "MLCGO"
		}
		c.flagReplaceMap["${version_type}"] = func() string {
			return "MLCGO"
		}
		c.flagReplaceMap["${classpath_separator}"] = func() string {
			return ";"
		}

		c.flagReplaceMap["${natives_directory}"] = func() string {
			if c.nativeDir == "" {
				c.nativeDir = filepath.Join(c.minecraftPath, "versions", c.version, "natives")
			}
			return c.nativeDir
		}
		c.flagReplaceMap["${library_directory}"] = func() string {

			return filepath.Join(c.minecraftPath, `libraries`)
		}
		c.flagReplaceMap["${classpath}"] = func() string {
			return c.generateCP()
		}
		switch c.authType {
		case auth.OfflineType:
			c.flagReplaceMap["${user_type}"] = func() string {
				return "legacy"
			}
			c.flagReplaceMap["${auth_uuid}"] = func() string {
				if c.uuid == "" {
					c.uuid = utils.GenUUID()
				}
				return c.uuid
			}
			c.flagReplaceMap["${auth_access_token}"] = func() string {
				if c.uuid == "" {
					c.uuid = utils.GenUUID()
				}
				return c.uuid
			}
			c.flagReplaceMap["${auth_session}"] = func() string {
				if c.uuid == "" {
					c.uuid = utils.GenUUID()
				}
				return c.uuid
			}
			c.flagReplaceMap["${clientid}"] = func() string {
				if c.uuid == "" {
					c.uuid = utils.GenUUID()
				}
				return c.uuid
			}
			c.flagReplaceMap["${auth_xuid}"] = func() string {
				if c.uuid == "" {
					c.uuid = utils.GenUUID()
				}
				return c.uuid
			}
		case auth.MicrosoftAuthType, auth.AuthlibInjectorAuthType:
			c.flagReplaceMap["${user_type}"] = func() string {
				return "mojang"
			}
			c.flagReplaceMap["${auth_uuid}"] = func() string {
				if c.uuid == "" {
					c.uuid = utils.GenUUID()
				}
				return c.uuid
			}
			c.flagReplaceMap["${auth_xuid}"] = func() string {
				if c.uuid == "" {
					c.uuid = utils.GenUUID()
				}
				return c.uuid
			}
			c.flagReplaceMap["${auth_access_token}"] = func() string {
				return c.accessToken
			}
			c.flagReplaceMap["${auth_session}"] = func() string {
				return c.accessToken
			}
		}

	})
	for i, v := range arguments {

		re := regexp.MustCompile(`\$\{.+?\}`)
		arguments[i] = re.ReplaceAllStringFunc(v, func(s string) string {
			if m, ok := c.flagReplaceMap[s]; ok {
				return m()
			}
			return s
		})

	}
	return arguments
}
