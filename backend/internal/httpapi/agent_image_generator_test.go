package httpapi

import "testing"

func TestAgentLocalDoubaoImageUsesSelectedModel(t *testing.T) {
	input := AgentImageGenerationInput{Provider: "doubao_local_executor", ModelID: "Seedream 5.0 Pro"}
	model, ok := agentLocalDoubaoImageModel(input)
	if !ok || model != "Seedream 5.0 Pro" {
		t.Fatalf("model=%q local=%v", model, ok)
	}
}

func TestAgentConfiguredImageDoesNotEnterLocalExecutor(t *testing.T) {
	input := AgentImageGenerationInput{Provider: "personal_api", ModelID: "configured-image"}
	if model, ok := agentLocalDoubaoImageModel(input); ok || model != "" {
		t.Fatalf("model=%q local=%v", model, ok)
	}
}
