# Minecraft Launcher Core by Golang (MLCG)

[English](README.md) 简体中文

一个简单的我的世界启动器核心

## 功能 ✈️

- [x] 离线登录
- [x] 自动补全
- [x] 微软登录
- [x] Authlib-injector 登录启动

## 用法 🚀

### 启动

```golang
func TestCore(t *testing.T) {
    t.Log(NewCore().
        OfflineLogin("enjoy").
        SetJavaPath(`C:\Program Files\Java\jdk-17.0.2\bin\java.exe`).
        SetMinecraftPath(`C:\Users\enjoy\AppData\Roaming\.minecraft`).
        SetRAM(2048).
        SetVersion("1.18").
        Launch(context.Background()))
}
```

## 注意事项 😀

你应该标明本仓库的地址

- 使用这个项目
- 编辑这个项目

## 支持我 ❤️

## 参考 👍

感谢

[教程/编写启动器](https://minecraft.fandom.com/zh/wiki/%E6%95%99%E7%A8%8B/%E7%BC%96%E5%86%99%E5%90%AF%E5%8A%A8%E5%99%A8) [WIKI]

[
gomclauncher
](https://github.com/xmdhs/gomclauncher) [MIT]

[Microsoft Authentication Scheme](https://wiki.vg/Microsoft_Authentication_Scheme) [WIKI]
