"""Inventory and decode the immutable original snapshot; no target code is executed."""
from pathlib import Path
import hashlib,json,re,struct,subprocess,sys

root=Path(__file__).resolve().parent
target=root.parents[2]/'_local-archive-20261028/luoshui-novel-native-20261119/novels-b5b22086.pyd'
raw=target.read_bytes();sha=hashlib.sha256(raw).hexdigest()
assert sha=='b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d'
modules=[]
for m in re.finditer(rb'novels(?:\.[A-Za-z_][A-Za-z0-9_]*)*\x00',raw):
    if m.end()+6>=len(raw):continue
    size=struct.unpack_from('<I',raw,m.end())[0]
    if not 3<size<500000 or m.end()+4+size>len(raw):continue
    if raw[m.end()+3+size]!=46:continue
    name=m.group()[:-1].decode()
    slug=name.replace('.','--')
    run=subprocess.run([sys.executable,str(root/'decode_constants.py'),'--module',name,'--slug',slug,'--module-offset',str(m.start())],capture_output=True,text=True,encoding='utf-8')
    path=root/'constants'/f'{slug}.json'
    if run.returncode or not path.exists():
        modules.append(dict(module=name,error=run.stderr[-400:],complete=False));continue
    result=json.loads(path.read_text(encoding='utf-8'))
    entry=dict(module=name,path=f'constants/{slug}.json',complete=result['complete'],declared=result['declaredConstants'],parsed=result['parsedConstants'],payloadRange=result['payloadRange'],error=result['error'])
    # Literal-only rendering is deliberately not called source or assembled system prompt.
    lines=[f'# {name}: recovered constant literals',f'Source SHA256: {sha}', 'These are ordered constants, not reconstructed function bodies.\n']
    literals=[];seen=set()
    def walk(node):
        if isinstance(node,list):
            for child in node:walk(child)
        elif isinstance(node,dict):
            if 'fileRange' in node and isinstance(node.get('value'),str):
                key=tuple(node['fileRange'])
                if key not in seen:
                    seen.add(key);literals.append(node)
            for key,child in node.items():
                if key!='previous':walk(child)
    walk(result['entries'])
    for e in literals:
        if isinstance(e['value'],str):
            text=e['value']
            if len(text)>=80 or ('prompt' in text.lower() or 'run_' in text):
                lines.extend([f"## Constant {e.get('index','nested')} | bytes {e['fileRange'][0]}..{e['fileRange'][1]-1}",'```text',text,'```',''])
    readable=root/'readable';readable.mkdir(exist_ok=True)
    (readable/f'{slug}.md').write_text('\n'.join(lines),encoding='utf-8')
    modules.append(entry)
summary=dict(source=str(target),sha256=sha,moduleCount=len(modules),complete=sum(x['complete'] for x in modules),modules=modules,limitations=['Constant recovery is not source decompilation.','Function binding, branch order and assembled prompt must be proved separately.','No target import/execution, model requests or app writes.'])
(root/'manifest.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps({k:summary[k] for k in ['sha256','moduleCount','complete']}))
print(json.dumps([x for x in modules if not x['complete']],ensure_ascii=False))
