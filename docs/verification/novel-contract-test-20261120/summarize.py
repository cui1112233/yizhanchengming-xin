import pathlib,json,re,datetime,collections
ROOT=pathlib.Path(__file__).resolve().parents[3]
OUT=pathlib.Path(__file__).resolve().parent
report={}
for name in ('backend','http'):
 p=OUT/(name+'.jsonl')
 if not p.exists():continue
 rows=[]
 for line in p.read_text(encoding='utf-8-sig').splitlines():
  try:rows.append(json.loads(line))
  except json.JSONDecodeError:pass
 completed=[r for r in rows if r.get('Test') and r['Action'] in ('pass','fail','skip')]
 report[name]={'testResults':dict(collections.Counter(r['Action'] for r in completed)),'packages':[r for r in rows if not r.get('Test') and r['Action'] in ('pass','fail')],'failedTests':[r['Test'] for r in completed if r['Action']=='fail'],'skippedTests':[r['Test'] for r in completed if r['Action']=='skip']}
tap=OUT/'frontend.tap'
if tap.exists():
 text=tap.read_text(encoding='utf-8-sig');summary=dict(re.findall(r'^(?:# |ℹ )(tests|pass|fail|cancelled|skipped|duration_ms) (.+)$',text,re.M))
 failures=[]
 if '✖ failing tests:' in text:
  tail=text.split('✖ failing tests:',1)[1]
  for m in re.finditer(r'^test at (.+)\n✖ (.+)\n([\s\S]*?)(?=^test at |\Z)',tail,re.M):
   first=m[3].strip().splitlines()[0] if m[3].strip() else ''
   failures.append({'location':m[1],'name':m[2],'error':first,'classification':'browser inspection timeout' if 'test page inspection timed out' in m[3] else 'assertion failure; product versus fixture requires review'})
 report['frontend']={'summary':summary,'complete':'tests' in summary,'failedTests':failures}
cut=datetime.datetime(2026,11,20,17,20,42).timestamp()
observed=[]
base=ROOT/'_local-archive-20261028/novel-task7-browser'
for p in base.iterdir():
 if not p.is_dir() or p.stat().st_ctime<cut:continue
 f=p/'result.html'
 if not f.exists():continue
 s=f.read_text(encoding='utf-8');m=re.search(r'data-result="([^"]+)"',s);e=re.search(r'data-error="([^"]+)"',s)
 observed.append({'case':p.name,'result':m[1] if m else 'unknown','error':e[1][:400] if e else None})
report['mockBrowserArtifacts']={'completedCases':len(observed),'failed':[o for o in observed if o['result']!='passed']}
(OUT/'summary.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps({k:{a:b for a,b in v.items() if a not in ('skippedTests','packages')} for k,v in report.items()},ensure_ascii=False,indent=2))
