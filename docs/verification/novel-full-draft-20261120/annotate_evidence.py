from pathlib import Path
import json,re,hashlib
import pefile,capstone
root=Path('docs/reverse-engineering/luoshui-novel-native-20261119');out=Path('docs/verification/novel-full-draft-20261120')
values={}
for m in json.loads((root/'function-bindings.json').read_text())['modules']:
 if 'pool' not in m:continue
 for e in json.loads((root/'constants'/f"{m['module'].replace('.','--')}.json").read_text(encoding='utf8'))['entries']:values[int(m['pool'],16)+8*e['index']]=(m['module'],e['index'],e['value'])
for file in out.glob('*.fallback.c'):
 lines=[]
 for line in file.read_text().splitlines():
  refs=[]
  for a in re.findall(r'DAT_([0-9a-f]+)',line):
   v=values.get(int(a,16))
   if v:refs.append(f'{v[0]}#{v[1]}='+json.dumps(v[2],ensure_ascii=False)[:500])
  if refs:lines.append('/* '+' ; '.join(dict.fromkeys(refs)).replace('*/','* /')+' */')
  lines.append(line)
 file.with_suffix('.annotated.c').write_text('\n'.join(lines),encoding='utf8')
raw=Path('_local-archive-20261028/luoshui-novel-native-20261119/novels-b5b22086.pyd').read_bytes(); pe=pefile.PE(data=raw);base=pe.OPTIONAL_HEADER.ImageBase
assert hashlib.sha256(raw).hexdigest()=='b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d'
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_64);md.detail=True
for addr in [0x18014e500,0x180150000,0x18014f130,0x18014f860,0x180151160,0x180152350,0x1801533f0,0x180159910,0x1801d6160,0x1801d7d70,0x1801d8730,0x1801d5660,0x180155eb0,0x18014e1c0]:
 entry=next(e for e in pe.DIRECTORY_ENTRY_EXCEPTION if base+e.struct.BeginAddress==addr);end=base+entry.struct.EndAddress;lines=[]
 for i in md.disasm(pe.get_data(addr-base,end-addr),addr):
  refs=[]
  for op in i.operands:
   if op.type==capstone.x86.X86_OP_MEM and op.mem.base==capstone.x86.X86_REG_RIP:
    a=i.address+i.size+op.mem.disp;v=values.get(a)
    if v:refs.append(f'{v[0]}#{v[1]}='+json.dumps(v[2],ensure_ascii=False)[:500])
  lines.append(f'{i.address:#x} {i.bytes.hex():24} {i.mnemonic} {i.op_str}'+ (' ; '+' ; '.join(refs) if refs else ''))
 (out/f'{addr:#x}.assembly.txt').write_text('\n'.join(lines),encoding='utf8')
