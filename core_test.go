package mlcgo

import (
	"context"
	"mlcgo/model"
	"testing"
)

func TestCore(t *testing.T) {
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

	log.Println(NewCore().
		MicrosoftLogin().
		SetJavaPath(`C:\Program Files\Java\jdk-17.0.2\bin\java.exe`).
		SetMinecraftPath(`F:\mc1.19\.minecraft`).
		SetRAM(2048).
		Debug().
		SetVersion("1.19.2-fabric").
		SetStepChannel(ch).
		Launch(context.Background()))
}
