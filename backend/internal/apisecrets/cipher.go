// Package apisecrets 提供「接口配置」中渠道密钥的落盘加密能力。
//
// 安全默认行为（缺密钥时的处理）：
//  1. 密钥来源优先级：环境变量 QIANTIE_CONFIG_SECRET_KEY > 环境变量
//     QIANTIE_CONFIG_SECRET_KEY_FILE 指向的文件 > 调用方给出的默认本地密钥文件。
//     密钥只从环境变量或本地文件读取，绝不硬编码进仓库。
//  2. 默认本地密钥文件只在「需要写密文」时惰性创建（32 字节随机，权限 0600），
//     只读路径不会创建任何文件。
//  3. 如果既没有环境密钥、也无法创建本地密钥文件，读写层会返回错误；调用方
//     （httpapi.UserConfigStore）按「明确告警 + 保持历史明文行为」降级，保证
//     线上配置仍可保存、已有渠道不会因为本次升级失效。
//  4. 密文格式为 enc:v1:<base64(nonce|ciphertext)>，算法 AES-256-GCM；
//     不带该前缀的值一律按历史明文处理，因此可以兼容读取存量明文，
//     并在下一次保存时自动转成密文。
package apisecrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Prefix 是密文的固定前缀，用于区分历史明文。
	Prefix = "enc:v1:"
	// KeySize 是 AES-256 要求的密钥长度。
	KeySize = 32
	// KeyEnvVar 是首选的密钥来源，值是 base64 或 32 字节原文。
	KeyEnvVar = "QIANTIE_CONFIG_SECRET_KEY"
	// KeyFileEnvVar 指向本地密钥文件，文件内容同样是 base64 或 32 字节原文。
	KeyFileEnvVar = "QIANTIE_CONFIG_SECRET_KEY_FILE"

	// aad 固定附加认证数据，避免同一密钥下的密文被跨用途挪用。
	aad = "qiantie:v88:api-config:secret:v1"
)

var (
	// ErrNoKey 表示当前进程没有可用的加密密钥。
	ErrNoKey = errors.New("未配置接口配置加密密钥")
	// ErrBadCiphertext 表示密文格式损坏、被篡改或密钥不匹配。
	ErrBadCiphertext = errors.New("接口配置密文无效或无法解密")
)

// IsEncrypted 判断值是否为本包产生的密文。
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, Prefix)
}

// Cipher 是 AES-256-GCM 的加解密器。
type Cipher struct {
	key    []byte
	source string
}

// Source 返回密钥来源（环境变量名或文件路径），便于日志说明，绝不包含密钥内容。
func (c *Cipher) Source() string {
	if c == nil {
		return ""
	}
	return c.source
}

// NewCipher 用 32 字节密钥构造加解密器。
func NewCipher(key []byte, source string) (*Cipher, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("接口配置加密密钥长度必须是 %d 字节", KeySize)
	}
	copied := make([]byte, KeySize)
	copy(copied, key)
	return &Cipher{key: copied, source: source}, nil
}

// Provider 负责按优先级解析密钥，本身不缓存，便于测试里切换环境变量。
type Provider struct {
	// KeyFile 是默认本地密钥文件路径；为空表示没有默认位置。
	KeyFile string
}

// ReadCipher 只读取已有密钥，不会创建任何文件。
func (p *Provider) ReadCipher() (*Cipher, error) {
	if p == nil {
		return nil, ErrNoKey
	}
	if key, ok, err := keyFromEnv(); ok || err != nil {
		if err != nil {
			return nil, err
		}
		return NewCipher(key, KeyEnvVar)
	}
	path := p.effectiveKeyFile()
	if path == "" {
		return nil, ErrNoKey
	}
	key, err := readKeyFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoKey
		}
		return nil, err
	}
	return NewCipher(key, path)
}

