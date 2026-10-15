//go:build windows

package worker

import (
	"os/exec"
	"syscall"
)

// createNoWindow 让被拉起的 python.exe 不创建可见控制台窗口。
const createNoWindow = 0x08000000

// hideConsoleWindow 把 OCR 搬运工进程藏到后台，避免黑色命令行窗口反复弹出。
func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
