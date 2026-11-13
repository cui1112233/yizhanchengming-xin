export function buildRuleConfigPatch(kind, rules, data = {}) {
  const app = data.appConfig || data.config || {};
  const knowledge = { ...(app.knowledge || {}), ...(data.knowledgeLibraries || {}) };
  const key = kind.startsWith('chapter_') ? 'chapter_rules' : kind === 'pair_fill_rules' ? 'symbol_rules' : kind;
  const merged = { ...(knowledge[key] || {}), ...rules };
  const patch = { knowledge: { [key]: merged } };
  if (kind === 'layout_rules') {
    patch.layout = { ...(app.layout || {}), ...Object.fromEntries(Object.entries(rules).filter(([key]) => key !== 'indent_spaces' && key !== 'blank_lines_before')) };
  }
  return patch;
}
