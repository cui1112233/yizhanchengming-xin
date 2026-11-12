export const agentToolNames = {
 analyze_audio_silence: '分析真实音频静音', remove_audio_silence: '按音频删除静音',
 list_asr_models: '查询可用字幕模型', get_asr_jobs: '查询字幕任务状态',
 get_project_snapshot: '读取工程', get_transcript: '读取字幕和剧本', place_asset: '放入素材', add_bgm: '添加背景音乐', move_clip: '移动片段', delete_clip: '删除片段', split_clip: '切分片段', set_playback_rate: '调整速度', replace_subtitle_text: '修正字幕', transcribe_asset: '提交字幕识别', smart_split_asset: '智能切镜', extract_audio: '分离音频', clean_oral_pauses: '清理口误停顿', undo_timeline: '撤销', redo_timeline: '重做', select_caption_ranges: '按字幕选片', compact_caption_gaps: '压缩句间停顿', set_clip_animation: '设置片段动画', set_clip_color: '调整画面色彩', apply_edit_operations: '应用剪辑操作'
};
export function agentFollowupFromMessages(messages = []) {
 const last = messages.filter(item => item.role === 'assistant' || item.role === 'user').at(-1);
 if (last?.role !== 'assistant') return null;
 const offset = last.content?.lastIndexOf('\n{');
 if (offset < 0 || offset === undefined) return null;
 try { const value = JSON.parse(last.content.slice(offset + 1)); return value.question && Array.isArray(value.fields) ? value : null; } catch { return null; }
}
export function formatAgentToolMessage(content) {
 try { const value = JSON.parse(content); return `${agentToolNames[value.name] || value.name || '工具'}：${value.message || (value.ok ? '已完成' : '未完成')}`; } catch { return content; }
}
export function agentStepRows(events = []) {
 const rows = new Map();
 for (const event of events) {
  if (!event.type?.startsWith('tool_')) continue;
  const key = event.toolCallId || event.sequence;
  const old = rows.get(key);
  let args = old?.arguments;
  if (event.type === 'tool_start') { try { args = JSON.parse(event.message).arguments; } catch {} }
  rows.set(key, { ...old, ...event, arguments: args, title: agentToolNames[event.toolName] || event.toolName, state: event.type === 'tool_done' ? 'done' : event.type === 'tool_failed' ? 'failed' : 'running' });
 }
 return [...rows.values()];
}
