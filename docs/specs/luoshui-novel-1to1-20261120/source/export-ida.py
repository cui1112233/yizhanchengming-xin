import idapro, ida_hexrays, ida_funcs, ida_name, json, pathlib, time
out=pathlib.Path('F:/脚本测试/一战晟铭-v88/docs/specs/luoshui-novel-1to1-20261120/source/ida')
out.mkdir(parents=True,exist_ok=True)
rc=idapro.open_database('C:/rs-ida/novels-sdk.pyd.i64',True)
print('OPEN',rc,flush=True)
results=[]
if rc==0:
 try:
  ready=ida_hexrays.init_hexrays_plugin()
  for ea in (0x180165140,0x180157690,0x1800d76c0):
   print('DECOMPILING',hex(ea),flush=True)
   try:
    c=ida_hexrays.decompile(ea)
    text=str(c) if c else ''
    name=hex(ea)+'.c'
    (out/name).write_text(text,encoding='utf-8')
    results.append({'address':hex(ea),'file':name,'lines':len(text.splitlines()),'bytes':len(text.encode()),'status':'exported' if text else 'empty'})
   except Exception as ex:
    results.append({'address':hex(ea),'status':'failed','error':str(ex)})
   (out/'export-manifest.json').write_text(json.dumps(results,indent=2),encoding='utf-8')
   print(results[-1],flush=True)
 finally:
  idapro.close_database(True)
  print('SAVED',flush=True)
