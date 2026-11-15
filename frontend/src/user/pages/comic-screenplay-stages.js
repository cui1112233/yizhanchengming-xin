import {episodeHeadingMatches} from './comic-screenplay-episode-headings.js';
export const comicStages = [
  { key: 'source', label: '剧本' }, { key: 'assets', label: '资产' }, { key: 'storyboard', label: '分镜' },
];
export function comicStageLocation(key) {
  return ({ chapters: { stage: 'source', tool: 'chapters' }, assistant: { stage: 'source', agent: true }, production: { stage: 'storyboard', tool: 'production' }, media: { stage: 'storyboard', tool: 'media' } })[key] || { stage: comicStages.some(item => item.key === key) ? key : 'source' };
}
export function comicScriptEpisodes(content) {
  const text = String(content || '');
  if (!text.trim()) return [];
  const matches = episodeHeadingMatches(text);
  if (!matches.length) return [{ no: 1, title: '第1集', content: text.trim(), raw: true }];
  return matches.map(({match,no}, index) => ({ no, title: (match[2]||'').trim().replace(/^[：:·—\-\s]+/, '') || `第${no}集`, content: text.slice(match.index + match[0].length, matches[index+1]?.match.index ?? text.length).trim(), prefix: index === 0 ? text.slice(0, match.index).trim() : '' }));
}
export function comicEpisodesDocument(episodes) {
  if (episodes.length === 1 && episodes[0].raw) return episodes[0].content;
  return episodes.map((episode, index) => `${episode.prefix ? `${episode.prefix}\n\n` : ''}# 第${episode.no ?? index+1}集${episode.title && !/^第\d+集$/.test(episode.title) ? ` ${episode.title}` : ''}\n\n${episode.content}`).join('\n\n');
}
export function reorderComicEpisodes(episodes, from, to) {
  if (from === to || from < 0 || to < 0 || from >= episodes.length || to >= episodes.length) return episodes;
  const rows = episodes.map(item => ({ ...item, raw: false })), [moved] = rows.splice(from, 1);
  rows.splice(to, 0, moved);
  return rows.map((item, index) => ({ ...item, no: index + 1, prefix: index === 0 ? episodes[0]?.prefix || '' : '' }));
}
export function comicNextState(stage, workspace, dirty = false) {
  if (stage === 'storyboard') return null;
  if (stage === 'assets') return { next: 'storyboard', title: '资产设定待确认', subtitle: '确认后进入分镜', ready: true };
  const ready = comicScriptEpisodes(workspace?.document?.content).some(episode => episode.content.trim()) && !dirty;
  return { next: 'assets', title: ready ? '剧本内容整理完毕' : dirty ? '剧本有未保存修改' : '等待剧本内容', subtitle: ready ? '已确定，可以进入资产' : '可查看资产；解析前请保存剧本', ready };
}

export function positioningImageFingerprint(item, state) {
  const content = {};
  for (const field of ['id','episode','scene_code','scene_codes','scene_name','characters','position_detail','remark','reason']) content[field] = item?.[field];
  return JSON.stringify([content,state?.prefix_text,state?.image_model,state?.show_character_names]);
}
export function mergePositioningRecovery(draft, remote) {
  const serverFields=['image_path','image_media_item_id','status','error','request_key'];
  return { ...draft, items: draft.items.map(item => { const saved=remote.items.find(row=>row.id===item.id);if(!saved)return item;const next={...item};for(const field of serverFields)next[field]=saved[field];return next; }) };
}
export function positioningAttemptState(item, key) {
  if(item?.request_key!==key)return 'unknown';
  return ['succeeded','failed','cancelled','running'].includes(item.status)?item.status:'unknown';
}

export function comicAssetMediaImage(media, assetId) {
  if (!assetId) return null;
  const selectedId=media?.settings?.selected?.[assetId], items=media?.items||[];
  const ready=item=>item.kind==='image'&&['succeeded','ready','completed'].includes(item.status);
  const selected=items.find(item=>item.id===selectedId&&ready(item));
  if(selected)return selected;
  return items.filter(item=>ready(item)&&(item.targetId===assetId||item.target_id===assetId)).map((item,index)=>({item,index,time:Date.parse(item.createdAt||item.created_at||'')||0})).sort((a,b)=>a.time-b.time||a.index-b.index).at(-1)?.item||null;
}

export async function positioningServerRequestKey(purpose, itemId, clientKey, cryptoProvider = globalThis.crypto) {
  const digest=await cryptoProvider.subtle.digest('SHA-256',new TextEncoder().encode(`${purpose}:${itemId}:${clientKey}`));
  return 'positioning-'+Array.from(new Uint8Array(digest)).map(value=>value.toString(16).padStart(2,'0')).join('');
}

export function positioningRefreshDecision(draft, remote, dirty, expectedRevision) {
  if(!dirty)return { draft: remote, revision: remote.revision, conflict: false };
  return { draft: mergePositioningRecovery(draft,remote), revision: expectedRevision, conflict: remote.revision !== expectedRevision };
}
