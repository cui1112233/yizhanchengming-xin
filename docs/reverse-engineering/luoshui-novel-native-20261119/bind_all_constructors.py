"""Conservative x64 constructor argument scan; static bindings, not a callgraph."""
from pathlib import Path
import json
import pefile,capstone
root=Path(__file__).resolve().parent
raw=(root.parents[2]/'_local-archive-20261028/luoshui-novel-native-20261119/novels-b5b22086.pyd').read_bytes()
pe=pefile.PE(data=raw);base=pe.OPTIONAL_HEADER.ImageBase
meta=json.loads((root/'function-bindings.json').read_text(encoding='utf8'));constants={}
for module in meta['modules']:
 if 'pool'not in module:continue
 entries=json.loads((root/'constants'/f"{module['module'].replace('.','--')}.json").read_text(encoding='utf8'))['entries']
 for e in entries:
  if isinstance(e['value'],str):constants[int(module['pool'],16)+8*e['index']]=(module['module'],e['index'],e['value'],e['fileRange'])
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_64);md.detail=True
def rip(i):
 for op in i.operands:
  if op.type==capstone.x86.X86_OP_MEM and op.mem.base==capstone.x86.X86_REG_RIP:return i.address+i.size+op.mem.disp
bindings=[];unknown=[]
for ex in pe.DIRECTORY_ENTRY_EXCEPTION:
 f=ex.struct;rows=list(md.disasm(pe.get_data(f.BeginAddress,f.EndAddress-f.BeginAddress),base+f.BeginAddress));rcx=rdx=None
 for i in rows:
  if i.mnemonic=='call':
   if i.op_str=='0x1803609a0':
    c=constants.get(rdx[0]) if rdx else None
    if c and rcx:
     bindings.append(dict(name=c[2],module=c[0],constantIndex=c[1],constantFileRange=c[3],nativeEntry=hex(rcx[0]),constructorCall=hex(i.address),argumentEvidence=[rcx[1],rdx[1]]))
    else:unknown.append(hex(i.address))
   rcx=rdx=None
   continue
  for reg in ['rcx','rdx']:
   if i.op_str.startswith(reg+',') and i.mnemonic in ['lea','mov','xor','pop']:
    a=rip(i)
    v=(a,dict(address=hex(i.address),bytes=i.bytes.hex(),instruction=i.mnemonic+' '+i.op_str)) if a is not None and ((reg=='rcx'and i.mnemonic=='lea')or(reg=='rdx'and i.mnemonic=='mov'))else None
    if reg=='rcx':rcx=v
    else:rdx=v
result=dict(bindings=bindings,unknownConstructorCalls=unknown,classification='Exact same-basic-fragment RIP-relative direct constructor arguments; unresolved values omitted, not guessed.',limitations=['No dynamic flow or runtime equivalence inferred.','PE unwind fragments can separate argument setup; these remain unknown.','Argument identity is corroborated by exported REA dossiers for selected functions only.'])
(root/'all-constructor-bindings.json').write_text(json.dumps(result,ensure_ascii=False,indent=2),encoding='utf8')
print(json.dumps(dict(bindings=len(bindings),modules=len(set(b['module']for b in bindings)),unknown=len(unknown))))
