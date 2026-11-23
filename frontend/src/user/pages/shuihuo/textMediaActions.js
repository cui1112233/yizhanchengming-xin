import { TEXT_PRODUCTION_SECTIONS } from './textProductionConfig.js';

export const layoutFields = {
  imageAspectRatio:'aspectRatio', textFontId:'fontId', textFontSize:'fontSize', textFontWeight:'fontWeight',
  textLineHeight:'lineHeight', textParagraphSpacing:'paragraphSpacing', textPadding:'padding',
  textAlign:'align', textColor:'textColor', textHighlightColor:'highlightColor', textMaxHeightRatio:'maxHeightRatio',
};
export const pageControls = TEXT_PRODUCTION_SECTIONS.find(s=>s.key==='layout').fields.filter(f=>layoutFields[f.key]);

export function initialPageDraft(doc,page,cover=false) {
  const snapshot=doc?.configSnapshot || {};
  const source=cover ? snapshot.cover : page;
  const layout={};
  for (const f of pageControls) {
    const key=layoutFields[f.key];
    const actual=source?.layout?.[key] ?? snapshot.layout?.[key];
    layout[key]=actual || ((key==='padding'||key==='paragraphSpacing') && actual===0 ? 0 : f.initial);
  }
  layout.targetCharacters=source?.layout?.targetCharacters || snapshot.layout?.targetCharacters || 150;
  const bg=source?.background?.kind ? source.background : snapshot.background;
  return {text:cover ? snapshot.title || '' : page?.text || '', layout,
    background:bg?.kind ? {...bg} : {kind:'color',color:'#111318'}};
}
export function pageChanges(base,draft,expectedRevision,expectedDocumentRevision) {
  const patch={expectedRevision,expectedDocumentRevision};
  if(base.text!==draft.text) patch.text=draft.text;
  if(Object.keys(draft.layout).some(k=>base.layout[k]!==draft.layout[k])) patch.layout=draft.layout;
  if(['kind','color','id'].some(k=>base.background[k]!==draft.background[k])) patch.background=draft.background;
  return patch;
}
export function splitChoices(page,offsets=[]) {
  if(!page)return [];
  const chars=Array.from(page.text);
  return offsets.filter(i=>i>page.startOffset && i<page.endOffset)
    .map(offset=>({offset,label:Array.from(chars.slice(0,offset-page.startOffset).join('').trim()).slice(-35).join('')}))
    .filter(v=>v.label && chars.slice(v.offset-page.startOffset).join('').trim());
}
export function uploadLabel(collection) {
  return collection?.published===true || ['uploaded','published'].includes(collection?.status) ? '再次上传' : '上传网络';
}
