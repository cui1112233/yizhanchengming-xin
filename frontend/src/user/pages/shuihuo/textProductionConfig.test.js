import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { TEXT_PRODUCTION_BOOK_REGIONS, TEXT_PRODUCTION_CONFIG_KEYS, TEXT_PRODUCTION_SECTIONS, applyTextProductionChanges, defaultTextProductionConfig, normalizeTextProductionConfig, textProductionBookOverride, textProductionInitialSection, updateTextProductionSettings, textProfileChanges } from './textProductionConfig.js';

test('文本单书配置固定分为引擎、排版、发布三个区块', () => {
  assert.deepEqual(TEXT_PRODUCTION_SECTIONS.map(section => section.label), ['引擎配置', '排版配置', '发布配置']);
});

test('标红规则、图片行号可以继承、单书覆盖和恢复',()=>{
 const batch={textHighlightRules:'dialogue,lines',textHighlightLines:'1,3'};
 assert.equal(textProductionBookOverride(batch,{}).value.textHighlightRules,'dialogue,lines');
 const book=applyTextProductionChanges({},{textHighlightRules:'off'});
 assert.equal(textProductionBookOverride(batch,book).value.textHighlightRules,'off');
 const restored=applyTextProductionChanges(book,{textHighlightRules:undefined});
 assert.equal(textProductionBookOverride(batch,restored).value.textHighlightRules,'dialogue,lines');
 assert.equal(textProductionBookOverride(batch,restored).value.textHighlightLines,'1,3');
});

test('智能排版默认开启，目标字数保留但仅在目标模式生效，单书可独立覆盖',()=>{
  assert.equal(defaultTextProductionConfig().textPaginationMode,'auto');
  const batch=applyTextProductionChanges({textPageTarget:150},{textPaginationMode:'auto'});
  const book=applyTextProductionChanges({},{textPaginationMode:'target',textPageTarget:200});
  const reopened=textProductionBookOverride(batch,JSON.parse(JSON.stringify(book)));
  assert.equal(reopened.value.textPaginationMode,'target');
  assert.equal(reopened.value.textPageTarget,200);
  const restored=applyTextProductionChanges(book,{textPaginationMode:undefined});
  assert.equal(textProductionBookOverride(batch,restored).value.textPaginationMode,'auto');
});

test('工作台单书配置显示三个入口并分别打开对应页签', () => {
  assert.deepEqual(TEXT_PRODUCTION_BOOK_REGIONS.map(region => [region.key,region.label]), [
    ['models','引擎配置'],
    ['layout','排版配置'],
    ['publish','发布配置']
  ]);
  assert.equal(textProductionInitialSection('models'),'models');
  assert.equal(textProductionInitialSection('layout'),'layout');
  assert.equal(textProductionInitialSection('publish'),'publish');
  assert.equal(textProductionInitialSection('unknown'),'models');
});

test('新建及统一保存配置档后重新打开有名称、组织、分类及发布参数，单书覆盖可独立还原', () => {
  const id = 'publishSettings.websiteProfileId';
  const catalog = 'publishSettings.websiteProfiles';
  const profiles = [
    {id:'121-4',name:'滚屏',organization:'3'},
    {id:'121-8',config_id:'8',name:'解压',advanced:{organization_id:7,jieyaNum:8,gunpingNum:0,book_category:'new',password:'secret',fontColorStyles:[1,2]},apiKey:'secret'}
  ];
  const batch = applyTextProductionChanges({publishSettings:{bookCategory:'已有分类'}},textProfileChanges(profiles,'121-4'));
  const reopen = settings => JSON.parse(JSON.stringify(settings));
  const override = applyTextProductionChanges({},textProfileChanges(profiles,'121-8'));
  const fixture = new URL('../../../../../backend/internal/httpapi/testdata/text_profile_snapshot.json',import.meta.url);
  assert.deepEqual(override,JSON.parse(readFileSync(fixture,'utf8')));
  const book = textProductionBookOverride(reopen(batch),reopen(override));
  assert.equal(book.value[id],'121-8');
  assert.equal(book.value[catalog][1].name,'解压');
  assert.equal(override.publishSettings.versionProfile,'解压');
  assert.equal(override.publishSettings.organization,'7');
  assert.equal(book.value['publishSettings.bookCategory'],'已有分类');
  assert.deepEqual(override.publishSettings.websiteProfiles[1].advanced,{jieyaNum:8,gunpingNum:0,book_category:'new',fontColorStyles:[1,2]});
  assert.ok(!JSON.stringify(override).includes('secret'));
  const restored = applyTextProductionChanges(override,{[id]:undefined});
  assert.deepEqual(restored,{});
  assert.equal(textProductionBookOverride(batch,restored).value[id],'121-4');
  const refreshed = applyTextProductionChanges(batch,textProfileChanges(profiles.slice(1),'121-4'));
  assert.equal(refreshed.publishSettings.websiteProfileId,'');
  assert.equal(refreshed.publishSettings.organization,'');
  assert.equal(refreshed.publishSettings.bookCategory,'已有分类');
});

