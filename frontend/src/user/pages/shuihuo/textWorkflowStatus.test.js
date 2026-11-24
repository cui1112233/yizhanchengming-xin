import test from 'node:test';
import assert from 'node:assert/strict';
import {textUpdateState,collectionIsOutdated} from './textWorkflowStatus.js';
const old={layout:{paginationMode:'auto',fontSize:44},background:{kind:'color',color:'#111318'}};
const collection={versionId:'v',paginationVersion:1,items:[{pageId:'p',artifactId:'a'}]};
const fixture=()=>({document:{activeVersionId:'v',paginationVersion:1,configSnapshot:structuredClone(old)},settingsSnapshot:structuredClone(old),pages:[{id:'p',status:'ready',activeArtifactId:'a'}]});
test('distinguishes geometry, paint, pending pages and collection state',()=>{
 const x=fixture();assert.equal(textUpdateState(x,collection).stage,'ready');
 x.settingsSnapshot.layout.fontSize=56;assert.equal(textUpdateState(x,collection).stage,'reflow');
 x.settingsSnapshot=structuredClone(old);x.settingsSnapshot.background.color='#ffffff';assert.equal(textUpdateState(x,collection).stage,'render');
 x.settingsSnapshot=structuredClone(old);x.pages[0].status='pending';assert.equal(textUpdateState(x,collection).pending,1);
 x.pages[0].status='ready';assert.equal(textUpdateState(x).stage,'compose');
});
test('automatic layout ignores inactive target while keeping real overrides',()=>{
 const x=fixture();x.settingsSnapshot.layout.targetCharacters=500;
 assert.equal(textUpdateState(x,collection).stage,'ready');
 x.settingsSnapshot.layout.paginationMode='target';assert.equal(textUpdateState(x,collection).stage,'reflow');
});
test('old collection is detected from version and immutable page images',()=>{
 const x=fixture();assert.equal(collectionIsOutdated(collection,x.document,x.pages),false);
 x.pages[0].activeArtifactId='new';assert.equal(collectionIsOutdated(collection,x.document,x.pages),true);
 x.document.paginationVersion=2;assert.equal(collectionIsOutdated(collection,x.document,x.pages),true);
});
