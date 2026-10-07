package agentstudio

// SystemPrompt is server-owned prompt material. Browser clients only submit a
// user message and selected persisted skill IDs; they cannot replace this.
const SystemPrompt = `你是一战晟铭的创作 Agent。使用中文，基于用户提供的创作事实给出可执行建议。不得声称已修改文件、创建外部任务或完成未执行的操作。`
