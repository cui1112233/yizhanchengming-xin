import test from 'node:test';
import assert from 'node:assert/strict';
import {TEXT_LAYOUT_TEMPLATES,publishOverrideRestores,textColorScheme,hasLowTextContrast,readableTextColor,lightTextChanges} from './textLayoutTemplates.js';
import {applyTextProductionChanges,textProductionBookOverride} from './textProductionConfig.js';
test('light text enables dialogue highlighting and retains chosen rules',()=>{
 assert.equal(lightTextChanges('off').textHighlightRules,'dialogue');
 assert.equal(lightTextChanges('').textHighlightRules,'dialogue');
 assert.deepEqual(lightTextChanges('lines,dialogue').textHighlightRules.split(',').sort(),['dialogue','lines']);
 assert.equal(lightTextChanges('keywords').textHighlightRules,'keywords,dialogue');
 assert.equal(lightTextChanges('off').textColor,'#FFFFFF');
 assert.equal(lightTextChanges('off').textBackground.color,'#111318');
});
test('handwriting uses Long Cang and paired colors remain readable',()=>{
 assert.equal(TEXT_LAYOUT_TEMPLATES.find(p=>p.id==='handwriting').values.textFontId,'long-cang');
 assert.equal(TEXT_LAYOUT_TEMPLATES.find(p=>p.id==='clear').values.textFontId,'lxgw-wenkai');
 for(const mode of ['dark','light']){
  const scheme=textColorScheme(mode);
  assert.equal(hasLowTextContrast(scheme.textColor,scheme.textBackground),false);
 }
 assert.equal(hasLowTextContrast('#222222',{kind:'color',color:'#111318'}),true);
 assert.equal(hasLowTextContrast('#FFFFFF',{kind:'color',color:'#FFFFFF'}),true);
 assert.equal(hasLowTextContrast('#222222',{kind:'library',id:'image',color:'#111318'}),false);
 assert.equal(readableTextColor({kind:'color',color:'#111318'}),'#FFFFFF');
 assert.equal(readableTextColor({kind:'color',color:'#F6EFDF'}),'#222222');
});
test('presets alter only layout and do not replace publishing or selected source',()=>{
 const initial={textSourceLimit:4000,comicPublishSettings:{organizationId:'8',producerAdminName:'test'}};
 const next=applyTextProductionChanges(initial,TEXT_LAYOUT_TEMPLATES[0].values);
 assert.deepEqual(next.comicPublishSettings,initial.comicPublishSettings);
 assert.equal(next.textSourceLimit,4000);assert.equal(next.textFontSize,40);
});
test('paired publishing restores preserve other overrides',()=>{
 const batch={comicPublishSettings:{producerAdminName:'parent',creativeAdminName:'parent',organizationId:'8'}};
 const book={comicPublishSettings:{producerAdminName:'child',creativeAdminName:'child',organizationId:'9'}};
 const groups=publishOverrideRestores(['comicPublishSettings.producerAdminName','comicPublishSettings.creativeAdminName']);
 assert.equal(groups.length,1);
 const next=applyTextProductionChanges(book,Object.fromEntries(groups[0].keys.map(key=>[key,undefined])));
 const {value}=textProductionBookOverride(batch,next);
 assert.equal(value['comicPublishSettings.producerAdminName'],'parent');
 assert.equal(value['comicPublishSettings.creativeAdminName'],'parent');
 assert.equal(value['comicPublishSettings.organizationId'],'9');
});
