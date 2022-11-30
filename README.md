# Minecraft Launcher Core by Golang (MLCG)

English [简体中文](README.zh.md)

A Easy Minecraft Launcher Core

## Features ✈️

- [x] Offline Launcher
- [x] Auto Completion
- [x] Microsoft Oauth Launcher
- [x] Authlib-injector Launcher
- [x] Download Game

## Use 🚀

### Launcher

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

## Special Note 😀

you **must** mark the link of this repository in your software.

- Use this project
- Edit this project

## Support Me ❤️

## Reference 👍

Thanks

[教程/编写启动器](https://minecraft.fandom.com/zh/wiki/%E6%95%99%E7%A8%8B/%E7%BC%96%E5%86%99%E5%90%AF%E5%8A%A8%E5%99%A8) [WIKI]

[gomclauncher](https://github.com/xmdhs/gomclauncher) [MIT]

[Microsoft Authentication Scheme](https://wiki.vg/Microsoft_Authentication_Scheme) [WIKI]
