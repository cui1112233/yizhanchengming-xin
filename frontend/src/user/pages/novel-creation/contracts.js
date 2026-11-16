export const MODES = [
  ['long_original', '长篇原创'], ['short_original', '短篇原创'], ['long_imitation', '仿写长篇'],
  ['short_imitation', '仿写短篇'], ['fanfiction', '番外创作'], ['continuation', '续写创作']
];
export const MODE_FIELDS = {
  long_imitation: ['referenceText', '参考文本'], short_imitation: ['referenceText', '参考文本'],
  fanfiction: ['existingSetting', '既有人物与世界规则'], continuation: ['previousText', '被续写的前文']
};
export const PHASES = [
  ['foundation-plan', '基础设定'], ['chapter-plan', '章节计划'], ['chapter-draft', '正文写作'],
  ['length-normalize', '长度整理'], ['post-write', '正文后处理'], ['ai-tells', '表达检查'],
  ['chapter-audit', '章节审核'], ['chapter-revision', '章节修订'], ['truth-settlement', '事实结算'], ['memory-index', '记忆索引']
];
export const THINKING = [{ value: 'fast', label: '快速' }, { value: 'standard', label: '标准思考' }, { value: 'deep', label: '深度思考' }];
export const TERMINAL = new Set(['succeeded', 'succeeded_with_warnings', 'failed', 'cancelled']);
export const STATUS = {
  queued: '排队中', planning: '生成创作蓝图', writing: '写作正文', normalizing: '整理正文', auditing: '审核章节', revising: '修订章节', settling: '结算人物与伏笔', indexing: '建立记忆索引', analyzing: '正在分析资料',
  awaiting_confirmation: '等待确认', paused: '已暂停', succeeded: '创作完成', succeeded_with_warnings: '完成，有审核提醒', failed: '创作失败', cancelled: '已停止',
  ready: '等待恢复执行', leased: '正在执行', save_pending: '结果保存待恢复', done: '已保存', running: '运行中', waiting: '等待中', stopped: '已停止', completed: '已完成', not_started: '未启动',
  awaiting_model_confirmation: '请确认后继续下一次模型调用', busy: '等待作品空闲', daily_cap: '今日已达章节上限', awaiting_blueprint_confirmation: '请确认创作蓝图后继续'
};
export const parseJSON = value => { if (typeof value === 'object') return value || {}; try { return JSON.parse(value || '{}'); } catch { return {}; } };
export const draftKey = (owner, project, session, mode) => `novel:draft:v2:${[owner, project || '_', session || '_', mode].map(encodeURIComponent).join(':')}`;
export function readSaved(key, fallback = {}) { try { return JSON.parse(localStorage.getItem(key)) ?? fallback; } catch { return fallback; } }
export function mergeSnapshot(previous, next) {
  const events = new Map((previous?.run?.id === next?.run?.id ? previous?.events || [] : []).map(e => [e.eventId, e]));
  for (const e of next?.events || []) events.set(e.eventId, e);
  return { ...previous, ...next, events: [...events.values()].sort((a, b) => a.eventId - b.eventId) };
}
export function modelOptions(models, selected) {
  return selected && !models.some(m => m.value === selected) ? [{ value: selected, label: `原模型已不可用（${selected}），请重新选择`, disabled: true }, ...models] : models;
}
export const availableModel = (models, id) => !!id && models.some(m => m.value === id);
export function downloadBlob(blob, filename) { const url = URL.createObjectURL(blob); const a = document.createElement('a'); a.href = url; a.download = filename; document.body.appendChild(a); a.click(); a.remove(); URL.revokeObjectURL(url); }
