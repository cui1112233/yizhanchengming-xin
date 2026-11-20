"""Validate traceability/coverage of documents, never execute the target or models."""
import pathlib,json,re,hashlib,datetime

OUT=pathlib.Path(__file__).resolve().parent.parent
ROOT=OUT.parents[2]
def read(p): return json.loads(p.read_text(encoding='utf-8-sig'))
ui=read(OUT/'source/ui-inventory.json')
def esc(s): return str(s).replace('|','\\|').replace('\n','<br>')
parts=['# 08 当前23组件的输入、输出和边界原式\n',
 '每个组件按实际props、emits和控件条件登记。表中表达式尚未全部手工解混淆；只代表静态合同，不是运行验收。无emits的展示组件不会被凭空加保存/生成按钮。\n']
for c in ui['components']:
    name=c['name']; file=c['file']
    parts += [f'## {name}\n',f'来源：`source/frontend/{file}:1`。原bundle范围：{c["originalUtf16Range"]}。\n',
      '### 输入\n\n| 属性 | 类型/默认表达式 | 行号 |\n|---|---|---|']
    props=[p for p in ui['props'] if p['component']==name]
    for p in props: parts.append(f'| {esc(p["name"])} | `{esc(p["definition"])}` | {p["line"]} |')
    if not props: parts.append('| 无显式props | 通过内部上下文/状态获取；完整外部依赖待查 | — |')
    es=[e for e in ui['emits'] if e['component']==name]
    parts += ['\n### 输出\n', ', '.join('`'+e['event']+'` (:'+str(e['line'])+')' for e in es) if es else '没有显式emits声明；不推定有后台写入。',
      '\n### 边界和动作原式\n\n| 编号 | 属性 | 表达式 | 行号 |\n|---|---|---|---|']
    for a in ui['controls']:
        if a['component']==name: parts.append(f'| {a["id"]} | {esc(a["kind"])} | `{esc(a["expression"])}` | {a["line"]} |')
    errors=[s for s in ui['strings'] if s['component']==name and re.search('失败|不存在|不支持|至少|请|取消|暂无|等待|稍后',s['text'])]
    parts += ['\n### 失败、空态和取消文案\n']
    parts += [f'- `:{s["line"]}` {esc(json.dumps(s["text"],ensure_ascii=False))}' for s in errors] or ['无独立失败文案字面量；父层/动态分支见对应源码，不能编造。']
    parts += ['\n运行状态：未实测；完整语义与剩余缺口见03/04/06/07。\n']
(OUT/'08-component-contracts.md').write_text('\n'.join(parts),encoding='utf-8')

errors=[];checks=[]
identity=read(OUT/'source/evidence-index.json')
for kind in ('frontend','native'):
    p=pathlib.Path(identity[kind]['path'])
    ok=p.is_file() and hashlib.sha256(p.read_bytes()).hexdigest()==identity[kind]['sha256']
    checks.append({'check':'current_'+kind+'_identity','passed':ok})
    if not ok:errors.append(kind+' identity mismatch')
for f in identity['recoveredFiles']:
    p=OUT/f['file']
    if not p.is_file() or hashlib.sha256(p.read_bytes()).hexdigest()!=f['sha256']:errors.append('source hash '+f['file'])
checks.append({'check':'recovered_file_hashes','count':len(identity['recoveredFiles']),'passed':not errors})
refs=0
for doc in OUT.glob('*.md'):
    text=doc.read_text(encoding='utf-8-sig')
    for m in re.finditer(r'(?<![\w-])(Novel[A-Za-z]+\.readable\.js):(\d+)(?:[–-](\d+))?',text):
        p=OUT/'source/frontend'/m[1]; refs+=1
        if not p.exists():errors.append(f'{doc.name}: missing {m[1]}');continue
        length=len(p.read_text(encoding='utf-8').splitlines())
        if int(m[2])>length or int(m[3] or m[2])>length:errors.append(f'{doc.name}: out of bounds {m[0]} ({length} lines)')
    for target in re.findall(r'\]\(([^)]+)\)',text):
        if '://' in target:continue
        clean=target.split('#')[0]
        if clean and not (doc.parent/clean).exists():errors.append(f'{doc.name}: broken link {target}')
checks.append({'check':'frontend_source_line_references','count':refs,'passed':not any('out of bounds' in e or 'missing Novel' in e for e in errors)})
inventory_doc=(OUT/'02-ui-inventory.md').read_text(encoding='utf-8')
appendix=(OUT/'08-component-contracts.md').read_text(encoding='utf-8')
missing_controls=[c['id'] for c in ui['controls'] if c['id'] not in inventory_doc or c['id'] not in appendix]
checks.append({'check':'all_controls_in_both_inventories','count':len(ui['controls']),'passed':not missing_controls})
if missing_controls:errors.append('missing controls '+str(missing_controls))
missing_components=[c['name'] for c in ui['components'] if '## '+c['name']+'\n' not in appendix]
checks.append({'check':'all_component_contracts','count':len(ui['components']),'passed':not missing_components})
if missing_components:errors.append('missing components '+str(missing_components))
prompts=read(OUT/'source/prompt-catalog.json')
for p in prompts:
    src=ROOT/p['source']
    if not src.exists() or not p['line']:errors.append('prompt missing source '+str(p))
    else:
        line=src.read_text(encoding='utf-8').splitlines()[p['line']-1]
        if src.suffix=='.json':
            entry=next((e for e in read(src)['entries'] if e['index']==p['constant']),None)
            if line.strip().rstrip(',')!='"index": '+str(p['constant']) or not entry or hashlib.sha256(entry['value'].encode()).hexdigest()!=p['sha256']:errors.append('prompt JSON source mismatch '+str(p))
        elif not line.startswith('## Constant '+str(p['constant'])+' |'):errors.append('prompt line mismatch '+str(p))
checks.append({'check':'prompt_source_headers','count':len(prompts),'passed':not any('prompt' in e for e in errors)})
ida=read(OUT/'source/ida/export-manifest.json')
for f in ida:
    p=OUT/'source/ida'/f['file']
    if f['status']!='exported' or not p.exists() or len(p.read_text(encoding='utf-8').splitlines())!=f['lines']:errors.append('IDA export invalid '+str(f))
checks.append({'check':'ida_exports','count':len(ida),'passed':not any('IDA export' in e for e in errors)})
report={'time':datetime.datetime.now(datetime.timezone(datetime.timedelta(hours=8))).isoformat(), 'scope':'document inventory and source traceability only; no runtime/model acceptance', 'checks':checks,'errors':errors,'openGapGroups':12,'paidCalls':0,'targetCodeExecuted':False}
(OUT/'source/coverage-report.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(report,ensure_ascii=False,indent=2))
raise SystemExit(bool(errors))
