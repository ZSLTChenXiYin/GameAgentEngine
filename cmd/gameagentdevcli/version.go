package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/version"
)

// devCliVersion 在打包时通过 -ldflags -X 注入，默认值来自 version 包。
var devCliVersion = version.Version

var devCliVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示 DevCli 及连接的 Engine 版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GameAgentDevCli %s (min compatible: %s)\n", devCliVersion, version.MinCompatibleVersion)

		engineVer, engineMin, err := newClient().GetVersion()
		if err != nil {
			fmt.Println("无法获取引擎版本信息：", err)
			return
		}
		fmt.Printf("Engine %s (min compatible: %s)\n", engineVer, engineMin)

		compatible, msg := version.Check(devCliVersion, engineVer)
		if compatible {
			fmt.Printf("✓ %s\n", msg)
		} else {
			fmt.Printf("✗ %s\n", msg)
		}
	},
}

// checkEngineVersion 连接引擎并检查版本兼容性，不兼容时终止进程。
func checkEngineVersion() {
	engineVer, _, err := newClient().GetVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告：无法连接引擎获取版本信息：%v\n", err)
		fmt.Fprintf(os.Stderr, "请确认引擎服务正在运行（默认 http://127.0.0.1:8080）\n")
		return
	}

	compatible, msg := version.Check(devCliVersion, engineVer)
	if !compatible {
		fmt.Fprintf(os.Stderr, "版本不兼容：%s\n", msg)
		fmt.Fprintf(os.Stderr, "DevCli %s 无法连接 Engine %s\n", devCliVersion, engineVer)
		os.Exit(1)
	}
}

