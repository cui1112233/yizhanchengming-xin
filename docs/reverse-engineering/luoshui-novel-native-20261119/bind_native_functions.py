"""Read-only x64 constructor bindings. Does not execute or import the target."""
from pathlib import Path
import bisect,hashlib,json,re,struct
import capstone,pefile
root=Path(__file__).resolve().parent
raw=(root.parents[2]/'_local-archive-20261028/luoshui-novel-native-20261119/novels-b5b22086.pyd').read_bytes()
pe=pefile.PE(data=raw);base=pe.OPTIONAL_HEADER.ImageBase
text=next(s for s in pe.sections if s.Name.rstrip(b'\0')==b'.text');data=text.get_data();tb=base+text.VirtualAddress
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_64);md.detail=True
owners=sorted((base+e.struct.BeginAddress,base+e.struct.EndAddress) for e in pe.DIRECTORY_ENTRY_EXCEPTION)
begins=[x[0] for x in owners]
def owner(a):
 i=bisect.bisect_right(begins,a)-1
 return owners[i] if i>=0 and a<owners[i][1] else None
def rip(ins):
 for op in ins.operands:
  if op.type==capstone.x86.X86_OP_MEM and op.mem.base==capstone.x86.X86_REG_RIP:return ins.address+ins.size+op.mem.disp
def row(i):return dict(address=hex(i.address),bytes=i.bytes.hex(),mnemonic=i.mnemonic,operands=i.op_str,ripTarget=hex(rip(i)) if rip(i)!=None else None)
bundle=json.loads((root.parents[2]/'_local-archive-20261028/luoshui-novel-native-20261119/rea-native-module-evidence.json').read_text(encoding='utf-8'))
full_bodies={r['normalized_result']['procedure']['address']:r['normalized_result']['procedure']['body']['ranges'] for r in bundle['records'] if r['operation']=='analyze_function'}
manifest=json.loads((root/'manifest.json').read_text(encoding='utf-8'));modules=[]
# Compiler-generated loader: r8=name,rdx=module constants array, call loadConstantsBlob.
loader=0x180388a30;constructor=0x1803609a0
for entry in manifest['modules']:
 name=entry['module'];needle=name.encode()+b'\0';found=[]
 for m in re.finditer(re.escape(needle),raw[:3879424]):
  va=base+pe.get_rva_from_offset(m.start())
  for lm in re.finditer(b'\x4c\x8d\x05',data):
   at=lm.start()
   if tb+at+7+struct.unpack_from('<i',data,at+3)[0]!=va:continue
   instructions=list(md.disasm(data[at:at+45],tb+at));pool=None;call=None
   for i in instructions:
    if i.mnemonic=='lea' and i.op_str.startswith('rdx,'):pool=rip(i)
    if i.mnemonic=='call':
     if i.op_str==hex(loader):call=i.address
     break
   if pool and call:found.append(dict(pool=pool,initializer=owner(tb+at),loaderCall=hex(call),instructions=[row(i)for i in instructions[:7]]))
 if len(found)!=1:modules.append(dict(module=name,unknown='expected one constant-loader binding',candidates=found));continue
 f=found[0];const=json.loads((root/entry['path']).read_text(encoding='utf-8'))['entries'];pool=f['pool'];start,end=f['initializer'];block=pe.get_data(start-base,end-start);ranges=full_bodies.get(hex(start),[{'start':hex(start),'end':hex(end-1)}]);instructions=[i for span in ranges for i in md.disasm(pe.get_data(int(span['start'],16)-base,int(span['end'],16)-int(span['start'],16)+1),int(span['start'],16))];bindings=[]
 for pos,i in enumerate(instructions):
  if i.mnemonic!='call' or i.op_str!=hex(constructor):continue
  window=instructions[max(0,pos-35):pos]
  rcx=rdx=None
  for x in window:
   if x.mnemonic in ['lea','mov'] and x.op_str.startswith('rcx,'):rcx=rip(x) if x.mnemonic=='lea' else None
   if x.mnemonic in ['lea','mov'] and x.op_str.startswith('rdx,'):rdx=rip(x) if x.mnemonic=='mov' else None
  index=(rdx-pool)//8 if rdx is not None else -1
  if rdx is not None and rdx>=pool and (rdx-pool)%8==0 and index<len(const) and isinstance(const[index]['value'],str) and rcx is not None and owner(rcx):
   bindings.append(dict(name=const[index]['value'],constantIndex=index,constantFileRange=const[index]['fileRange'],nativeEntry=hex(rcx),constructorCall=hex(i.address),instructions=[row(x)for x in window[-16:]]))
 modules.append(dict(module=name,pool=hex(pool),initializer=[hex(start),hex(end)],loaderCall=f['loaderCall'],loaderInstructions=f['instructions'],bindings=bindings,coverage='direct compiler constructor bindings only; not dynamic callgraph'))
out=dict(sha256=hashlib.sha256(raw).hexdigest(),constructor=hex(constructor),loader=hex(loader),tool='capstone '+capstone.__version__,modules=modules,limitations=['Constructor bindings are static observations, not proof of runtime branches.','Only RIP-relative direct name/code arguments are admitted; computed values remain unknown.','Helper identities are corroborated against REA Ghidra module dossiers, not Python source.'])
(root/'function-bindings.json').write_text(json.dumps(out,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(dict(modules=len(modules),boundModules=sum('bindings'in x for x in modules),bindings=sum(len(x.get('bindings',[]))for x in modules))))
for m in modules:
 if m['module'] in ['novels.pipeline.chapter_review_cycle','novels.prompts.inkos.reviser_prompts','novels.prompts.inkos.writer_prompts']:
  print(m['module'],[(x['name'],x['nativeEntry'])for x in m.get('bindings',[])])