test('取文失败策略和生产失败说明分开，修改生产模式保留既有取文策略', () => {
  const fields = TEXT_PRODUCTION_SECTIONS.find(section=>section.key==='models').fields;
  assert.match(fields.find(field=>field.key==='sensitiveAIOnFailure').label,/取文阶段/);
  assert.match(fields.find(field=>field.key==='sensitiveMode').aiHint,/停止生产并显示失败原因/);
  const saved = applyTextProductionChanges({sensitiveAIOnFailure:'review'},{sensitiveMode:'ai'});
  assert.deepEqual(saved,{sensitiveMode:'ai',sensitiveAIOnFailure:'review'});
});

test('同一套字段覆盖三个表单，实际边距、段距和标红保留零值', () => {
  for (const key of ['textModelId','sensitiveMode','sensitiveAIModels','sensitiveAIOnFailure','highlightReuseEnabled','contentRange','textPageTarget','imageAspectRatio','textFontId','textFontSize','textFontWeight','textLineHeight','textParagraphSpacing','textPadding','textAlign','textColor','textHighlightColor','textMaxHeightRatio','textBackground','textCover.layout.fontId','textCover.layout.fontSize','textCover.layout.align','comicPublishSettings.producerAdminName','comicPublishSettings.organizationId','publishSettings.websiteProfileId','publishSettings.bookCategory','textAutoCollection','textAutoUpload','textUploadRetryLimit']) assert.ok(TEXT_PRODUCTION_CONFIG_KEYS.includes(key),key);
  assert.ok(!TEXT_PRODUCTION_CONFIG_KEYS.includes('textCover.background'));
  const d=defaultTextProductionConfig();
  assert.deepEqual([d.textPageTarget,d.textMaxPages,d.contentRange,d.imageAspectRatio,d.textMaxHeightRatio],[150,35,4000,'3:4',0.75]);
  const off=normalizeTextProductionConfig({textPadding:0,textParagraphSpacing:0,highlightReuseEnabled:false,publishSettings:{biaohongReuse:true}});
  assert.deepEqual([off.textPadding,off.textParagraphSpacing,off.highlightReuseEnabled],[0,0,false]);
  assert.equal(normalizeTextProductionConfig({publishSettings:{biaohongReuse:true}}).highlightReuseEnabled,true);
  const oldBook=textProductionBookOverride({highlightReuseEnabled:false,sensitiveMode:'off'},{publishSettings:{biaohongReuse:true},sensitiveAIEnabled:true});
  assert.equal(oldBook.value.highlightReuseEnabled,false);
  assert.equal(oldBook.value.sensitiveMode,'off');
  assert.ok(!oldBook.explicitKeys.includes('highlightReuseEnabled'));
});

test('单书设置可显式保持与批次相同的值，子字段恢复继承保留同组其他覆盖', () => {
  const batch={textFontSize:60,textCover:{layout:{align:'left',fontSize:70}},publishSettings:{bookCategory:'新书'}};
  const book={textCover:{layout:{align:'right',fontSize:80}},publishSettings:{extra:1}};
  const {value, explicitKeys}=textProductionBookOverride(batch,book);
  assert.equal(value.textFontSize,60);
  assert.equal(value['textCover.layout.align'],'right');
  assert.ok(!explicitKeys.includes('textFontSize'));
  const update=updateTextProductionSettings(book,{'textFontSize':60,'textCover.layout.align':undefined,'publishSettings.bookCategory':'旧书'});
  assert.deepEqual(update,{patch:{textFontSize:60,textCover:{layout:{fontSize:80}},publishSettings:{extra:1,bookCategory:'旧书'}},restoreKeys:[]});
  assert.deepEqual(updateTextProductionSettings({contentRange:8000,contentCaptureCharacters:8000},{contentRange:undefined}),{patch:{},restoreKeys:['contentRange','contentCaptureCharacters']});
  assert.deepEqual(updateTextProductionSettings({},{contentRange:8000}),{patch:{contentRange:8000,contentCaptureCharacters:8000},restoreKeys:[]});
});

test('恢复旧标红/敏感覆盖后不再次读出本书的旧值', () => {
  const original={sensitiveAIEnabled:true,publishSettings:{biaohongReuse:true,other:2}};
  assert.deepEqual(updateTextProductionSettings(original,{highlightReuseEnabled:undefined,sensitiveMode:undefined}),{patch:{publishSettings:{other:2}},restoreKeys:['sensitiveAIEnabled']});
});
