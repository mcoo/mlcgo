package mlcgo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mlcgo/auth"
	mlclog "mlcgo/log"
	"mlcgo/model"
	"mlcgo/resolver"
	"mlcgo/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

const (
	Version = "v0.0.1"
)

type Core struct {
	// F:\mc\.minecraft 默认 ./
	minecraftPath string
	// 4068
	ram int
	// enjoy
	name string

	uuid string

	accessToken string
	// 1.16.5
	version string
	// 隔离时使用
	gameDir string

	extraJvmArgs []string

	extraMinecraftArgs []string

	javaPath string

	isAutoCompletion bool

	isDemo bool

	isCustomResolution bool

	customResolutionWidth int

	customResolutionHigh int

	nativeDir string

	authType auth.AuthType

	clientInfo *model.ClientInfo

	maxDownloadCount int

	stdout io.Writer

	stderr io.Writer

	flagReplaceOnce sync.Once

	flagReplaceMap map[string]func() string

	authCore auth.AuthInterface

	stepCh chan model.Step

	debug bool

	authlibEmail string

	authlibPassword string

	authlibRootUrl string
}

var log = mlclog.Log

// Minecraft路径 eg:F:\mc\.minecraft
func (c *Core) SetMinecraftPath(path string) *Core {
	c.minecraftPath = path
	return c
}

// Game路径 隔离时指定
func (c *Core) SetGamePath(path string) *Core {
	c.gameDir = path
	return c
}

// Java 路径
func (c *Core) SetJavaPath(path string) *Core {
	c.javaPath = path
	return c
}

func (c *Core) SetStepChannel(ch chan model.Step) *Core {
	c.stepCh = ch
	return c
}

func (c *Core) SetRAM(ram int) *Core {
	c.ram = ram
	return c
}

func (c *Core) SetVersion(version string) *Core {
	c.version = version
	return c
}
func (c *Core) SetNativePath(path string) *Core {
	c.nativeDir = path
	return c
}
func (c *Core) SetAccessToken(accessToken string) *Core {
	c.accessToken = accessToken
	return c
}

func (c *Core) SetMaxDownloadCount(maxDownloadCount int) *Core {
	c.maxDownloadCount = maxDownloadCount
	return c
}

// 自动补全
func (c *Core) AutoCompletion() *Core {
	c.isAutoCompletion = true
	return c
}

// 离线登录
func (c *Core) OfflineLogin(username string) *Core {
	c.authType = auth.OfflineType
	c.name = username
	return c
}

// 微软登录
func (c *Core) MicrosoftLogin() *Core {
	c.authType = auth.MicrosoftAuthType
	return c
}

// Authlib 登录
func (c *Core) AuthlibLogin(rootUrl, email, password string) *Core {
	c.authType = auth.AuthlibInjectorAuthType
	c.authlibEmail = email
	c.authlibPassword = password
	c.authlibRootUrl = rootUrl
	return c
}

func (c *Core) CustomResolution(width int, height int) *Core {
	c.customResolutionWidth = width
	c.customResolutionHigh = height
	c.isCustomResolution = true
	return c
}

func (c *Core) Demo() *Core {
	c.isDemo = true
	return c
}
func (c *Core) Debug() *Core {
	c.debug = true
	log.SetLevel(logrus.DebugLevel)
	return c
}

func (c *Core) AddExtraJVMArgs(args []string) *Core {
	c.extraJvmArgs = args
	return c
}

func (c *Core) AddExtraMinecraftArgs(args []string) *Core {
	c.extraMinecraftArgs = args
	return c
}

func CheckDownloadJob(libPath, sha1, url string) *model.DownloadFile {
	if sha1 == "" {
		return nil
	}

	if utils.PathExists(libPath) && strings.ToLower(func(path string) string {
		s, _ := utils.SHA1File(path)
		return s
	}(libPath)) == strings.ToLower(sha1) {
		return nil
	}

	return &model.DownloadFile{
		Url:  url,
		Path: libPath,
		Sha1: sha1,
	}
}

