# Minecraft Launcher Core by Golang (MLCG)

English [简体中文](README.zh.md)

A Easy Minecraft Launcher Core

## Features ✈️

- [x] Offline Launcher
- [x] Auto Completion
- [x] Microsoft Oauth Launcher
- [x] Authlib-injector Launcher
- [x] Download Game
- [x] Cross platform support

## Use 🚀

### install

```sh
go get -u github.com/mcoo/mlcgo@latest
```

```golang
import "github.com/mcoo/mlcgo"
```

### Launch Game

```golang
mlcgo.NewCore().
    OfflineLogin("enjoy").
    SetJavaPath(`C:\Program Files\Java\jdk-17.0.2\bin\java.exe`).
    SetMinecraftPath(`C:\Users\enjoy\AppData\Roaming\.minecraft`).
    SetRAM(2048).
    SetVersion("1.18").
    Launch(context.Background())
```

Get launch status

```golang
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
```

### Find Java Path

windows only

```golang
utils.FindJavaPath()
```

### Get Local Versions

```golang
utils.GetLocalVersions()
```

### Download Minecraft Game

```golang
versions, err := utils.GetAllMinecraftVersion()
	if err != nil {
		log.Errorln(err)
	}
	var c = NewCore().
		SetJavaPath(`C:\Program Files\Java\jdk-17.0.2\bin\java.exe`).
		SetMinecraftPath(`F:\mctest\.minecraft`).
		SetRAM(2048).
		OfflineLogin("enjoy").
		Debug()
	log.Println(c.DownloadGame(context.Background(), versions.Versions[4]))
	c.Launch(context.Background())
```

### Auth

#### Microsoft Oauth

You should keep the 8809 port free.

#### AuthlibInjector Auth

```golang
AuthlibLogin(url,email,password)
```

## Special Note 😀

you **must** mark the link of this repository in your software.

- Use this project
- Modify this project

## Support Me ❤️

## Reference 👍

Thanks

[教程/编写启动器](https://minecraft.fandom.com/zh/wiki/%E6%95%99%E7%A8%8B/%E7%BC%96%E5%86%99%E5%90%AF%E5%8A%A8%E5%99%A8) [WIKI]

[gomclauncher](https://github.com/xmdhs/gomclauncher) [MIT]

[Microsoft Authentication Scheme](https://wiki.vg/Microsoft_Authentication_Scheme) [WIKI]
