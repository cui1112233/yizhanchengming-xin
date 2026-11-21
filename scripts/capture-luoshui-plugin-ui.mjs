// Open only observed "去使用" entries and capture their UI. Never submit a form.
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
const root='docs/reference/luoshui-ui-v2.9.5-20261121';
const run=(...args)=>execFileSync(process.execPath,['scripts/capture-luoshui-ui.mjs',...args],{encoding:'utf8',timeout:30000});
const state=()=>JSON.parse(run('state'));
const inventory=JSON.parse(fs.readFileSync(root+'/states/plugins.json','utf8')).controls.filter(c=>c.label.startsWith('去使用'));
const names=['资产参考','角色修改','封面生成','角色三视图','AI-分段','AI-CUT插件版','AI-Shot','AI-Upscale','AI-Split','分镜大师','草稿合并','剪映关键帧','AI-Subtitle','AI-Filler','AI-双语字幕','AI-BGM','字幕对轴','声音克隆','字幕合并','音色设计','小说转剧','AI改文','爆款开头','女频性转','格式化文本','长篇拆书','世界观','文档合并','书名生成'];
const results=[];
const previous=process.argv.includes('--retry')?JSON.parse(fs.readFileSync(root+'/plugin-coverage.json','utf8')):[];
for(let i=0;i<inventory.length;i++){
 if(previous[i]?.status==='captured'){results.push(previous[i]);continue;}
 const expected=inventory[i].label;const name=names[i]||expected;
 try{
  let s=state();let entry=s.controls.find(c=>c.label===expected);
  if(!entry){const expand=s.controls.find(c=>c.label==='展开侧边栏');if(expand){run('click',String(expand.index),expand.label);s=state();}const nav=s.controls.find(c=>c.label==='插件');if(!nav)throw Error('Plugin navigation absent');run('click',String(nav.index),nav.label);s=state();entry=s.controls.find(c=>c.label===expected);}
  if(!entry){const back=s.controls.find(c=>c.index>26&&c.label==='返回');if(back){run('click',String(back.index),back.label);s=state();entry=s.controls.find(c=>c.label===expected);}}
  if(!entry)throw Error('Observed entry no longer exists');
  const before=s.text;run('click',String(entry.index),entry.label);s=state();
  // Capture waits for document text to settle. Do not infer success from a click.
  const id='plugin-'+String(i+1).padStart(2,'0');run('capture',id,name+'·插件页面','插件详情','插件中心→'+expected);
  s=state();const changed=s.text!==before;
  results.push({id,name,entry:expected,status:changed?'captured':'no-visible-transition',note:changed?'仅进入界面，未运行处理':'未观察到当前主窗口变化，可能需要其他窗口；不得称详情已采集'});
  run('key','Escape');s=state();
  // Close a form only through an explicit, observed close/cancel control.
  for(let n=0;n<2&&!s.controls.some(c=>c.label===expected);n++){
   const close=s.controls.filter(c=>c.index>26&&/^(close|关闭|取消|返回插件)$/.test(c.label)).at(-1);
   if(!close)break;run('click',String(close.index),close.label);s=state();
  }
  console.log(JSON.stringify(results.at(-1)));
 }catch(e){results.push({entry:expected,status:'unresolved',note:String(e.message).slice(0,300)});console.log(JSON.stringify(results.at(-1)));}
 fs.writeFileSync(root+'/plugin-coverage.json',JSON.stringify(results,null,2));
}
