package agent

import promptcatalog "qiantie/backend/internal/prompt"

var WorkspaceChat = promptcatalog.Definition{
	ID: "agent.workspace-chat", Version: 1,
	SystemPrompt:   promptcatalog.AgentWorkspaceSystemPrompt,
	RequiredInputs: []string{"user_message", "project_context"},
	OutputContract: "返回面向用户的中文纯文本；不得伪造未执行的动作。",
	Errors:         map[string]string{"missing_message": "用户消息为空", "upstream_failed": "创作助手暂时不可用"},
}

var DirectorPlan = promptcatalog.Definition{
	ID: "agent.director-plan", Version: 1,
	SystemPrompt:   promptcatalog.AgentDirectorPlanSystemPrompt,
	RequiredInputs: []string{"user_goal", "canvas_revision", "selected_context"},
	OutputContract: "只返回一个等待确认的画布修改方案，不声称已经执行。",
	Errors:         map[string]string{"missing_goal": "导演目标为空", "revision_conflict": "画布版本已变化"},
}

var TextNode = promptcatalog.Definition{
	ID: "agent.text-node", Version: 1,
	SystemPrompt:   promptcatalog.AgentTextNodeSystemPrompt,
	RequiredInputs: []string{"prompt", "node_context"},
	OutputContract: "只返回可写入文字节点的正文，不输出流程说明。",
	Errors:         map[string]string{"missing_prompt": "文字生成要求为空", "upstream_failed": "文字生成失败"},
}