func (c *Core) Launch(ctx context.Context) error {
	defer func() {
		if c.stepCh != nil {
			c.stepCh <- model.StopStep
		}
	}()
	var cmdArgs []string
	if c.stepCh != nil {
		c.stepCh <- model.StartLaunchStep
	}
	// Login First
	if c.stepCh != nil {
		c.stepCh <- model.AuthAccountStep
	}
	switch c.authType {
	case auth.OfflineType:
	case auth.MicrosoftAuthType:
		c.authCore = &auth.MicrosoftAuth{}
		loginInfo, err := c.authCore.Auth(map[string]string{
			"configPath": filepath.Join(c.minecraftPath, "microsoft_auth.config"),
		})
		if err != nil {
			return err
		}
		c.uuid = loginInfo.UUID
		c.accessToken = loginInfo.AccessToken
		c.name = loginInfo.Name
	case auth.AuthlibInjectorAuthType:
		c.authCore = &auth.AuthlibInjectorAuth{}
		loginInfo, err := c.authCore.Auth(map[string]string{
			"configPath":  filepath.Join(c.minecraftPath, "authlib_auth.config"),
			"email":       c.authlibEmail,
			"password":    c.authlibPassword,
			"root_url":    c.authlibRootUrl,
			"authJarPath": filepath.Join(c.minecraftPath, "authlib.jar"),
		})
		if err != nil {
			return err
		}
		c.uuid = loginInfo.UUID
		c.accessToken = loginInfo.AccessToken
		c.name = loginInfo.Name
		resp, _ := req.R().Get(c.authlibRootUrl)

		cmdArgs = append(cmdArgs, "-javaagent:"+filepath.Join(c.minecraftPath, "authlib.jar")+"="+c.authlibRootUrl, "-Dauthlibinjector.yggdrasil.prefetched="+base64.StdEncoding.EncodeToString(resp.Bytes()))
	}
	if c.stepCh != nil {
		c.stepCh <- model.GenerateCmdStep
	}

	clientJsonPath := filepath.Join(c.minecraftPath, "versions", c.version, c.version+".json")
	if !utils.PathExists(clientJsonPath) {
		return errors.New("client json file not found")
	}
	clientJsonByte, err := os.ReadFile(clientJsonPath)
	if err != nil {
		return errors.New("client json read error")
	}
	clientInfo, err := resolver.ResolverClient(clientJsonByte, c.isDemo, c.isCustomResolution, c.minecraftPath)
	if err != nil {
		return err
	}
	c.clientInfo = clientInfo

	cmdArgs = append(cmdArgs, c.argumentsReplace(c.clientInfo.JvmArguments)...)
	// 内存
	cmdArgs = append(cmdArgs, "-Xmx"+strconv.Itoa(c.ram)+"m")

	cmdArgs = append(cmdArgs, c.extraJvmArgs...)
	// main class
	cmdArgs = append(cmdArgs, c.clientInfo.MainClass)

	cmdArgs = append(cmdArgs, c.argumentsReplace(c.clientInfo.GameArguments)...)

	cmdArgs = append(cmdArgs, c.extraMinecraftArgs...)
	//log.Println(cmdArgs)

	// 补全 libraries 和 资源文件

	if c.isAutoCompletion {
		if c.stepCh != nil {
			c.stepCh <- model.CompleteFilesStep
		}
		var wg sync.WaitGroup
		if c.maxDownloadCount == 0 {
			c.maxDownloadCount = runtime.NumCPU()
		}
		assets, err := resolver.ResolverAssets(c.clientInfo, c.minecraftPath)
		if err != nil {
			return err
		}
		downloadCh := make(chan model.DownloadFile, 10)
		go func() {
			defer close(downloadCh)
			for _, v := range c.clientInfo.Libraries {
				if v.Native {
					libPath := filepath.Join(c.minecraftPath, "libraries", v.NativePath)
					job := CheckDownloadJob(libPath, v.NativeSha1, v.NativeURL)
					if job != nil {
						select {
						case <-ctx.Done():
							log.Debugln("接收到结束指令")
							close(downloadCh)
							return
						case downloadCh <- *job:
							if v.Path == v.NativePath {
								continue
							}
						}
					}
				}
				libPath := filepath.Join(c.minecraftPath, "libraries", v.Path)
				job := CheckDownloadJob(libPath, v.Sha1, v.Url)
				if job != nil {
					select {
					case <-ctx.Done():
						log.Debugln("接收到结束指令")
						close(downloadCh)
						return
					case downloadCh <- *job:
					}
				}
			}
			for _, v := range assets {
				job := CheckDownloadJob(v.Path, v.Hash, v.Url)
				if job != nil {
					select {
					case <-ctx.Done():
						log.Debugln("接收到结束指令")
						close(downloadCh)
						return
					case downloadCh <- *job:
					}
				}
			}

		}()

		for i := 0; i < c.maxDownloadCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				DownloadWork(ctx, downloadCh)
			}()
		}
		wg.Wait()

		select {
		case <-ctx.Done():
			return errors.New("接收到结束指令")
		default:
		}

		// 解压 native
		for _, v := range c.clientInfo.Libraries {
			if v.Native {
				utils.UnzipNative(filepath.Join(c.minecraftPath, "libraries", v.NativePath), c.nativeDir)
			}
		}
	}
	c.generateLauncherProfiles()
	if c.javaPath == "" {
		c.javaPath = "javaw.exe"
	}
	if c.stepCh != nil {
		c.stepCh <- model.ExecCmdStep
	}
	//cmd := exec.Command(c.javaPath, cmdArgs...)
	cmd := exec.Command(c.javaPath, cmdArgs...)
	cmd.Stderr = c.stderr
	cmd.Stdout = c.stdout
	cmd.Dir = c.minecraftPath

	if c.debug {
		log.Debugln("'" + strings.Join(cmdArgs, "' '") + "'")
	}
	return cmd.Run()
}

