package httpapi

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"qiantie/backend/internal/apisecrets"
)

// configSecretProvider 是 UserConfigStore 的密钥加解密入口。
//
// 默认密钥文件放在用户目录的上一级（生产即 data/system/config-secret.key），
// 可用 QIANTIE_CONFIG_SECRET_KEY / QIANTIE_CONFIG_SECRET_KEY_FILE 覆盖。
// 只有在需要写密文时才创建默认密钥文件，读取路径不会产生任何磁盘写入。
type configSecretProvider struct {
	secrets *apisecrets.Provider
	keyFile string
	mu      sync.Mutex
	warned  map[string]bool
}

func newConfigSecretProvider(baseDir string) *configSecretProvider {
	keyFile := ""
	if trimmed := strings.TrimSpace(baseDir); trimmed != "" {
		keyFile = filepath.Join(filepath.Dir(filepath.Clean(trimmed)), "system", "config-secret.key")
	}
	return &configSecretProvider{
		secrets: &apisecrets.Provider{KeyFile: keyFile},
		keyFile: keyFile,
		warned:  map[string]bool{},
	}
}

// decrypt 把配置里的密文还原成明文；兼容历史明文。
// 没有密钥或密钥不匹配时保留原值（避免密文被明文覆盖后无法恢复），并告警一次。
func (p *configSecretProvider) decrypt(config map[string]any) {
	if p == nil || config == nil {
		return
	}
	cipher, err := p.secrets.ReadCipher()
	if err != nil || cipher == nil {
		if encryptedConfigSecretCount(config) > 0 {
			p.warnOnce("read-no-key", "接口配置中存在加密的渠道 Key，但当前没有可用密钥（"+p.keyHint()+"），已保持密文返回；请配置密钥后重试")
		}
		return
	}
	transformConfigSecrets(config, func(value string) string {
		if !apisecrets.IsEncrypted(value) {
			return value
		}
		plain, err := cipher.Decrypt(value)
		if err != nil {
			p.warnOnce("read-bad-cipher", "接口配置中的渠道 Key 无法解密（"+p.keyHint()+"），已保留密文以避免数据丢失")
			return value
		}
		return plain
	})
}

// encrypt 在落盘前把明文密钥转成密文；已加密的值不会二次加密。
// 没有密钥时按「明确告警 + 保持历史明文行为」降级，避免线上配置无法保存。
func (p *configSecretProvider) encrypt(config map[string]any) {
	if p == nil || config == nil {
		return
	}
	cipher, err := p.secrets.WriteCipher()
	if err != nil || cipher == nil {
		p.warnOnce("write-no-key", "接口配置加密密钥不可用（"+p.keyHint()+"）：渠道 Key 将以明文写入，请设置 "+apisecrets.KeyEnvVar+" 或确保默认密钥文件可写")
		return
	}
	// 只在进程内首次写入时提示一次密钥来源（不含密钥内容），便于确认密钥文件位置。
	p.warnOnce("write-encrypt-active", "渠道 Key 已启用落盘加密（"+cipher.Source()+"）")
	transformConfigSecrets(config, func(value string) string {
		if value == "" || apisecrets.IsEncrypted(value) {
			return value
		}
		encrypted, err := cipher.Encrypt(value)
		if err != nil {
			p.warnOnce("write-encrypt-failed", "接口配置渠道 Key 加密失败，本次以明文写入："+err.Error())
			return value
		}
		return encrypted
	})
}

func (p *configSecretProvider) warnOnce(key, message string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.warned == nil {
		p.warned = map[string]bool{}
	}
	if p.warned[key] {
		return
	}
	p.warned[key] = true
	log.Printf("[api-secrets] %s", message)
}

// keyHint 只说明密钥来源，不包含任何密钥内容。
func (p *configSecretProvider) keyHint() string {
	if strings.TrimSpace(os.Getenv(apisecrets.KeyEnvVar)) != "" {
		return "来源：环境变量 " + apisecrets.KeyEnvVar
	}
	if custom := strings.TrimSpace(os.Getenv(apisecrets.KeyFileEnvVar)); custom != "" {
		return "来源：密钥文件 " + custom
	}
	if p != nil && p.keyFile != "" {
		return "来源：默认密钥文件 " + p.keyFile
	}
	return "未配置密钥来源"
}

// transformConfigSecrets 对 api-config.json 里所有渠道密钥字段应用 fn。
// 字段范围与 apiIntegrationFields 中标记为密钥的键保持一致（有测试守护）。
func transformConfigSecrets(config map[string]any, fn func(string) string) int {
	if config == nil || fn == nil {
		return 0
	}
	changed := 0
	apply := func(container map[string]any, key string) {
		if container == nil {
			return
		}
		value, ok := container[key].(string)
		if !ok {
			return
		}
		if next := fn(value); next != value {
			container[key] = next
			changed++
		}
	}

	apply(config, "apiKey")
	if image, ok := config["image"].(map[string]any); ok {
		apply(image, "apiKey")
	}
	if video, ok := config["video"].(map[string]any); ok {
		apply(video, "apiKey")
		apply(video, "ydApiKey")
		apply(video, "h3ApiKey")
	}
	stored, ok := config[apiIntegrationsSettingKey].(map[string]any)
	if !ok {
		return changed
	}
	if values, ok := stored["values"].(map[string]any); ok {
		for key, secret := range apiIntegrationFields {
			if !secret {
				continue
			}
			apply(values, key)
		}
	}
	eachIntegrationChannelMap(stored, func(channel map[string]any) {
		apply(channel, "apiKey")
	})
	return changed
}

func eachIntegrationChannelMap(stored map[string]any, fn func(map[string]any)) {
	switch list := stored["customChannels"].(type) {
	case []any:
		for _, item := range list {
			if channel, ok := item.(map[string]any); ok {
				fn(channel)
			}
		}
	case []map[string]any:
		for _, channel := range list {
			fn(channel)
		}
	}
}

func encryptedConfigSecretCount(config map[string]any) int {
	count := 0
	transformConfigSecrets(config, func(value string) string {
		if apisecrets.IsEncrypted(value) {
			count++
		}
		return value
	})
	return count
}

// deepCloneConfigValue 深拷贝 JSON 结构，保证写入加密不会污染调用方持有的配置。
func deepCloneConfigValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = deepCloneConfigValue(item)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = deepCloneConfigValue(item)
		}
		return result
	case []map[string]any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = deepCloneConfigValue(item)
		}
		return result
	default:
		return value
	}
}

// CountPlaintextConfigSecrets 统计文档中仍需迁移的明文密钥数量，供迁移命令预览。
// 只读：不会修改传入的文档，也不会落盘。
func CountPlaintextConfigSecrets(document map[string]any) int {
	count := 0
	transformConfigSecrets(document, func(value string) string {
		if value != "" && !apisecrets.IsEncrypted(value) {
			count++
		}
		return value
	})
	return count
}

// EncryptConfigSecretsForMigration 供迁移命令复用与线上写入完全相同的字段口径。
// 只处理明文值，已带 enc:v1: 前缀的值会跳过，因此可以重复执行（幂等）。
// 本函数只修改传入的 document，不落盘。
func EncryptConfigSecretsForMigration(document map[string]any, cipher *apisecrets.Cipher) (int, error) {
	if cipher == nil {
		return 0, apisecrets.ErrNoKey
	}
	var firstErr error
	changed := transformConfigSecrets(document, func(value string) string {
		if value == "" || apisecrets.IsEncrypted(value) {
			return value
		}
		encrypted, err := cipher.Encrypt(value)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return value
		}
		return encrypted
	})
	return changed, firstErr
}
