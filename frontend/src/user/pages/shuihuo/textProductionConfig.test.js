import test from 'node:test';
import assert from 'node:assert/strict';
import { TEXT_PRODUCTION_CONFIG_KEYS, defaultTextProductionConfig, normalizeTextProductionConfig, textProductionBookOverride, updateTextProductionSettings } from './textProductionConfig.js';

test('同一套字段覆盖三个表单，实际边距、段距和标红保留零值', () => {
  for (const key of ['textModelId','sensitiveMode','sensitiveAIModels','sensitiveAIOnFailure','highlightReuseEnabled','contentRange','textPageTarget','imageAspectRatio','textFontId','textFontSize','textFontWeight','textLineHeight','textParagraphSpacing','textPadding','textAlign','textColor','textHighlightColor','textMaxHeightRatio','textBackground','textCover.background','textCover.layout.fontId','textCover.layout.fontSize','textCover.layout.align','comicPublishSettings.producerAdminName','comicPublishSettings.organizationId','publishSettings.websiteProfileId','publishSettings.bookCategory','textAutoCollection','textAutoUpload','textUploadRetryLimit']) assert.ok(TEXT_PRODUCTION_CONFIG_KEYS.includes(key),key);
  const d=defaultTextProductionConfig();
  assert.deepEqual([d.textPageTarget,d.contentRange,d.imageAspectRatio,d.textMaxHeightRatio],[150,4000,'3:4',0.75]);
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
