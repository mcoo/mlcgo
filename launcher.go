package mlcgo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"mlcgo/auth"
	"mlcgo/model"
	"mlcgo/resolver"
	"mlcgo/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/imroc/req/v3"
)

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
		c.autoCompletion(ctx, nil)
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

	log.Debugln("'" + strings.Join(cmdArgs, "' '") + "'")

	return cmd.Run()
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
