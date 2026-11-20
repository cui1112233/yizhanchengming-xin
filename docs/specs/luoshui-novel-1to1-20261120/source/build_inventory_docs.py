"""Evidence inventories, not a claim that recovered code is original source."""
import json, pathlib, hashlib, datetime, re

ROOT = pathlib.Path(__file__).resolve().parents[4]
OUT = pathlib.Path(__file__).resolve().parent.parent
NATIVE = ROOT / 'docs/reverse-engineering/luoshui-novel-native-20261119'

def read(path):
    return json.loads(path.read_text(encoding='utf-8-sig'))

def write(name, text):
    (OUT / name).write_text(text, encoding='utf-8')

def cell(text):
    return str(text).replace('|', '\\|').replace('\n', '<br>')

def link(file, line):
    return f'[{file}:{line}](source/frontend/{file}#L{line})'

ui = read(OUT / 'source/ui-inventory.json')
identity = read(OUT / 'source/identity.json')
files = []
for item in ui['components']:
    p = OUT / 'source/frontend' / item['file']
    data = p.read_bytes()
    files.append({'file': 'source/frontend/' + item['file'], 'sha256': hashlib.sha256(data).hexdigest(), 'lines':len(data.decode('utf-8').splitlines())})
manifest = read(NATIVE / 'manifest.json')
identity['constantModules'] = len(manifest['modules'])
identity['recoveredFiles'] = files
identity['sourceType'] = 'shipped-text and native pseudocode; not author source repository'
(OUT / 'source/evidence-index.json').write_text(json.dumps(identity,ensure_ascii=False,indent=2),encoding='utf-8')

parts = ['# 01 版本与证据规则\n',
    f"当前前端：`{identity['frontend']['path']}`，SHA256 `{identity['frontend']['sha256']}`。\n",
    f"当前原生：`{identity['native']['path']}`，SHA256 `{identity['native']['sha256']}`。\n",
    '前端经过 Babel 仅解析、格式化，没有执行目标程序；文件头记录原 UTF16 范围。行号指恢复片段行号，不是原作者工程行号。\n',
    '原生证据区分 IDA 伪代码、旧有 Ghidra 伪代码、常量。常量出现不证明调用发生；静态事件不证明按钮可用；文档覆盖不证明 CM 已对齐。\n',
    '旧前端 SHA `5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a` 只作历史参考，不能替代当前前端。\n',
    '## 当前组件清单\n\n| 组件 | 恢复文件 | 原 UTF16 范围 | SHA256 |\n|---|---|---|---|']
for c in ui['components']:
    parts.append(f"| {c['name']} | {link(c['file'],1)} | {c['originalUtf16Range']} | `{c['recoveredSha256']}` |")
parts += ['\n## 原生分析位置\n',
    '- IDA 数据库：`_local-archive-20261028/luoshui-ida-20261120/native/novels-sdk.pyd.i64`；工作副本 `C:/rs-ida/novels-sdk.pyd.i64`。本轮新增 IDA 导出见 `source/ida/`。',
    '- 已解码常量：`docs/reverse-engineering/luoshui-novel-native-20261119/constants/`（95模块；完整常量不等于完整函数体）。',
    '- 原生函数绑定：同目录 `function-bindings.json`；已有函数体覆盖见 `native-index.json`。',
    '- 既有短篇完整体/Listing：`docs/verification/novel-full-draft-20261120/`；生产确认与终审：`docs/verification/novel-rea-followup-20261120/`。',
    '- 证据文件逐个哈希见 [evidence-index.json](source/evidence-index.json)。\n']
write('01-evidence.md','\n'.join(parts))

parts=['# 02 页面、事件与界面文案清单\n',
 '该清单按当前安装包23个Novel组件提取。440条记录包含点击、更新事件、禁用条件、占位文案等，**不是440个独立功能或按钮**。事件条目只登记表达式，未冒称已完成调用链核验。\n',
 '## 组件与导航事件\n\n| 组件 | 发出事件 | 行号 |\n|---|---|---|']
