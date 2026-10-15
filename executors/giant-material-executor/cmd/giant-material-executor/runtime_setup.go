//go:build windows

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// createNoWindow 与 worker 包保持一致的隐藏窗口标志。
const createNoWindow = 0x08000000

// ensureRuntime 保证本机 Python 具备 OCR 依赖（paddlepaddle/paddleocr 等）。
// 首次使用时按 worker/requirements-lock.txt 自动安装，用户无需手动 pip install。
func ensureRuntime(ctx context.Context, python, exeDir string) error {
	if strings.TrimSpace(python) == "" {
		return errors.New("python command is required")
	}
	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	check := exec.CommandContext(checkCtx, python, "-c", "import paddleocr, cv2")
	hideExecWindow(check)
	check.Dir = exeDir
	if err := check.Run(); err == nil {
		log.Printf("runtime check ok: paddleocr/cv2 available")
		return nil
	}
	requirements := filepath.Join(exeDir, "worker", "requirements-lock.txt")
	if _, err := os.Stat(requirements); err != nil {
		return fmt.Errorf("requirements-lock.txt 不存在，无法自动安装 OCR 依赖: %w", err)
	}
	log.Printf("runtime missing; installing python deps from %s (first use may take several minutes)", requirements)
	install := exec.CommandContext(ctx, python, "-m", "pip", "install", "-r", requirements, "-i", "https://pypi.tuna.tsinghua.edu.cn/simple", "--disable-pip-version-check")
	hideExecWindow(install)
	install.Dir = exeDir
	stdout, err := install.StdoutPipe()
	if err != nil {
		return err
	}
	install.Stderr = install.Stdout
	if err := install.Start(); err != nil {
		return fmt.Errorf("启动 pip 安装失败: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 64*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			log.Printf("[pip] %s", line)
		}
	}
	waitErr := install.Wait()
	if waitErr != nil {
		return fmt.Errorf("OCR 依赖安装失败: %w", waitErr)
	}
	log.Printf("runtime install finished")
	return nil
}

func hideExecWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
