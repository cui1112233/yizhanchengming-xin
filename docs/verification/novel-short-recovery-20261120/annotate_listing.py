from pathlib import Path
import json,re,struct,pefile,capstone
r=Path('docs/reverse-engineering/luoshui-novel-native-20261119');out=Path('docs/verification/novel-short-recovery-20261120');values={}
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
raw=Path('_local-archive-20261028/luoshui-novel-native-20261119/novels-b5b22086.pyd').read_bytes();pe=pefile.PE(data=raw);base=pe.OPTIONAL_HEADER.ImageBase;s=next(s for s in pe.sections if s.Name.startswith(b'.text'));data=s.get_data();tb=base+s.VirtualAddress
wanted={a for a,v in values.items() if v[0]=='novels.pipeline.short_fiction_parsers' and v[1] in [136,137,185,186]};lines=[]
for m in re.finditer(rb'[\x48\x4c][\x8b\x89].',data):
 at=m.start()
 if data[at+2]&0xc7!=5:continue
 target=tb+at+7+struct.unpack_from('<i',data,at+3)[0]
 if target not in wanted:continue
 lines.append('# selected constant reference')
 lines.extend(line(i) for i in md.disasm(data[at:at+95],tb+at))
(out/'parser-module-selected-bindings.assembly.txt').write_text('\n'.join(lines),encoding='utf8')
print(len(lines))
