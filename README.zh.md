# Minecraft Launcher Core by Golang (MLCG)

[English](README.md) 简体中文

一个简单的我的世界启动器核心

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mcoo/mlcgo?style=for-the-badge) ![GitHub Workflow Status](https://img.shields.io/github/workflow/status/mcoo/mlcgo/goreleaser?style=for-the-badge) ![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/mcoo/mlcgo?include_prereleases&style=for-the-badge)

## 功能 ✈️

- [x] 离线登录
- [x] 自动补全
- [x] 微软登录
- [x] Authlib-injector 登录启动
- [x] 游戏下载
- [x] 跨平台支持
- [x] Forge,Fabric 安装

## 用法 🚀

### 启动游戏

```golang
mlcgo.NewCore().
    OfflineLogin("enjoy").
    SetJavaPath(`C:\Program Files\Java\jdk-17.0.2\bin\java.exe`).
    SetMinecraftPath(`C:\Users\enjoy\AppData\Roaming\.minecraft`).
    SetRAM(2048).
    SetVersion("1.18").
    Launch(context.Background())
```

获取启动状态

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

### 寻找JAVA路径

windows 可用

```golang
utils.FindJavaPath()
```

### 获取本地已安装版本

```golang
utils.GetLocalVersions()
```

### 下载 Minecraft 游戏

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

### 登录验证

#### 微软 Oauth

需要保证8809端口可用

#### AuthlibInjector 验证

```golang
AuthlibLogin(url,email,password)
```

## 注意事项 😀

你应该标明本仓库的地址

- 使用这个项目
- 编辑这个项目

## 支持我 ❤️

## 参考 👍

感谢

[教程/编写启动器](https://minecraft.fandom.com/zh/wiki/%E6%95%99%E7%A8%8B/%E7%BC%96%E5%86%99%E5%90%AF%E5%8A%A8%E5%99%A8) [WIKI]

[gomclauncher](https://github.com/xmdhs/gomclauncher) [MIT]

[Microsoft Authentication Scheme](https://wiki.vg/Microsoft_Authentication_Scheme) [WIKI]
