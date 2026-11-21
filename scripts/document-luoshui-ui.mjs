import fs from 'node:fs/promises';
import crypto from 'node:crypto';
import vm from 'node:vm';
const root='docs/reference/luoshui-ui-v2.9.5-20261121';
const manifest=JSON.parse(await fs.readFile(root+'/manifest.json','utf8'));
const plugins=JSON.parse(await fs.readFile(root+'/plugin-coverage.json','utf8'));
const groups={};let secretMatches=0;
for(const p of manifest){
 const bytes=await fs.readFile(root+'/'+p.screenshot);
 if(crypto.createHash('sha256').update(bytes).digest('hex')!==p.sha256)throw Error('Hash mismatch: '+p.id);
 const s=await fs.readFile(root+'/'+p.state,'utf8');JSON.parse(s);
 secretMatches+=(s.match(/\bsk-[A-Za-z0-9]{20,}|AIza[A-Za-z0-9_-]{30,}/g)||[]).length;
 (groups[p.group]??=[]).push(p);
}
if(secretMatches)throw Error('Possible credential found; review locally before publishing');
const verified=manifest.filter(p=>p.status==='captured');
const rows=manifest.map(p=>`| ${p.group} | [${p.title}](${p.screenshot}) | ${p.entry.replaceAll('|','/')} | ${p.status} |`).join('\n');
await fs.writeFile(root+'/coverage.md',`# 原版 UI 覆盖登记\n\n采集日期：2026-11-21。原版：洛水 2.9.5，目录 F:\\软件\\洛水2026\\洛水2026。1920×1050 视口。\n\n已确认 ${verified.length} 个实拍状态，覆盖 16 个一级入口。插件入口 ${plugins.length} 项，其中成功进入 ${plugins.filter(p=>p.status==='captured').length} 项。每个状态附截图、页面文字、控件坐标、时间、来源 URL 和 SHA256，见 manifest.json。\n\n## 范围与缺口\n\n- 这是页面参考与截图交互 mock，不是已经实现或验证全部业务功能。热点只是已采集页面之间的参考跳转。\n- 没有运行模型、生成图视频、提交作品、保存设置或删除数据。API 密钥输入在截图时遮挡，不采集输入框的值。\n- 当前小说作品库为 0；正文生成、审稿结果、生产暂停/恢复/失败、导出成品等依赖作品或模型的状态未采集。\n- 上传文件后的状态、付费/权限异常、全部下拉选项与错误组合未穷举；不能据此宣称全功能一比一完成。\n- character-personal 与角色库截图相同，登记为 no-visible-transition，不计入已确认页面。\n- 插件逐项结果见 plugin-coverage.json；未展开项不得作为已完成页面。\n\n## 页面清单\n\n| 分类 | 实拍 | 进入路径 | 状态 |\n| --- | --- | --- | --- |\n${rows}\n\n## 插件登记\n\n${plugins.map(p=>'- '+p.entry+'：'+p.status+'；'+p.note).join('\n')}\n`);
await fs.writeFile(root+'/README.md',`# 洛水原版页面参考与 Mock\n\n打开 [index.html](index.html) 或 [mock.html](mock.html)，点击左侧页面，或“全部缩略图”。支持搜索页面/控件、原尺寸、原图与页面热点跳转。\n\n当前 ${verified.length} 个已确认实拍状态；覆盖记录见 [coverage.md](coverage.md)，原始控件及页面文字在 states/，完整截图在 screenshots/。\n\n界面来自原版 2.9.5 实拍，未重新设计。此 mock 用于核对布局和入口，不连接模型或原版后台，不保存作品。它不等于产品已完成复刻。\n\n复采：在已授权的原版 CDP 9226 上使用 scripts/capture-luoshui-ui.mjs；更新后运行 scripts/build-luoshui-ui-mock.mjs，再运行 scripts/document-luoshui-ui.mjs。\n`);
const html=await fs.readFile(root+'/index.html','utf8');new vm.Script(html.match(/<script>([\s\S]*)<\/script>/)[1]);
const report={date:'2026-11-21',records:manifest.length,captured:verified.length,hashes:'all matched',stateJson:'all parsed',viewerJavaScript:'parsed',credentialPatternMatches:secretMatches};
await fs.writeFile(root+'/validation.json',JSON.stringify(report,null,2));
console.log(JSON.stringify(report));
