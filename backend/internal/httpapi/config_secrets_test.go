package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qiantie/backend/internal/apisecrets"
)

func useTempConfigSecretKey(t *testing.T) string {
	t.Helper()
	keyFile := filepath.Join(t.TempDir(), "config-secret.key")
	t.Setenv(apisecrets.KeyEnvVar, "")
	t.Setenv(apisecrets.KeyFileEnvVar, keyFile)
	return keyFile
}

func writeRawConfigFile(t *testing.T, baseDir, username, body string) string {
	t.Helper()
	dir := filepath.Join(baseDir, username)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "api-config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestUserConfigStoreEncryptsSecretsAtRest(t *testing.T) {
	useTempConfigSecretKey(t)
	store := NewUserConfigStore(t.TempDir())
	config := map[string]any{
		"apiIntegrations": map[string]any{
			"values": map[string]any{
				"openrouterApiKey": "sk-plain-openrouter",
				"ttsServiceUrl":    "http://tts.example/v1/audio/speech",
			},
			"customChannels": []any{
				map[string]any{"id": "ch_1", "name": "自建渠道", "apiKey": "sk-plain-custom"},
			},
		},
		"image": map[string]any{"apiKey": "sk-plain-image"},
		"video": map[string]any{"ydApiKey": "sk-plain-video"},
	}
	if err := store.write("alice", config); err != nil {
		t.Fatalf("write: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(store.BaseDir, "alice", "api-config.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := string(raw)
	for _, plaintext := range []string{"sk-plain-openrouter", "sk-plain-custom", "sk-plain-image", "sk-plain-video"} {
		if strings.Contains(text, plaintext) {
			t.Fatalf("plaintext %q must not be persisted", plaintext)
		}
	}
	if !strings.Contains(text, apisecrets.Prefix) {
		t.Fatal("persisted config should contain encrypted values")
	}
	if !strings.Contains(text, "http://tts.example/v1/audio/speech") {
		t.Fatal("non-secret fields should stay readable in storage")
	}

	// 写入不能污染调用方内存：handler 后续还会用同一份配置做响应。
	stored := apiIntegrationValues(apiIntegrationStoredMap(config))
	if stored["openrouterApiKey"] != "sk-plain-openrouter" {
		t.Fatal("write must not mutate caller's plaintext config")
	}

	// 读取时透明还原成明文，生成链路无感。
	reloaded := store.read("alice")
	values := apiIntegrationValues(apiIntegrationStoredMap(reloaded))
	if values["openrouterApiKey"] != "sk-plain-openrouter" {
		t.Fatalf("secret should be decrypted on read, got %v", values["openrouterApiKey"])
	}
	channels := apiIntegrationChannels(apiIntegrationStoredMap(reloaded))
	if len(channels) != 1 || channels[0].(map[string]any)["apiKey"] != "sk-plain-custom" {
		t.Fatal("custom channel key should be decrypted on read")
	}
	if image, _ := reloaded["image"].(map[string]any); image["apiKey"] != "sk-plain-image" {
		t.Fatal("legacy image key should be decrypted on read")
	}
}

func TestUserConfigStoreReadsLegacyPlaintextAndUpgradesOnSave(t *testing.T) {
	useTempConfigSecretKey(t)
	baseDir := t.TempDir()
	path := writeRawConfigFile(t, baseDir, "bob", `{
  "apiIntegrations": {
    "values": {"openrouterApiKey": "sk-legacy-plain"},
    "customChannels": [{"id": "ch_1", "name": "旧渠道", "apiKey": "sk-legacy-custom"}]
  }
}`)
	store := NewUserConfigStore(baseDir)

	reloaded := store.read("bob")
	values := apiIntegrationValues(apiIntegrationStoredMap(reloaded))
	if values["openrouterApiKey"] != "sk-legacy-plain" {
		t.Fatalf("legacy plaintext must stay usable, got %v", values["openrouterApiKey"])
	}

	// 任意一次正常保存都应把存量明文升级成密文。
	if err := store.write("bob", reloaded); err != nil {
		t.Fatalf("write: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(raw), "sk-legacy-plain") || strings.Contains(string(raw), "sk-legacy-custom") {
		t.Fatal("legacy plaintext should be encrypted after the next save")
	}
	if !strings.Contains(string(raw), apisecrets.Prefix) {
		t.Fatal("saved config should contain ciphertext")
	}
	again := apiIntegrationValues(apiIntegrationStoredMap(store.read("bob")))
	if again["openrouterApiKey"] != "sk-legacy-plain" {
		t.Fatal("upgraded config must still decrypt to the original value")
	}
}

func TestUserConfigStoreWritesPlaintextWhenKeyUnavailable(t *testing.T) {
	// 用「普通文件当作目录」制造无法创建密钥文件的场景，验证明确降级而非写入失败。
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	t.Setenv(apisecrets.KeyEnvVar, "")
	t.Setenv(apisecrets.KeyFileEnvVar, filepath.Join(blocker, "system", "config-secret.key"))

	store := NewUserConfigStore(t.TempDir())
	config := map[string]any{
		"apiIntegrations": map[string]any{"values": map[string]any{"openrouterApiKey": "sk-no-key"}},
	}
	if err := store.write("carol", config); err != nil {
		t.Fatalf("write should degrade to plaintext instead of failing: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(store.BaseDir, "carol", "api-config.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(raw), "sk-no-key") {
		t.Fatal("without a usable key the legacy plaintext behavior must be kept")
	}
}

func TestPublicUserConfigNeverLeaksIntegrationSecrets(t *testing.T) {
	config := map[string]any{
		"apiIntegrations": map[string]any{
			"values": map[string]any{
				"openrouterApiKey": "sk-leak-openrouter",
				"ttsServiceUrl":    "http://tts.example/v1",
			},
			"customChannels": []any{
				map[string]any{"id": "ch_1", "name": "自建", "apiKey": "sk-leak-custom"},
			},
		},
	}
	public := publicUserConfig(config)
	stored, ok := public[apiIntegrationsSettingKey].(map[string]any)
	if !ok {
		t.Fatalf("public config should still carry %s view", apiIntegrationsSettingKey)
	}
	if _, leaked := stored["values"].(map[string]any)["openrouterApiKey"]; leaked {
		t.Fatal("public config must not expose integration secret values")
	}
	channel := stored["customChannels"].([]map[string]any)[0]
	if _, leaked := channel["apiKey"]; leaked {
		t.Fatal("public config must not expose custom channel keys")
	}
	if configured, _ := stored["configured"].(map[string]bool); !configured["openrouterApiKey"] {
		t.Fatal("secret should still be reported as configured")
	}
}

func TestTransformConfigSecretsCoversEveryWhitelistedSecretField(t *testing.T) {
	values := map[string]any{}
	secrets := map[string]string{}
	plainFields := map[string]string{}
	for key, secret := range apiIntegrationFields {
		if secret {
			value := "plain-" + key
			values[key] = value
			secrets[key] = value
			continue
		}
		plainFields[key] = "keep-" + key
		values[key] = plainFields[key]
	}
	document := map[string]any{
		apiIntegrationsSettingKey: map[string]any{"values": values},
	}
	changed := transformConfigSecrets(document, func(value string) string {
		return "enc(" + value + ")"
	})
	if changed != len(secrets) {
		t.Fatalf("expected %d secret fields transformed, got %d", len(secrets), changed)
	}
	for key, value := range secrets {
		if values[key] != "enc("+value+")" {
			t.Fatalf("secret field %s was not transformed", key)
		}
	}
	for key, value := range plainFields {
		if values[key] != value {
			t.Fatalf("non-secret field %s must not be transformed", key)
		}
	}
}

func TestEncryptConfigSecretsForMigrationIsIdempotent(t *testing.T) {
	key := make([]byte, apisecrets.KeySize)
	for i := range key {
		key[i] = byte(i + 7)
	}
	cipher, err := apisecrets.NewCipher(key, "test")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	document := map[string]any{
		apiIntegrationsSettingKey: map[string]any{
			"values": map[string]any{
				"openrouterApiKey": "sk-migrate",
				"ttsServiceUrl":    "http://tts.example/v1",
			},
			"customChannels": []any{
				map[string]any{"id": "ch_1", "name": "自建", "apiKey": "sk-migrate-custom"},
			},
		},
	}
	pending := CountPlaintextConfigSecrets(document)
	if pending != 2 {
		t.Fatalf("expected 2 plaintext secrets, got %d", pending)
	}
	changed, err := EncryptConfigSecretsForMigration(document, cipher)
	if err != nil || changed != 2 {
		t.Fatalf("first migration should encrypt 2 values, got %d err=%v", changed, err)
	}
	again, err := EncryptConfigSecretsForMigration(document, cipher)
	if err != nil || again != 0 {
		t.Fatalf("second migration must be a no-op, got %d err=%v", again, err)
	}
	if CountPlaintextConfigSecrets(document) != 0 {
		t.Fatal("no plaintext secrets should remain")
	}
	values := apiIntegrationValues(apiIntegrationStoredMap(document))
	plain, err := cipher.Decrypt(values["openrouterApiKey"].(string))
	if err != nil || plain != "sk-migrate" {
		t.Fatalf("migrated value should decrypt: %q err=%v", plain, err)
	}
	if values["ttsServiceUrl"] != "http://tts.example/v1" {
		t.Fatal("non-secret field must stay untouched by migration")
	}
}

func TestUserConfigStoreEncryptedFileRoundTripsThroughLegacyReaders(t *testing.T) {
	useTempConfigSecretKey(t)
	store := NewUserConfigStore(t.TempDir())
	config := map[string]any{
		"apiIntegrations": map[string]any{"values": map[string]any{"openrouterApiKey": "sk-reader"}},
		"modelCatalog": []any{
			map[string]any{"id": "m1", "credential": "sk-catalog"},
		},
	}
	if err := store.write("dave", config); err != nil {
		t.Fatalf("write: %v", err)
	}
	// 兼容性：app.go / novelfetchworkshop 会直接读取该文件，但只取 modelCatalog。
	raw, err := os.ReadFile(filepath.Join(store.BaseDir, "dave", "api-config.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var direct map[string]any
	if err := json.Unmarshal(raw, &direct); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	catalog, _ := direct["modelCatalog"].([]any)
	if len(catalog) != 1 || catalog[0].(map[string]any)["credential"] != "sk-catalog" {
		t.Fatal("modelCatalog credentials stay outside this change and must not be altered")
	}
}
