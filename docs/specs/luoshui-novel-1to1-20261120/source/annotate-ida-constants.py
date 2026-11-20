"""Analyst comments from verified constant slots; original IDA exports stay intact."""
import json,pathlib,re
ROOT=pathlib.Path(__file__).resolve().parents[4]
OUT=pathlib.Path(__file__).resolve().parent
NATIVE=ROOT/'docs/reverse-engineering/luoshui-novel-native-20261119'
mods=json.loads((NATIVE/'function-bindings.json').read_text(encoding='utf-8'))['modules']
slots={}
for m in mods:
 f=NATIVE/'constants'/(m['module'].replace('.','--')+'.json')
 if not f.exists():continue
 entries=json.loads(f.read_text(encoding='utf-8'))['entries'];base=int(m['pool'],16)
 for e in entries:
  if 'value' in e:slots[base+8*e['index']]={'module':m['module'],'index':e['index'],'range':e['fileRange'],'text':e['value']}
bindings=json.loads((OUT/'native-entrypoint-bindings.json').read_text(encoding='utf-8'))['results']
names={int(b['nativeEntry'],16):b['qualifiedName'] for b in bindings}
dest=OUT/'ida-annotated';dest.mkdir(exist_ok=True)
report=[]
for f in sorted((OUT/'ida').glob('*.c')):
 lines=['// Analyst annotations only; original export: source/ida/'+f.name,
        '// qword comments derive module constant slot + original file range; not runtime values.']
 hits=0
 for i,line in enumerate(f.read_text(encoding='utf-8').splitlines(),1):
  annotations=[]
  for addr in set(re.findall(r'qword_([A-Fa-f0-9]+)',line)):
   e=slots.get(int(addr,16))
   if e:
    t=json.dumps(e['text'],ensure_ascii=False)
    if len(t)>240:t=t[:240]+'... (truncated literal)'
    annotations.append(f"qword_{addr} -> {e['module']}#{e['index']} bytes {e['range']}: {t}")
  for addr in set(re.findall(r'sub_([A-Fa-f0-9]+)',line)):
   name=names.get(int(addr,16))
   if name:annotations.append('sub_'+addr+' -> '+name)
  if annotations:
   lines.append('// Original IDA line '+str(i)+' | '+'; '.join(annotations));hits+=len(annotations)
  lines.append(line)
 (dest/f.name).write_text('\n'.join(lines)+'\n',encoding='utf-8')
 report.append({'file':f.name,'annotations':hits,'originalLines':len(f.read_text(encoding='utf-8').splitlines())})
(dest/'annotation-manifest.json').write_text(json.dumps(report,indent=2),encoding='utf-8')
print(json.dumps(report))
