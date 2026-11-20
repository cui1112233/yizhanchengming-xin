"""Check source offsets, constructor bindings, and inventories without target execution."""
import pathlib,json,hashlib,re,collections
OUT=pathlib.Path(__file__).resolve().parent
ROOT=OUT.parents[3]
def load(p):return json.loads(p.read_text(encoding='utf-8-sig'))
calls=load(OUT/'current-bridge-call-sites.json')
raw=pathlib.Path(calls['sourcePath']).read_bytes()
errors=[];checks=[]
assert hashlib.sha256(raw).hexdigest()==calls['sourceSha256']
for c in calls['results']:
 a,b=c['propertyRange']['utf8']
 if raw[a:b].decode('utf-8')!=c['propertyExpression']:errors.append('bad property offset '+c['method'])
 for caller in c['callers']:
  a,b=caller['range']['utf8']
  if not (0<=a<b<=len(raw)):errors.append('bad caller offset '+c['method'])
checks.append({'name':'original_property_bytes','count':len(calls['results']),'passed':not errors})
native=load(OUT/'native-entrypoint-bindings.json')['results']
mods=load(ROOT/'docs/reverse-engineering/luoshui-novel-native-20261119/function-bindings.json')['modules']
for n in native:
 m=next(m for m in mods if m['module']==n['module']);base=int(m['pool'],16)
 constants=load(ROOT/'docs/reverse-engineering/luoshui-novel-native-20261119/constants'/(m['module'].replace('.','--')+'.json'))['entries']
 indexes=[(int(s,16)-base)//8 for s in n['nameSlots']]
 if indexes!=n['constantIndexes'] or [constants[i]['value'] for i in indexes]!=[n['name'],n['qualifiedName']]:errors.append('invalid slot '+n['name'])
 f=OUT.parent/n['initializerSource'];lines=f.read_text(encoding='utf-8').splitlines();start=n['initializerLine']-1
 snippet='\n'.join(lines[start:start+8]).lower()
 if 'sub_1803609a0(' not in snippet or 'sub_'+n['nativeEntry'][2:] not in snippet:errors.append('invalid constructor source '+n['name'])
checks.append({'name':'native_constructor_slots','count':len(native),'passed':not any('slot' in e or 'constructor' in e for e in errors)})
contracts=load(OUT/'bridge-contracts.json')['methods'];expected={x['value'] for x in load(OUT/'api-name-catalog.json')}
if {c['method'] for c in contracts}!=expected:errors.append('method coverage differs')
for c in contracts:
 if not (OUT.parent/c['contractFile']).exists():errors.append('missing contract '+c['method'])
 if c['frontendFunctionCandidates']==0 and c['method'] not in {'novels_get_book_rules','novels_save_book_rules'}:errors.append('unexpected missing frontend '+c['method'])
checks.append({'name':'all_46_contracts','count':len(contracts),'passed':not any('contract' in e or 'coverage' in e or 'frontend' in e for e in errors)})
exports=load(OUT/'ida/entrypoints-manifest.json')
for e in exports:
 p=OUT/'ida'/e.get('file','')
 if e['status']!='exported' or not p.is_file() or len(p.read_text(encoding='utf-8').splitlines())!=e.get('lines'):errors.append('IDA export invalid '+e['address'])
checks.append({'name':'new_ida_exports','count':len(exports),'passed':not any('IDA' in e for e in errors)})
snapshot=load(OUT/'cm-studio-routing-snapshot.json')
if hashlib.sha256((ROOT/snapshot['source']).read_bytes()).hexdigest()!=snapshot['sha256']:errors.append('CM source drift; compare again')
checks.append({'name':'cm_source_snapshot_current','passed':not any('CM source' in e for e in errors)})
report={'scope':'static source/traceability, not runtime equivalence','checks':checks,'errors':errors,'rawPropertyAccesses':len(calls['results']),'classifications':dict(collections.Counter(c['classification'] for c in calls['results'])),'frontendMethodCandidates':len({c['method'] for c in calls['results'] if c['classification']=='async_bridge_candidate'}),'publicNativeMethods':len(contracts),'unresolvedQanProperties':len(calls['unknown']),'sourceSha256':calls['sourceSha256']}
(OUT/'bridge-verification.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(report,ensure_ascii=False,indent=2))
raise SystemExit(bool(errors))