// WriteCipher 返回用于加密的 Cipher；默认密钥文件不存在时会惰性创建。
func (p *Provider) WriteCipher() (*Cipher, error) {
	if p == nil {
		return nil, ErrNoKey
	}
	if key, ok, err := keyFromEnv(); ok || err != nil {
		if err != nil {
			return nil, err
		}
		return NewCipher(key, KeyEnvVar)
	}
	path := p.effectiveKeyFile()
	if path == "" {
		return nil, ErrNoKey
	}
	key, err := ensureKeyFile(path)
	if err != nil {
		return nil, err
	}
	return NewCipher(key, path)
}

func (p *Provider) effectiveKeyFile() string {
	if custom := strings.TrimSpace(os.Getenv(KeyFileEnvVar)); custom != "" {
		return custom
	}
	return strings.TrimSpace(p.KeyFile)
}

// Encrypt 把明文转成 enc:v1: 前缀的密文。
// 空值和已经是密文的值原样返回，保证重复保存不会二次加密。
func (c *Cipher) Encrypt(plain string) (string, error) {
	if plain == "" || IsEncrypted(plain) {
		return plain, nil
	}
	if c == nil || len(c.key) != KeySize {
		return "", ErrNoKey
	}
	gcm, err := c.gcm()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nil, nonce, []byte(plain), []byte(aad))
	payload := append(nonce, sealed...)
	return Prefix + base64.StdEncoding.EncodeToString(payload), nil
}

// Decrypt 解开密文；不带前缀的值按历史明文原样返回。
func (c *Cipher) Decrypt(value string) (string, error) {
	if !IsEncrypted(value) {
		return value, nil
	}
	if c == nil || len(c.key) != KeySize {
		return "", ErrNoKey
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, Prefix))
	if err != nil {
		return "", ErrBadCiphertext
	}
	gcm, err := c.gcm()
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(payload) <= nonceSize {
		return "", ErrBadCiphertext
	}
	plain, err := gcm.Open(nil, payload[:nonceSize], payload[nonceSize:], []byte(aad))
	if err != nil {
		return "", ErrBadCiphertext
	}
	return string(plain), nil
}

func (c *Cipher) gcm() (cipher.AEAD, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// keyFromEnv 解析环境变量密钥；ok 表示环境变量是否存在。
func keyFromEnv() (key []byte, ok bool, err error) {
	raw := strings.TrimSpace(os.Getenv(KeyEnvVar))
	if raw == "" {
		return nil, false, nil
	}
	key, err = decodeKey(raw)
	if err != nil {
		return nil, true, fmt.Errorf("%s 内容无效：%w", KeyEnvVar, err)
	}
	return key, true, nil
}

func readKeyFile(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	key, err := decodeKey(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("密钥文件 %s 内容无效：%w", path, err)
	}
	return key, nil
}

// ensureKeyFile 读取本地密钥文件；不存在时以 0600 权限生成 32 字节随机密钥。
func ensureKeyFile(path string) ([]byte, error) {
	if key, err := readKeyFile(path); err == nil {
		return key, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(path, []byte(encoded+"\n"), 0o600); err != nil {
		return nil, err
	}
	// Windows 上 WriteFile 的权限位不生效，显式再设一次，其他平台也无副作用。
	_ = os.Chmod(path, 0o600)
	return key, nil
}

// decodeKey 支持 base64（标准）和 32 字节原文两种写法。
func decodeKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, errors.New("密钥为空")
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == KeySize {
		return decoded, nil
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(decoded) == KeySize {
		return decoded, nil
	}
	if len(raw) == KeySize {
		return []byte(raw), nil
	}
	return nil, fmt.Errorf("密钥必须是 %d 字节原文或对应的 base64", KeySize)
}

// GenerateKeyBase64 生成一把新密钥的 base64 文本，供部署时写入环境变量或密钥文件。
// 迁移命令的 -print-env 提示会用到它。
func GenerateKeyBase64() (string, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
