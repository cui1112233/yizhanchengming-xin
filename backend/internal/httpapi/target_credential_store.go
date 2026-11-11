package httpapi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type target121Credential struct {
	TargetUsername string `json:"targetUsername"`
	Password       string `json:"password"`
	BaseURL        string `json:"baseUrl"`
}

type encrypted121Credential struct {
	Version    int    `json:"version"`
	Algorithm  string `json:"algorithm"`
	IV         string `json:"iv"`
	Tag        string `json:"tag"`
	Ciphertext string `json:"ciphertext"`
}

func load121Credential(usersDir, owner, secret string) (target121Credential, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || filepath.Base(owner) != owner || owner == "." || owner == ".." {
		return target121Credential{}, fmt.Errorf("账号目录无效")
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return target121Credential{}, fmt.Errorf("自动登录密钥未配置")
	}
	raw, err := os.ReadFile(filepath.Join(usersDir, owner, "novel-fetch-workshop", "121-credentials.json"))
	if err != nil {
		return target121Credential{}, fmt.Errorf("未找到已保存的视频管理系统账号密码")
	}
	var record encrypted121Credential
	if json.Unmarshal(raw, &record) != nil || record.Version != 1 || record.Algorithm != "aes-256-gcm" {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码格式无效")
	}
	iv, err := base64.StdEncoding.DecodeString(record.IV)
	if err != nil {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码无法解密")
	}
	tag, err := base64.StdEncoding.DecodeString(record.Tag)
	if err != nil {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码无法解密")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(record.Ciphertext)
	if err != nil {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码无法解密")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码无法解密")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(iv) != gcm.NonceSize() || len(tag) != gcm.Overhead() {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码无法解密")
	}
	plaintext, err := gcm.Open(nil, iv, append(ciphertext, tag...), nil)
	if err != nil {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码无法解密")
	}
	var credential target121Credential
	if json.Unmarshal(plaintext, &credential) != nil || strings.TrimSpace(credential.TargetUsername) == "" || credential.Password == "" {
		return target121Credential{}, fmt.Errorf("已保存的视频管理系统账号密码不完整")
	}
	return credential, nil
}
