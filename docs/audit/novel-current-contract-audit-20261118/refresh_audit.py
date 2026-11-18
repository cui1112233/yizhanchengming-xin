"""Refresh evidence locations, not a behavioral test or parity certification."""
from pathlib import Path
import json, hashlib, collections, subprocess

ROOT = Path('F:/脚本测试/一战晟铭-v88')
OUT = Path(__file__).parent
BASE = ROOT / 'docs/audit/novel-full-parity-20261118/features.json'
ORIGINAL = ROOT / '_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current'
PAGE = 'frontend/src/user/pages/NovelCreationPage.jsx'
UI = 'frontend/src/user/pages/novel-creation/'
GO = 'backend/internal/novelcreation/'

def loc(file, anchor):
    p = ROOT / file
    body = p.read_text(encoding='utf-8-sig')
    pos = body.find(anchor)
    return dict(file=file, line=body[:pos].count('\n')+1 if pos >= 0 else None,
                anchor=anchor, anchorFound=pos >= 0, sha256=hashlib.sha256(p.read_bytes()).hexdigest())

def save(name, value):
    (OUT / name).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')

# These revisions come from direct source reads and free current UI observations.
# Unchanged baseline statements retain their qualification; none is marked equal.
revisions = {
 'E01': ('部分实现', '长篇原创已能无书讨论、返回create_book建议并确认/取消；仅此动作接通，原版其他建议动作及完整书籍资料合同仍缺。', UI+'StudioChat.jsx', 'async function act'),
 'E03': ('部分实现', '已存在长篇仿写专用表单，默认5000字/100章、文件选章、多文风及备忘录；不再算缺失。原生私有提示词、真实生成及同状态全部交互未验证。', UI+'LongImitationDialog.jsx', 'export default function'),
 'E04': ('部分实现', '长篇仿写有独立强类型输入和不可变快照；通用白名单不代表长篇专用接口。其余模式不能沿用这项交付，原生拆解算法仍未知。', GO+'long_imitation.go', 'type LongImitationInput'),
 'E10': ('部分实现', '无书讨论仅long_original走StudioChat；短篇/番外仍先要求project，六模式关系没有统一对齐。', PAGE, "const studioView ="),
 'L01': ('部分实现', '六入口已改两列带lucide模式图标；这修复了旧单列问题，但未逐个证明与原图标/几何一致。', PAGE, 'const creationIcons'),
 'C02': ('部分实现', '长篇无书讨论可调4000至64000；正式作品通用会话仍固定120000，未提供同样入口。', UI+'ComposerToolbar.jsx', '{onBudgetChange &&'),
 'C09': ('部分实现', '使用自己的模型目录和提示词式思考要求，符合自家配置要求；原模型分组/能力控制与真实服务能力映射尚未逐项证实。', UI+'ComposerToolbar.jsx', 'aria-label="创作文本模型"'),
 'C10': ('部分实现', 'StudioChat已有create_book确认/取消卡；原write_next/audit/revise/short_run等建议动作没有整套接入。', UI+'StudioChat.jsx', '创建作品建议'),
 'S02': ('部分实现', '长篇仿写专用输入及续写导入支持多文风；普通创作仍是单styleProfileId，文风工具并未整体改为多选。', UI+'MaterialsPanel.jsx', '选用文风'),
 'S06': ('部分实现', '长篇仿写已有独立全参考分块链；普通资料仍明确限制每份12000字，两条路径不同，不能宣称所有长参考均已对齐。', UI+'MaterialsPanel.jsx', '每份参考请选择正文'),
}
extra = {
 'E02': [(GO+'exports.go','func (s *Service) GetShortResult')],
 'E05': [(PAGE,"if(id==='long_imitation')"),(PAGE,'<LongImitationDialog')],
 'E06': [(GO+'validation.go','maxTargetChapterWords'),(GO+'validation.go','"short_imitation"')],
 'E07': [(UI+'contracts.js',"['fanfiction', '番外创作']"),(GO+'validation.go','"fanfiction"')],
 'E08': [(UI+'contracts.js',"['fanfiction', '番外创作']")],
 'E09': [(UI+'ContinuationImportDialog.jsx','export default function')],
 'C02': [(PAGE,'contextBudget: 120000')],
 'C04': [(UI+'StudioChat.jsx','snapshot?.messages')],
 'C07': [(UI+'StudioChat.jsx','snapshot?.messages')],
 'V07': [(GO+'exports.go','func (s *Service) GetShortResult')],
 'V08': [(GO+'exports.go','func (s *Service) GetShortResult')],
 'V09': [(GO+'exports.go','func (s *Service) ExportShortResult')],
 'V10': [(GO+'exports.go','func (s *Service) ExportShortResult')],
 'A01': [(UI+'AgentPanel.jsx','PHASES.filter')],
 'A02': [(UI+'AgentPanel.jsx','<AgentRequirements />')],
 'D01': [(PAGE,"['章节数', dashboard.chapterCount]")],
 'D02': [(PAGE,"['章节数', dashboard.chapterCount]")],
 'D03': [(PAGE,"['章节数', dashboard.chapterCount]")],
}
rows = json.loads(BASE.read_text(encoding='utf-8-sig'))
for row in rows:
    row['baselineStatus'] = row['status']
    row['baselineFinding'] = row['finding']
    row['original'] = loc(row['original']['file'], row['original']['anchor'])
    row['current'] = loc(row['current']['file'], row['current']['anchor'])
    if row['id'] in revisions:
        status, finding, file, anchor = revisions[row['id']]
        row.update(status=status, finding=finding, current=loc(file, anchor))
    row['additionalCurrentEvidence'] = [loc(*v) for v in extra.get(row['id'], [])]
    row['reviewLevel'] = '原版恢复源码与当前静态合同复核；现场覆盖另见ui-observations.json'
    row['runtimeAcceptance'] = '未做原/CM付费同输入成品对照；不得据此标为全功能通过'
    row['implementation'] = '按当前状态及repair-spec.md登记；没有修改产品代码'
