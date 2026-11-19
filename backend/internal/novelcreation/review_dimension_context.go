package novelcreation

// The dimension selector is useful only when the reviewer and reviser receive
// the frozen reference material needed to check it. Never load fresh material
// here: queued runs must keep the context they were created with.
func withReviewDimensionContext(input RunContext, payload map[string]any, foundation foundationPlanOutput) map[string]any {
	if input.Agent == nil || input.Agent.ReviewDimensionIDs == nil {
		return payload // Preserve the exact payload of legacy journaled requests.
	}
	payload["creation_mode"] = modeContract(input.Session.Mode)
	payload["mode_inputs"] = input.ModeInputs
	payload["style_profile"] = input.StyleProfile
	payload["rules"] = input.Rules
	payload["notes"] = input.Notes
	payload["knowledge_materials"] = input.KnowledgeMaterials
	payload["foundation_plan"] = foundation
	return payload
}
