//go:build windows

package ui

import (
	"context"
	_ "embed"
	"errors"
	"os/exec"
	"strings"

	"fyne.io/systray"
)

//go:embed assets/icon.ico
var iconICO []byte

const setupURL = "http://127.0.0.1:17861/setup"

// Run 在系统托盘显示常驻图标（用户由此能看到执行器在后台运行），
// 并在首次启动时用默认浏览器打开配置页面。托盘菜单支持打开页面与退出程序。
func Run(ctx context.Context, config Config) error {
	if err := validateConfig(config); err != nil {
		return err
	}
	if err := openSetupPage(); err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		systray.Quit()
	}()
	systray.Run(func() {
		systray.SetIcon(iconICO)
		title := "巨量素材执行器"
		if v := strings.TrimSpace(config.Version); v != "" && v != "dev" {
			title += " " + v
		}
		systray.SetTooltip(title)
		mOpen := systray.AddMenuItem("打开配置页面", "在浏览器中打开执行器状态页")
		mOpen.SetIcon(iconICO)
		systray.AddSeparator()
		mExit := systray.AddMenuItem("退出程序", "停止执行器和 OCR 进程")
		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					_ = openSetupPage()
				case <-mExit.ClickedCh:
					if config.Shutdown != nil {
						config.Shutdown()
					}
					systray.Quit()
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}, nil)
	return nil
}

func openSetupPage() error {
	if err := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", setupURL).Start(); err != nil {
		return errors.New("无法打开执行器配对页面：" + err.Error())
	}
	return nil
}