for e in ui['emits']:
    parts.append(f"| {e['component']} | `{e['event']}` | {link(e['file'],e['line'])} |")
parts += ['\n## 控件和条件逐项登记\n\n| 编号 | 组件/属性 | 表达式 | 来源 | 核验状态 |\n|---|---|---|---|---|']
for c in ui['controls']:
    parts.append(f"| {c['id']} | {c['component']} / `{cell(c['kind'])}` | `{cell(c['expression'])}` | {link(c['file'],c['line'])} | 静态已提取；运行未测 |")
parts += ['\n## 原版可见中文文案\n\n| 组件 | 文案（仅静态字面量，动态拼接另查源码） | 来源 |\n|---|---|---|']
for s in ui['strings']:
    parts.append(f"| {s['component']} | {cell(s['text'])} | {link(s['file'],s['line'])} |")
write('02-ui-inventory.md','\n'.join(parts)+'\n')

api = read(NATIVE / 'constants/novels--api--novels_api.json')
methods=[e for e in api['entries'] if isinstance(e.get('value'),str) and e['value'].startswith('novels_') and e['index'] >= 347]
(OUT/'source/api-name-catalog.json').write_text(json.dumps(methods,ensure_ascii=False,indent=2),encoding='utf-8')
parts=['# 05 原生接口与提示词来源\n',
 '## 原生桥接接口名目录\n',
 '下表来自原生常量；证明名称存在，不证明前端绑定、输入签名、实际网络请求已全部还原。不能拿CM REST路径冒称原版桥接。\n',
 '| 名称 | 常量序号 | 原件字节范围 | 进一步核验 |\n|---|---|---|---|']
for e in methods:
    parts.append(f"| `{e['value']}` | {e['index']} | {e['fileRange']} | 见功能规格；混淆桥接逐项待绑定 |")
catalog=[]
for m in manifest['modules']:
    if not m['module'].startswith('novels.prompts'): continue
    p=NATIVE/m['path']; entries=read(p)['entries']
    for e in entries:
        v=e.get('value')
        if isinstance(v,str) and (len(v)>=100 or v.startswith('build_')):
            readable='readable/'+p.stem+'.md'
            rp=NATIVE/readable
            lines=rp.read_text(encoding='utf-8').splitlines()
            line=next((i+1 for i,s in enumerate(lines) if s.startswith('## Constant '+str(e['index'])+' |')),None)
            source=readable
            if line is None:
                source=m['path']
                line=next(i+1 for i,s in enumerate(p.read_text(encoding='utf-8').splitlines()) if s.strip().rstrip(',')=='"index": '+str(e['index']))
            catalog.append({'module':m['module'],'constant':e['index'],'fileRange':e['fileRange'],'source':'docs/reverse-engineering/luoshui-novel-native-20261119/'+source,'line':line,'characters':len(v),'sha256':hashlib.sha256(v.encode()).hexdigest(),'preview':v[:110], 'dynamicUse':'requires caller and builder evidence; not proven by literal alone'})
(OUT/'source/prompt-catalog.json').write_text(json.dumps(catalog,ensure_ascii=False,indent=2),encoding='utf-8')
parts+=['\n## 提示词目录\n',
 '收录规则：prompts模块中长度≥100的字符串，以及build_名称。它是可重复的字面量/构造器候选目录，不是“所有有效提示词”的计数。完整原文保留在来源文件，不能把目录数量当已接入数量。\n',
 '| 模块 | 常量 | 原件字节 | 原文来源行号 | 字符数 | 摘要 |\n|---|---|---|---|---|---|']
for e in catalog:
    parts.append(f"| `{e['module']}` | {e['constant']} | {e['fileRange']} | `{e['source']}:{e['line']}` | {e['characters']} | {cell(e['preview'])} |")
write('05-api-prompts.md','\n'.join(parts)+'\n')
print(json.dumps({'components':len(files),'controlExpressions':len(ui['controls']),'eventDeclarations':len(ui['emits']),'apiNameConstants':len(methods),'promptCandidates':len(catalog)},ensure_ascii=False))
