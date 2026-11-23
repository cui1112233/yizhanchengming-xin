import assert from 'node:assert/strict';
import test from 'node:test';
import { textPagePresentation, textBookCompletion, canCreateTextCollection, textPageDirtyImpact } from './textMediaState.js';

const ready = { id: 'p1', status: 'ready', activeArtifactId: 'a', renderVersion: 1 };
const locked = { ...ready, id: 'p2', activeArtifactId: 'b', sourceKind: 'user_upload', locked: true };

test('文字页状态和图片只取当前活动版本', () => {
  assert.equal(textPagePresentation(null).state, 'empty');
  assert.equal(textPagePresentation({status:'pending'}).state, 'pending');
  assert.equal(textPagePresentation({status:'pending',renderVersion:1}).state, 'stale');
  assert.equal(textPagePresentation({status:'rendering'}).state, 'rendering');
  assert.equal(textPagePresentation({status:'failed',lastError:'字号过大'}).error, '字号过大');
  assert.equal(textPagePresentation(ready).state, 'ready');
  assert.equal(textPagePresentation({...ready,activeArtifactId:''}).ready, false);
  const display = textPagePresentation({...locked,status:'stale'}, [
    {id:'candidate',pageId:'p2',kind:'page',sourceKind:'system'},
    {id:'b',pageId:'p2',kind:'page',sourceKind:'user_upload'},
  ]);
  assert.equal(display.state,'locked');
  assert.equal(display.artifact.id,'b');
  assert.equal(display.ready,true);
  assert.equal(textPagePresentation(locked, []).ready,false);
});

const document = {id:'doc', activeVersionId:'v', paginationVersion:2, revision:6, coverArtifactId:'cover', coverLocked:false};
const pages = [ready, locked, {id:'p3',status:'failed'}].map(p => ({...p, versionId:'v', paginationVersion:2}));
const cover = {id:'cover', kind:'cover', renderVersion:4, sourceKind:'system'};
const images = [
  {id:'a',kind:'page',pageId:'p1',renderVersion:1,sourceKind:'system'},
  {id:'b',kind:'page',pageId:'p2',renderVersion:1,sourceKind:'user_upload'},
];
test('加载完当前文档和活动封面后才允许合成，旧封面或失败封面不可用', () => {
  assert.equal(canCreateTextCollection(pages,['p1']),false);
  assert.equal(canCreateTextCollection(pages,['p1'],null,[cover,...images]),false);
  assert.equal(canCreateTextCollection(pages,['p1'],document),false);
  for (const changes of [{coverArtifactId:''},{activeVersionId:''},{coverLocked:true}])
    assert.equal(canCreateTextCollection(pages,['p1'],{...document,...changes},[cover,...images]),false);
  for (const changes of [{id:'old-cover'},{kind:'page'},{renderVersion:7},{renderVersion:0},{status:'failed'},{status:'rendering'},{status:'stale'}])
    assert.equal(canCreateTextCollection(pages,['p1'],document,[{...cover,...changes},...images]),false);
  assert.equal(canCreateTextCollection(pages,['p1'],{...document,coverLocked:true},
    [{...cover,sourceKind:'user_upload'},...images]),true);
});

test('只计算当前生产版本和所选页对应的成品', () => {
  assert.deepEqual(textBookCompletion(pages), { ready:2,total:3,percent:67 });
  assert.equal(textBookCompletion([]).percent,0);
  assert.equal(canCreateTextCollection(pages,['p2','p1'],document,[cover,...images]),true);
  for (const ids of [[], ['missing'], ['p1','p1'], ['p1','p3']])
    assert.equal(canCreateTextCollection(pages,ids,document,[cover,...images]),false);
  for (const changes of [{versionId:'old'},{paginationVersion:1},{renderVersion:2}])
    assert.equal(canCreateTextCollection([{...pages[0],...changes}],undefined,document,[cover,...images]),false);
  assert.equal(canCreateTextCollection(pages,['p1'],document,[cover]),false);
  assert.equal(canCreateTextCollection([],undefined,document,[cover,...images]),false);
});

test('编辑正文和改变版心需要重算页界，仅更换颜色或背景只需重绘', () => {
  const a={text:'你来了。',layout:{fontSize:44,padding:96,textColor:'#fff'},background:{kind:'image',id:'oak'}};
  assert.equal(textPageDirtyImpact(a,{...a,revision:4,background:{id:'oak',kind:'image'}}),'none');
  assert.equal(textPageDirtyImpact(a,{...a,text:'你来啦。'}),'reflow');
  assert.equal(textPageDirtyImpact(a,{...a,background:{kind:'color',color:'#000'}}),'render');
  assert.equal(textPageDirtyImpact(a,{...a,highlightReuseEnabled:true}),'render');
  assert.equal(textPageDirtyImpact(a,{...a,layout:{...a.layout,textColor:'#fea'}}),'render');
  assert.equal(textPageDirtyImpact(a,{...a,layout:{...a.layout,padding:0,textColor:'#fea'}}),'reflow');
  for (const key of ['aspectRatio','targetCharacters','fontId','fontSize','fontWeight','lineHeight','paragraphSpacing','maxHeightRatio'])
    assert.equal(textPageDirtyImpact({layout:{}},{layout:{[key]:1}}),'reflow',key);
  assert.equal(textPageDirtyImpact({textFontSize:44},{textFontSize:66}),'reflow');
  assert.equal(textPageDirtyImpact({textBackground:{kind:'color',color:'#fff'}},{textBackground:{kind:'color',color:'#eee'}}),'render');
});
