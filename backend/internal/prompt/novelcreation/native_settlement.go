package novelcreation

import promptcatalog "qiantie/backend/internal/prompt"

// CM JSON adapter; original preservation rules are literal settler_prompts#27.
// The complete original RUNTIME_STATE_DELTA protocol is not claimed here.
var IncrementalTruthSettlement = func() promptcatalog.Definition {
	d := TruthSettlement
	d.Version = 2
	d.SystemPrompt += "\nCM数据适配约定：只返回本章增量，JSON格式沿用现有输出合同。已有事实更新沿用id；新增事实使用全书唯一id。未变化的旧事实与伏笔不需要重复返回；空数组表示本章没有新增，不代表清空历史。\n\n\n## 关键规则\n\n1. 状态卡和伏笔池必须基于\"当前追踪文件\"做增量更新，不是从零开始\n2. 正文中的每一个事实性变化都必须反映在对应的追踪文件中\n3. 不要遗漏细节：数值变化、位置变化、关系变化、信息变化都要记录\n4. 角色交互矩阵中的\"信息边界\"要准确——角色只知道他在场时发生的事\n\n## 铁律：只记录正文中实际发生的事（严格执行）\n\n- **只提取正文中明确描写的事件和状态变化**。不要推断、预测、或补充正文没有写到的内容\n- 如果正文只写到角色走到门口还没进去，状态卡就不能写\"角色已进入房间\"\n- 如果正文只暗示了某种可能性但没有确认，不要把它当作已发生的事实记录\n- 不要从卷纲或大纲中补充正文尚未到达的剧情到状态卡\n- 不要删除或修改已有 hooks 中与本章无关的内容——只更新本章正文涉及的 hooks\n- 第 1 章尤其注意：初始追踪文件可能包含从大纲预生成的内容，只保留正文实际支持的部分，不要保留正文未涉及的预设\n- **伏笔例外**：正文中出现的未解疑问、悬念、伏笔线索必须在 hooks 中记录。这不是\"推断\"，而是\"提取正文中的叙事承诺\"。如果正文暗示了一个谜题/冲突/秘密但没有解答，那就是一个 hook，必须记录"
	return d
}()
