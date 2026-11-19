"""Render retained REA pseudocode and annotate proven module constant addresses."""
from pathlib import Path
import bisect,json,re
root=Path(__file__).resolve().parent
bundle=json.loads((root.parents[2]/'_local-archive-20261028/luoshui-novel-native-20261119/rea-native-chain-evidence.json').read_text(encoding='utf-8'))
modules=json.loads((root/'function-bindings.json').read_text(encoding='utf-8'))['modules']
values={}
for m in modules:
 if 'pool'not in m:continue
 constants=json.loads((root/'constants'/f"{m['module'].replace('.','--')}.json").read_text(encoding='utf-8'))['entries']
 for e in constants:
  values[int(m['pool'],16)+8*e['index']]=(m['module'],e['index'],e['value'])
out=root/'native';out.mkdir(exist_ok=True)
index=[];bindings=[]
for file in [root/'review-length.rea.json',root/'review-pass.rea.json']:
 if file.exists():bundle['records'].append(json.loads(file.read_text(encoding='utf8')))
for r in bundle['records']:
 if r['operation']!='analyze_function':continue
 n=r['normalized_result'];address=n['procedure']['address'];text=n['pseudocode'].replace('\r','')
 path=out/f'{address}.c';path.write_text(text,encoding='utf-8',newline='\n')
 lines=[f"/* REA {r['evidence_id']}; original SHA b5b22086...; constant comments are analyst annotations. */"]
 for line in text.splitlines():
  refs=[]
  for a in re.findall(r'DAT_([0-9a-f]+)',line):
   value=values.get(int(a,16))
   if value:
    module,num,v=value
    if isinstance(v,str):refs.append(f'{module}#{num}='+json.dumps(v[:240],ensure_ascii=False))
  if refs:lines.append('/* '+ ' ; '.join(dict.fromkeys(refs)).replace('*/','* /') +' */')
  lines.append(line)
 annotated=out/f'{address}.annotated.c';annotated.write_text('\n'.join(lines),encoding='utf-8',newline='\n')
 index.append(dict(address=address,evidenceId=r['evidence_id'],source=f'native/{address}.c',annotated=f'native/{address}.annotated.c',lines=len(text.splitlines()),body=n['procedure']['body'],limitations=r.get('limitations',[])))
 for f,c in re.findall(r'FUN_1803609a0\s*\(\s*FUN_([0-9a-f]+)\s*,\s*DAT_([0-9a-f]+)',text):
  value=values.get(int(c,16))
  if value:bindings.append(dict(nativeEntry='0x'+f,module=value[0],constantIndex=value[1],name=value[2],evidenceId=r['evidence_id']))
for file in out.glob('*.fallback.c'):
 lines=[]
 for line in file.read_text(encoding='utf8').splitlines():
  refs=[]
  for a in re.findall(r'DAT_([0-9a-f]+)',line):
   value=values.get(int(a,16))
   if value:
    module,num,v=value
    if isinstance(v,str):refs.append(f'{module}#{num}='+json.dumps(v[:240],ensure_ascii=False))
  if refs:lines.append('/* '+' ; '.join(dict.fromkeys(refs)).replace('*/','* /')+' */')
  lines.append(line)
 file.with_suffix('.annotated.c').write_text('\n'.join(lines),encoding='utf8',newline='\n')
(root/'native-index.json').write_text(json.dumps(dict(functions=index,constructorBindings=bindings),ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(dict(functions=len(index),constructorBindings=len(bindings))))
