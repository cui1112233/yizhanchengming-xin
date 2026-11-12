import test from 'node:test';
import assert from 'node:assert/strict';
import { buildAICutPreviewFrame } from '../aicutPreview.js';
import { catalogTextSettings, gradientGlyphColor, catalogPreviewPasses } from './catalogPreview.js';
import { editorCatalog } from './editorCatalog.js';

test('effect tracks reach each active visual only within the effect window', () => {
 const timeline = { tracks: [{ id:'video',kind:'video',clips:[{id:'clip',kind:'video',assetId:'asset',startMs:0,durationMs:10000}] },{id:'fx',kind:'effect',clips:[{id:'effect',kind:'effect',resourceId:'effect-fisheye',startMs:4000,durationMs:1000,parameters:{intensity:.8}}]}] };
 const assets=[{id:'asset',kind:'video'}];
 assert.equal(buildAICutPreviewFrame(timeline,assets,3999).visuals[0].layerEffects.length,0);
 assert.equal(buildAICutPreviewFrame(timeline,assets,4500).visuals[0].layerEffects[0].resourceId,'effect-fisheye');
 assert.equal(buildAICutPreviewFrame(timeline,assets,5000).visuals[0].layerEffects.length,0);
 timeline.tracks[1].muted=true;
 assert.equal(buildAICutPreviewFrame(timeline,assets,4500).visuals[0].layerEffects.length,0);
});
test('resource text parameters retain zero opacity and selected font size', () => {
 const settings=catalogTextSettings('text-bold-title',{fontSize:120,opacity:0});
 assert.equal(settings.fontSize,120);assert.equal(settings.textOpacity,0);
});
test('gradient has the same three deterministic glyph colors as export', () => {
 assert.equal(gradientGlyphColor(0,3),'#ffd66b');assert.equal(gradientGlyphColor(1,3),'#b78bff');assert.equal(gradientGlyphColor(2,3),'#63dcff');
});
test('every published filter and effect has an actual pass and honors zero intensity',()=>{
 for(const resource of editorCatalog.filter(item=>['filter','effect'].includes(item.category))){
  if(resource.id!=='filter-original')assert.ok(catalogPreviewPasses(resource.id,{intensity:.5}).length,resource.id);
  assert.deepEqual(catalogPreviewPasses(resource.id,{intensity:0}),[],resource.id+' zero intensity');
 }
});