save('features.json', rows)
bad = [(r['id'], scope, a['anchor']) for r in rows for scope,a in [('original',r['original']),('current',r['current'])]] + [(r['id'],'additional',a['anchor']) for r in rows for a in r['additionalCurrentEvidence']]
missing = []
for rid,scope,anchor in bad:
    row = next(r for r in rows if r['id']==rid)
    entries = row['additionalCurrentEvidence'] if scope=='additional' else [row[scope]]
    if any(a['anchor']==anchor and not a['anchorFound'] for a in entries): missing.append([rid,scope,anchor])
counts = dict(collections.Counter(r['status'] for r in rows))
save('summary.json', dict(count=75, counts=counts, missingEvidenceAnchors=missing,
    sourceHead=subprocess.check_output(['git','-C',str(ROOT),'rev-parse','HEAD'],text=True).strip(),
    runtimeObserved={'port':4001,'git_sha':'c4054404','entry':'/assets/user-LoEy2Yrp.js'},
    paidCalls=0, productFilesChanged=0, originalNativeParity=False,
    note='75项是已有审计范围，不是原软件功能总数。静态相似和自身测试通过不代表原版运行等价。'))
manifest=[]
for folder in [ROOT/UI,ROOT/GO,ORIGINAL]:
    for p in sorted(folder.iterdir()):
        if p.is_file() and p.suffix in ['.js','.jsx','.go','.css'] and not p.name.endswith('_test.go') and '.test.' not in p.name:
            manifest.append(dict(file=p.relative_to(ROOT).as_posix(),bytes=p.stat().st_size,sha256=hashlib.sha256(p.read_bytes()).hexdigest()))
manifest.append(dict(file=PAGE,sha256=hashlib.sha256((ROOT/PAGE).read_bytes()).hexdigest()))
save('source-manifest.json',manifest)
lines=['# 小说创作逐项复核（2026-11-18）','','结论：没有全量对齐。此次复核原版恢复源码与当前代码，更新旧审计75项，另做免费页面检查；没有修改产品，没有付费生成。','',
       '原包：F:/软件/洛水2026/洛水2026/dist/assets/index-C1aA08SG.js；SHA256 5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a。',
       '恢复源码：_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current。',
       '现场4001：c4054404，/assets/user-LoEy2Yrp.js；仓库文档提交8ab7e06f。',
       '', '## 本次结果', '', '、'.join(f'{k} {v}项' for k,v in counts.items())+'。没有一项被无证据地改成“完全一致”。',
       '', '旧审计仍保留。已纠正长篇原创、长篇仿写、双列图标和部分上下文设置的过时结论；修过一个切片不代表整项闭环。',
       '', '## 确认的问题', '',
       '1. 仿写短篇没有独立13000字表单；点击仅切模式，提交仍走通用章节。',
       '2. 短篇结果为ChapterVersion，缺独立storyId、大纲/成品文件/推荐文案的保存及重开。',
       '3. 番外仍映射fanfiction；正传书籍关联及续篇/前传/平行/IF分支没有对应参数。',
       '4. Agent只有8条可配置模型路径，缺原17分工及37维开关；补充提示词不等于内置维度。',
       '5. 内容检查点仍不等价：逐次模型调用确认与大纲文件编辑后继续是两件事。',
       '6. 新建作品/Agent表单允许10000字，通用后端只收800—8000；8001—10000存在真实边界冲突。独立长篇仿写接口另有范围，不混为一谈。',
       '7. 通用文风管理、资料文件、消息操作、章节审修快捷动作、看板趋势与伏笔仍有缺口。',
       '', '## 核验边界', '',
       '已实际打开当前六创作入口和六工具、关闭/返回部分弹层，记录在ui-observations.json。没有逐个点击收费、删除和启动自动任务按钮。',
       '原版已打开窗口的无障碍树可读，显示长篇仿写表单；本轮原版点击返回geometry unavailable，窗口截图超时，Raise/Escape未退出弹窗。原版其余状态以保留恢复源码及用户截图为证，不能称本轮逐按钮实测完成。',
       '已运行7个顶层免费Go测试，含4个模式子例，全部通过；仅证明当前校验/持久确认规则，不证明与原版一致。详见verification.md。',
       '', '## 逐项记录', '', '|编号|功能|当前判定|当前结论|原版位置|当前代码位置|', '|---|---|---|---|---|---|']
for r in rows:
    o,c = r['original'],r['current']
    lines.append(f"|{r['id']}|{r['feature']}|{r['status']}|{r['finding'].replace('|','/')}|{Path(o['file']).name}:{o['line']}|{c['file']}:{c['line']}|")
lines += ['', '详细字段和hash见features.json；修复顺序与输入/输出/边界/失败规则见repair-spec.md。原版原生私有提示词及实际模型次序仍未知，不能把CM自编提示词写成已经恢复原版。']
(OUT/'README.md').write_text('\n'.join(lines)+'\n',encoding='utf-8')
print(json.dumps({'counts':counts,'missingAnchors':missing},ensure_ascii=False))
