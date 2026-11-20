"""Use installed IDA SDK for static analysis only; never import the target module."""
import pathlib,json,sys
import idapro

out=pathlib.Path(__file__).resolve().parent/'ida'
rc=idapro.open_database('C:/rs-ida/novels-sdk.pyd.i64',True)
print('OPEN',rc,flush=True)
if rc: raise SystemExit(rc)
manifest=out/'entrypoints-manifest.json'
results=json.loads(manifest.read_text(encoding='utf-8')) if manifest.exists() else []
try:
 import ida_hexrays,ida_funcs,ida_name,idautils
 if not ida_hexrays.init_hexrays_plugin():raise RuntimeError('HexRays unavailable')
 for raw in sys.argv[1:]:
  ea=int(raw,16)
  print('DECOMPILING',hex(ea),flush=True)
  results=[r for r in results if r['address']!=hex(ea)]
  try:
   text=str(ida_hexrays.decompile(ea))
   file=hex(ea)+'.c';(out/file).write_text(text,encoding='utf-8')
   calls=[]
   for insn in idautils.FuncItems(ea):
    for dest in idautils.CodeRefsFrom(insn,False):
     f=ida_funcs.get_func(dest)
     if f and f.start_ea==dest:calls.append({'from':hex(insn),'to':hex(dest),'name':ida_name.get_name(dest)})
   results.append({'address':hex(ea),'file':file,'lines':len(text.splitlines()),'bytes':len(text.encode()),'status':'exported','codeReferences':calls})
  except Exception as e:results.append({'address':hex(ea),'status':'failed','error':str(e)})
  (out/'entrypoints-manifest.json').write_text(json.dumps(results,indent=2),encoding='utf-8')
  print({k:v for k,v in results[-1].items() if k!='codeReferences'},flush=True)
finally:
 idapro.close_database(True)
 print('SAVED',flush=True)
