package main

import "testing"

func setRequiredWorkerEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MYSQL_DSN", "user:pass@tcp(mysql:3306)/app")
	t.Setenv("REDIS_URL", "redis://redis:6379/0")
	t.Setenv("PIPELINE_QUEUE_KEY", "")
	t.Setenv("AGENT_TOOL_QUEUE_KEY", "")
	t.Setenv("121_FETCH_ENDPOINT", "")
	t.Setenv("AI_BASE_URL", "")
	t.Setenv("AI_MODEL", "")
	t.Setenv("AI_API_KEY", "")
	t.Setenv("TOS_ENDPOINT", "")
	t.Setenv("TOS_REGION", "")
	t.Setenv("TOS_ACCESS_KEY", "")
	t.Setenv("TOS_SECRET_KEY", "")
	t.Setenv("TOS_BUCKET", "")
	t.Setenv("PROVIDER_CREDENTIAL_KEY", "")
}

func TestConfigFromEnvRequiresMySQLAndRedisAndDefaults(t *testing.T) {
	setRequiredWorkerEnv(t)
	cfg, err := configFromEnv()
	if err != nil { t.Fatal(err) }
	if cfg.QueueKey != "qiantie:pipeline:ready" { t.Fatalf("queue=%q",cfg.QueueKey) }
	if cfg.AgentToolQueueKey != "qiantie:agent:tools" { t.Fatalf("agent queue=%q",cfg.AgentToolQueueKey) }
	if cfg.FetchEndpoint != "https://txt.121w.com/api.php" { t.Fatalf("endpoint=%q",cfg.FetchEndpoint) }
	if cfg.AIBaseURL != "" || cfg.AIModel != "" || cfg.AIAPIKey != "" { t.Fatalf("AI should remain optional when completely unset: %+v",cfg) }
}

func TestConfigFromEnvAcceptsCompleteAIConfig(t *testing.T) {
	setRequiredWorkerEnv(t)
	t.Setenv("AI_BASE_URL","https://example.ai/v1")
	t.Setenv("AI_MODEL","text-model")
	t.Setenv("AI_API_KEY","secret")
	cfg,err:=configFromEnv()
	if err!=nil{t.Fatal(err)}
	if cfg.AIBaseURL!="https://example.ai/v1"||cfg.AIModel!="text-model"||cfg.AIAPIKey!="secret"{t.Fatalf("unexpected AI config: %+v",cfg)}
}

func TestConfigFromEnvRejectsPartialAIConfig(t *testing.T) {
	setRequiredWorkerEnv(t)
	t.Setenv("AI_BASE_URL","https://example.ai/v1")
	if _,err:=configFromEnv();err==nil{t.Fatal("expected partial AI config error")}
}

func TestConfigFromEnvAcceptsAgentVideoRuntime(t *testing.T) {
	setRequiredWorkerEnv(t)
	t.Setenv("TOS_ENDPOINT","tos.example.com")
	t.Setenv("TOS_REGION","cn-guangzhou")
	t.Setenv("TOS_ACCESS_KEY","ak")
	t.Setenv("TOS_SECRET_KEY","sk")
	t.Setenv("TOS_BUCKET","media")
	t.Setenv("PROVIDER_CREDENTIAL_KEY","12345678901234567890123456789012")
	cfg,err:=configFromEnv()
	if err!=nil{t.Fatal(err)}
	if cfg.TOSBucket!="media"||len([]byte(cfg.ProviderCredentialKey))!=32{t.Fatalf("cfg=%#v",cfg)}
}

func TestConfigFromEnvRejectsInvalidProviderCredentialKey(t *testing.T) {
	setRequiredWorkerEnv(t)
	t.Setenv("PROVIDER_CREDENTIAL_KEY","short")
	if _,err:=configFromEnv();err==nil{t.Fatal("expected provider credential key error")}
}
