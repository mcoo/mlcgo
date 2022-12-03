package main

import (
	"context"
	"fmt"
	"mlcgo"
	"mlcgo/auth"
	mlclog "mlcgo/log"
	"mlcgo/model"
	"mlcgo/utils"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var log = mlclog.Log

func main() {
	var selectId, memory int
	var version, java, gameDir string
	var versionIsolation bool
	var authType auth.AuthType
	gameDir, _ = os.Getwd()
	gameDir = filepath.Join(gameDir, ".minecraft")
	if len(os.Args) > 1 {
		tmpTag := false
		for _, arg := range os.Args {
			if arg == "-debug" {
				log.SetLevel(logrus.DebugLevel)
			}
			if tmpTag {
				tmpTag = false
				gameDir = arg
			}
			if arg == "-gamedir" {
				tmpTag = true
			}

		}
	}
	versions, err := utils.GetLocalVersions(filepath.Join(gameDir, "versions"))
	if err != nil {
		log.Errorln(err)
	}
	log.Infoln("当前游戏文件版本有：")
	for i, v := range versions {
		log.Printf("[%d] %s", i+1, v)
	}
home:
	log.Info("请选择:")
	fmt.Scan(&selectId)
	if selectId <= 0 || selectId > len(versions) {
		goto home
	}
	version = versions[selectId-1]
	// 获取 Java

	javas, err := utils.FindJavaPath()
	if err != nil {
		log.Info("请输入Java路径:")
		fmt.Scan(&java)
	} else {
		log.Infoln("当前检测到的Java版本有")
		for i, v := range javas {
			log.Printf("[%d] %s", i+1, v)
		}
	javaSelect:
		log.Info("请选择:")
		fmt.Scan(&selectId)
		if selectId <= 0 || selectId > len(javas) {
			goto javaSelect
		}
		java = javas[selectId-1]
	}
memorySet:
	log.Info("最大内存设置为(MB):")
	fmt.Scan(&memory)
	if memory == 0 {
		goto memorySet
	}

	log.Info("版本隔离(0=不启用 1=启用):")
	fmt.Scan(&selectId)
	if selectId == 0 {
		versionIsolation = false
	} else {
		versionIsolation = true
	}

authTypeSet:
	log.Info("验证方式:\n[0] OfflineType\n[1] MicrosoftAuthType\n[2] AuthlibInjectorAuthType")
	fmt.Scan(&authType)
	if authType > 2 || authType < 0 {
		goto authTypeSet
	}
	c := mlcgo.NewCore()
	switch authType {
	case auth.OfflineType:
		log.Debugln("离线认证")
		var name string
		log.Info("用户名:")
		fmt.Scan(&name)
		c.OfflineLogin(name)

	case auth.MicrosoftAuthType:
		log.Debugln("微软认证")
		c.MicrosoftLogin()
	case auth.AuthlibInjectorAuthType:
		log.Debugln("AuthlibInjector认证")
		var url, email, password string
		log.Info("URL:")
		fmt.Scan(&url)

		log.Info("邮箱:")
		fmt.Scan(&email)

		log.Info("密码:")
		fmt.Scan(&password)

		c.AuthlibLogin(url, email, password)
	}

	log.Info("自动补全(0=不启用 1=启用):")
	fmt.Scan(&selectId)
	if selectId == 0 {
		c.NoAutoCompletion()
	}

	ch := make(chan model.Step)
	go func() {
		for {
			v := <-ch
			switch v {
			case model.StopStep:
				log.Println("启动线程停止")
			case model.StartLaunchStep:
				log.Println("开始启动")
			case model.AuthAccountStep:
				log.Println("验证账号")
			case model.GenerateCmdStep:
				log.Println("生成启动命令")
			case model.CompleteFilesStep:
				log.Println("补全游戏文件")
			case model.ExecCmdStep:
				log.Println("执行启动命令")
			}
		}
	}()
	if versionIsolation {
		c.VersionIsolation()
	}
	// dir, _ := os.Getwd()
	log.Infoln(version, java, memory, versionIsolation)
	log.Infoln(c.SetJavaPath(java).
		SetStepChannel(ch).
		Debug().
		SetMinecraftPath(gameDir).
		SetRAM(memory).
		SetVersion(version).Launch(context.Background()))
}