func DownloadWork(ctx context.Context, workCh chan model.DownloadFile) error {
	for job := range workCh {
		dir := filepath.Dir(job.Path)
		os.MkdirAll(dir, 0755)
		_, err := req.R().SetOutputFile(job.Path).Get(job.Url)
		if err != nil {
			log.Errorln(job.Path, job.Url, err)
		}
	}
	return nil
}

func (c *Core) generateCP() string {
	cp := ""
	//version jar
	cp += filepath.Join(c.minecraftPath, "versions", c.clientInfo.Id, c.clientInfo.Id+".jar") + ";"
	if p := filepath.Join(c.minecraftPath, "versions", c.version, c.version+".jar"); utils.PathExists(p) && p != filepath.Join(c.minecraftPath, "versions", c.clientInfo.Id, c.clientInfo.Id+".jar") {
		cp += filepath.Join(c.minecraftPath, "versions", c.version, c.version+".jar") + ";"
	}

	// libraries
	libs := []string{}
	for _, v := range c.clientInfo.Libraries {
		if !v.Native {
			libs = append(libs, filepath.Join(c.minecraftPath, "libraries", v.Path))
		}
	}
	return cp + strings.Join(libs, ";")
}

func (c *Core) generateLauncherProfiles() {
	var j model.LauncherProfiles
	j.SelectedProfileName = "MLCGO"
	j.Profiles.Mlcgo.GameDir = c.minecraftPath
	j.Profiles.Mlcgo.LastVersionID = c.version
	j.Profiles.Mlcgo.Name = "MLCGO"
	b, _ := json.Marshal(&j)
	os.WriteFile(filepath.Join(c.minecraftPath, "launcher_profiles.json"),
		b,
		0777,
	)
}

func (c *Core) argumentsReplace(arguments []string) []string {
	c.flagReplaceOnce.Do(func() {
		c.flagReplaceMap = make(map[string]func() string)
		c.flagReplaceMap["${auth_player_name}"] = func() string {
			return utils.PreReplace(c.name)
		}
		c.flagReplaceMap["${version_name}"] = func() string {
			return utils.PreReplace(c.clientInfo.Id)
		}
		c.flagReplaceMap["${game_directory}"] = func() string {
			if c.gameDir == "" {
				c.gameDir = utils.PreReplace(c.minecraftPath)
			}
			return utils.PreReplace(c.gameDir)
		}
		c.flagReplaceMap["${assets_root}"] = func() string {
			return utils.PreReplace(filepath.Join(c.minecraftPath, `assets`))
		}
		c.flagReplaceMap["${assets_index_name}"] = func() string {
			return utils.PreReplace(c.clientInfo.AssetIndex.ID)
		}
		c.flagReplaceMap["${version_type}"] = func() string {
			return "MLCGO"
		}
		c.flagReplaceMap["${launcher_version}"] = func() string {
			return utils.PreReplace(Version) + "/MLCGO"
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
				c.nativeDir = utils.PreReplace(filepath.Join(c.minecraftPath, "versions", c.version, "natives"))
			}
			return c.nativeDir
		}
		c.flagReplaceMap["${library_directory}"] = func() string {

			return utils.PreReplace(filepath.Join(c.minecraftPath, `libraries`))
		}
		c.flagReplaceMap["${classpath}"] = func() string {
			return utils.PreReplace(c.generateCP())
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

func NewCore() *Core {
	return &Core{isAutoCompletion: true, stdout: os.Stdout, stderr: os.Stderr}
}
