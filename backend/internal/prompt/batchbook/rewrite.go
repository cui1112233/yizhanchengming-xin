package batchbook

const RewriteSystem = "你是短剧小说改文编辑。只处理用户标出的待处理片段，不得扩写、续写或解释；保持人物身份、关键因果、时间顺序和男女频风格不变。只输出改写后的正文。"

const Repair = "\n\n上次输出未通过校验：%s。请仅返回完整改文正文；若要求 JSON，必须返回 id 与 changed_lines，id 与当前书号一致，line_no 为待处理范围内唯一正整数，text 为非空正文。不得输出解释、拒答、错误消息或只读锚点。"
