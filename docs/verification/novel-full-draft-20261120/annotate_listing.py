from pathlib import Path
import json,re,struct,pefile,capstone
r=Path('docs/reverse-engineering/luoshui-novel-native-20261119');out=Path('docs/verification/novel-full-draft-20261120');values={}
for m in json.loads((r/'function-bindings.json').read_text())['modules']:
 if 'pool' not in m:continue
 for e in json.loads((r/'constants'/f"{m['module'].replace('.','--')}.json").read_text(encoding='utf8'))['entries']:values[int(m['pool'],16)+8*e['index']]=(m['module'],e['index'],e['value'])
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_64);md.detail=True
def line(i):
 refs=[]
 for op in i.operands:
  if op.type==capstone.x86.X86_OP_MEM and op.mem.base==capstone.x86.X86_REG_RIP:
   v=values.get(i.address+i.size+op.mem.disp)
   if v:refs.append(f'{v[0]}#{v[1]}='+json.dumps(v[2],ensure_ascii=False))
 return f'{i.address:#x} {i.bytes.hex():24} {i.mnemonic} {i.op_str}'+(' ; '+' ; '.join(refs) if refs else '')
for p in out.glob('*.ghidra-assembly.txt'):
 lines=[]
 for s in p.read_text(encoding='utf8').splitlines():
  m=re.match(r'(0x[0-9a-f]+) ([0-9a-f]+) ',s)
  if not m:lines.append(s);continue
  lines.extend(line(i) for i in md.disasm(bytes.fromhex(m[2]),int(m[1],16)))
 p.with_name(p.name.replace('.ghidra-assembly.txt','.full.assembly.txt')).write_text('\n'.join(lines),encoding='utf8')
