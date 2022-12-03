package mlcgo

import (
	"io"
	"mlcgo/auth"
	mlclog "mlcgo/log"
	"mlcgo/model"
	"mlcgo/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

	versionIsolation bool
}

var log = mlclog.Log

// Minecraft路径 eg:F:\mc\.minecraft
func (c *Core) SetMinecraftPath(path string) *Core {
	path, _ = filepath.Abs(path)
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

func (c *Core) VersionIsolation() *Core {
	c.versionIsolation = true
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

// 不自动补全
func (c *Core) NoAutoCompletion() *Core {
	c.isAutoCompletion = false
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

	if utils.PathExists(libPath) && strings.EqualFold(func(path string) string {
		s, _ := utils.SHA1File(path)
		return s
	}(libPath), sha1) {
		return nil
	}

	return &model.DownloadFile{
		Url:  url,
		Path: libPath,
		Sha1: sha1,
	}
}

func NewCore() *Core {
	return &Core{isAutoCompletion: true, stdout: os.Stdout, stderr: os.Stderr}
}
