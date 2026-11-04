// Command pet-migration 只读扫描桌宠配置并输出清单。
//
// 已停止使用：用户已确认改为“仅未设置 / 未知 id 在读取期回落 little-companion，
// 已保存的 stacky / pixiu / CM / 旧对象格式全部保留原值”。本命令不再提供写盘入口，
// 传入 -apply 会被直接拒绝；保留该命令仅用于审计与历史记录复核。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"qiantie/backend/internal/petmigrate"
)

func main() {
	usersDirs := flag.String("users-dirs", "", "用户配置目录，多个用英文逗号分隔")
	reportDir := flag.String("report-dir", "", "清单输出目录，留空则不写文件")
	reportName := flag.String("report-name", "2026-11-04-pet-default-migration-scan", "清单文件名前缀")
	apply := flag.Bool("apply", false, "已禁用：迁移已停止使用，本命令只读")
	flag.Parse()

	if *apply {
		exitWith("迁移已停止使用：本命令不会写盘，请只运行默认的只读扫描")
	}
	if strings.TrimSpace(*usersDirs) == "" {
		exitWith("缺少 -users-dirs")
	}

	var records []petmigrate.Record
	for _, dir := range splitList(*usersDirs) {
		cleaned := filepath.Clean(dir)
		if _, err := os.Stat(cleaned); err != nil {
			exitWith("用户目录不存在：" + cleaned)
		}
		scanned, err := petmigrate.Scan(cleaned, cleaned)
		if err != nil {
			exitWith("扫描失败：" + cleaned + " -> " + err.Error())
		}
		records = append(records, scanned...)
	}
	fmt.Printf("扫描完成：%d 个用户配置，待迁移 %d 条（当前口径恒为 0）\n", len(records), pending(records))

	writeReport(*reportDir, *reportName, records, "", true, 0)
	fmt.Println("只读扫描结束：未写任何用户配置")
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func pending(records []petmigrate.Record) int {
	count := 0
	for _, record := range records {
		if record.Changed {
			count++
		}
	}
	return count
}

func writeReport(reportDir, name string, records []petmigrate.Record, backupPath string, dryRun bool, migrated int) {
	if strings.TrimSpace(reportDir) == "" {
		return
	}
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		exitWith("清单目录创建失败：" + err.Error())
	}
	markdownPath := filepath.Join(reportDir, name+".md")
	csvPath := filepath.Join(reportDir, name+".csv")
	if err := os.WriteFile(markdownPath, []byte(petmigrate.Markdown(records, backupPath, dryRun, migrated)), 0o644); err != nil {
		exitWith("清单写入失败：" + err.Error())
	}
	if err := os.WriteFile(csvPath, []byte(petmigrate.CSV(records)), 0o644); err != nil {
		exitWith("清单写入失败：" + err.Error())
	}
	fmt.Printf("清单：%s\n清单：%s\n", markdownPath, csvPath)
}

func exitWith(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
