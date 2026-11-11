// api-secret-migration 把用户 api-config.json 里的明文渠道密钥批量转成
// enc:v1: 密文（与线上保存时使用完全相同的字段口径）。
//
// 默认只读预览，不会修改任何文件：
//
//	go run ./cmd/api-secret-migration -users-dir data/users
//
// 确认预览结果后，再由负责人手动执行实际迁移：
//
//	go run ./cmd/api-secret-migration -users-dir data/users -apply
//
// 常用参数：
//
//	-users-dir   用户目录（默认取 QIANTIE_SCRIPT_HISTORY_DIR，否则 data/users）
//	-key-file    密钥文件路径（默认 <users-dir 的上一级>/system/config-secret.key）
//	-backup      实际迁移前备份原文件（默认 true）
//	-print-env   只打印一把新密钥的 base64，便于配置 QIANTIE_CONFIG_SECRET_KEY
//
// 幂等性：已经带 enc:v1: 前缀的值会跳过；重复执行不会二次加密；没有变化的文件不会重写。
// 安全：输出只包含文件路径和数量，不打印任何密钥明文或密文。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"qiantie/backend/internal/apisecrets"
	"qiantie/backend/internal/httpapi"
)

type options struct {
	usersDir string
	keyFile  string
	apply    bool
	backup   bool
	printEnv bool
}

func main() {
	opts := options{}
	flag.StringVar(&opts.usersDir, "users-dir", defaultUsersDir(), "用户目录（每个用户一个子目录，内含 api-config.json）")
	flag.StringVar(&opts.keyFile, "key-file", "", "密钥文件路径；默认 <users-dir 的上一级>/system/config-secret.key")
	flag.BoolVar(&opts.apply, "apply", false, "实际写入迁移结果；不加该参数时只做只读预览")
	flag.BoolVar(&opts.backup, "backup", true, "实际迁移前把原文件备份为 <file>.bak-<时间戳>")
	flag.BoolVar(&opts.printEnv, "print-env", false, "只打印一把新密钥的 base64，不扫描任何目录")
	flag.Parse()

	if opts.printEnv {
		key, err := apisecrets.GenerateKeyBase64()
		if err != nil {
			fail("生成密钥失败: %v", err)
		}
		fmt.Printf("QIANTIE_CONFIG_SECRET_KEY=%s\n", key)
		return
	}

	usersDir := strings.TrimSpace(opts.usersDir)
	if usersDir == "" {
		fail("缺少 -users-dir")
	}
	info, err := os.Stat(usersDir)
	if err != nil || !info.IsDir() {
		fail("用户目录不可用: %s", usersDir)
	}
	if strings.TrimSpace(opts.keyFile) == "" {
		opts.keyFile = filepath.Join(filepath.Dir(filepath.Clean(usersDir)), "system", "config-secret.key")
	}

	files, err := findConfigFiles(usersDir)
	if err != nil {
		fail("扫描用户目录失败: %v", err)
	}
	if len(files) == 0 {
		fmt.Printf("未找到任何 api-config.json：%s\n", usersDir)
		return
	}

	var cipher *apisecrets.Cipher
	if opts.apply {
		// WriteCipher 在密钥文件不存在时会惰性生成，保证迁移后新数据与旧数据使用同一把密钥。
		cipher, err = (&apisecrets.Provider{KeyFile: opts.keyFile}).WriteCipher()
		if err != nil {
			fail("加密密钥不可用（%s）：%v", opts.keyFile, err)
		}
	}

	totalFiles, totalFields, changedFiles, encryptedFields := 0, 0, 0, 0
	for _, path := range files {
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Printf("[跳过] %s：读取失败\n", path)
			continue
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			fmt.Printf("[跳过] %s：JSON 解析失败\n", path)
			continue
		}

		pending := httpapi.CountPlaintextConfigSecrets(document)
		if pending == 0 {
			fmt.Printf("[已是最新] %s\n", path)
			continue
		}
		totalFiles++
		totalFields += pending

		if !opts.apply {
			fmt.Printf("[预览] %s：待加密字段 %d 个\n", path, pending)
			continue
		}

		changed, encryptErr := httpapi.EncryptConfigSecretsForMigration(document, cipher)
		if encryptErr != nil && changed == 0 {
			fmt.Printf("[失败] %s：加密失败：%v\n", path, encryptErr)
			continue
		}
		encoded, marshalErr := json.MarshalIndent(document, "", "  ")
		if marshalErr != nil {
			fmt.Printf("[失败] %s：序列化失败\n", path)
			continue
		}
		if opts.backup {
			backupPath := fmt.Sprintf("%s.bak-%s", path, time.Now().Format("20060102-150405"))
			if err := os.WriteFile(backupPath, raw, 0o600); err != nil {
				fmt.Printf("[失败] %s：备份失败：%v\n", path, err)
				continue
			}
		}
		if err := writeAtomic(path, encoded); err != nil {
			fmt.Printf("[失败] %s：写入失败：%v\n", path, err)
			continue
		}
		changedFiles++
		encryptedFields += changed
		fmt.Printf("[完成] %s：已加密字段 %d 个\n", path, changed)
		if encryptErr != nil {
			fmt.Printf("[提示] %s：部分字段加密失败，已保留明文：%v\n", path, encryptErr)
		}
	}

	if opts.apply {
		fmt.Printf("迁移结束：变更文件 %d 个，加密字段 %d 个。\n", changedFiles, encryptedFields)
		return
	}
	fmt.Printf("预览结束：待迁移文件 %d 个，明文字段 %d 个；未写入任何文件。确认后加 -apply 执行。\n", totalFiles, totalFields)
}

func defaultUsersDir() string {
	if value := strings.TrimSpace(os.Getenv("QIANTIE_SCRIPT_HISTORY_DIR")); value != "" {
		return value
	}
	return "data/users"
}

func findConfigFiles(usersDir string) ([]string, error) {
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(usersDir, entry.Name(), "api-config.json")
		if info, statErr := os.Stat(path); statErr == nil && !info.IsDir() {
			files = append(files, path)
		}
	}
	sort.Strings(files)
	return files, nil
}

func writeAtomic(path string, content []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
