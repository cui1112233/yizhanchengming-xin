const AUTOMATIC_PATTERNS = [/全自动化/, /全自动(?:完成|执行|生成|跑)/, /自动执行到底/, /无需确认.*(?:完成|执行)/];
const SUPERVISED_PATTERNS = [/逐步确认/, /每一步.*确认/, /先(?:给我)?看看(?:方案|结果)/, /确认后再/, /不要自动/];

export function normalizeDirectorMode(value) {
  return value === 'automatic' ? 'automatic' : 'supervised';
}

export function detectDirectorMode(input, currentMode = 'supervised') {
  const text = String(input || '').trim();
  if (SUPERVISED_PATTERNS.some(pattern => pattern.test(text))) return 'supervised';
  if (AUTOMATIC_PATTERNS.some(pattern => pattern.test(text))) return 'automatic';
  return normalizeDirectorMode(currentMode);
}

export function summarizeDirectorPlan(plan = {}) {
  const commands = Array.isArray(plan.commands) ? plan.commands : [];
  return {
    commandCount: commands.length,
    destructiveCount: commands.filter(command => command?.type === 'remove_nodes').length,
    executionCount: commands.filter(command => ['start_generation', 'apply_result'].includes(command?.type)).length
  };
}
