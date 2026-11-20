"""Bind IDA's native constructor arguments to recovered immutable constant slots."""
import pathlib,json,re
ROOT=pathlib.Path(__file__).resolve().parents[4]
OUT=pathlib.Path(__file__).resolve().parent
original=ROOT/'docs/reverse-engineering/luoshui-novel-native-20261119'
modules=json.loads((original/'function-bindings.json').read_text(encoding='utf-8'))['modules']
targets={'novels.api.novels_api':'0x180091e60','novels.services.agent_session':'0x180217350','novels.pipeline.pipeline_runner':'0x1801308a0'}
results=[]
for name,entry in targets.items():
 m=next(m for m in modules if m['module']==name)
 pool=int(m['pool'],16)
 constants=json.loads((original/'constants'/(name.replace('.','--')+'.json')).read_text(encoding='utf-8'))['entries']
 text=(OUT/'ida'/(entry+'.c')).read_text(encoding='utf-8')
 for match in re.finditer(r'sub_1803609A0\(\s*(?:\(int\))?sub_([A-Fa-f0-9]+),\s*qword_([A-Fa-f0-9]+),\s*qword_([A-Fa-f0-9]+)',text):
  slots=[int(match[i],16) for i in (2,3)]
  indexes=[(slot-pool)//8 if (slot-pool)%8==0 else -1 for slot in slots]
  if not all(0<=i<len(constants) for i in indexes):continue
  values=[constants[i].get('value') for i in indexes]
  if not all(isinstance(v,str) for v in values):continue
  results.append({'module':name,'nativeEntry':'0x'+match[1].lower(),'name':values[0],'qualifiedName':values[1],'nameSlots':[hex(s) for s in slots],'constantIndexes':indexes,'constantFileRanges':[constants[i]['fileRange'] for i in indexes],'initializerSource':'source/ida/'+entry+'.c','initializerLine':text[:match.start()].count('\n')+1,'evidenceType':'IDA constructor argument + decoded module constant slot; not runtime callgraph'})
(OUT/'native-entrypoint-bindings.json').write_text(json.dumps({'targetSha256':'b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d','constructor':'0x1803609a0','results':results},ensure_ascii=False,indent=2),encoding='utf-8')
for r in results:print(r['module'],r['name'],r['nativeEntry'])
