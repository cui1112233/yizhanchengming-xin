import test from 'node:test';
import assert from 'node:assert/strict';
import {TEXT_LAYOUT_TEMPLATES,publishOverrideRestores} from './textLayoutTemplates.js';
import {applyTextProductionChanges,textProductionBookOverride} from './textProductionConfig.js';
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
