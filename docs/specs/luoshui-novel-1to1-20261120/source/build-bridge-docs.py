"""Render exact current bridge expressions; do not infer backend success."""
import pathlib,json,re,collections,hashlib
OUT=pathlib.Path(__file__).resolve().parent
def load(p):return json.loads(p.read_text(encoding='utf-8-sig'))
def esc(s):return str(s).replace('|','\\|').replace('\n','<br>')
calls=load(OUT/'current-bridge-call-sites.json')
native=load(OUT/'native-entrypoint-bindings.json')['results']
api=load(OUT/'api-name-catalog.json')
dest=OUT/'bridge-contracts';dest.mkdir(exist_ok=True)
rows=[];report=[]
for item in api:
 name=item['value'];entries=[r for r in calls['results'] if r['method']==name and r['classification']=='async_bridge_candidate']
 bindings=[b for b in native if b['module']=='novels.api.novels_api' and b['name']==name]
 assert len(bindings)==1,(name,bindings)
 b=bindings[0]
 text=['// Inspection only; target source is never executed.', '// Original frontend SHA256: '+calls['sourceSha256'], '// Native entry '+b['nativeEntry']+'; constructor '+b['initializerSource']+':'+str(b['initializerLine'])]
 unique={}
 for e in entries:unique.setdefault(tuple(e['functionRange']['utf16']),e)
 entryList=list(unique.values())
 for i,e in enumerate(entryList,1):
  text+=['',f'// Frontend candidate {i}; original UTF16 {e["functionRange"]["utf16"]}; property UTF8 {e["propertyRange"]["utf8"]}; component {e["component"]}',e['source']]
  for j,c in enumerate(e['callers'],1):
   text += ['',f'// Caller {j}; {c["component"]}/{c["callerName"]}; UTF16 {c["range"]["utf16"]}', '// Exact argument expressions: '+json.dumps(c['arguments'],ensure_ascii=False)]
   if c['callerSource']:text.append(c['callerSource'])
 if not entryList:text += ['// No frontend property candidate was observed in the resolved Qan scope.', '// This is not proof of global absence; native implementation is separately exported.']
 file=name+'.inspection.js'
 (dest/file).write_text('\n'.join(text)+'\n',encoding='utf-8')
 params=sorted(set('('+', '.join(e['parameters'])+')' for e in entryList))
 sources=sorted(set(e['component'] or '共享桥接函数 '+str(e['functionName']) for e in entryList))
 exprs=[]
 for e in entryList:
  for c in e['callers']:
   expr='; '.join(c['arguments']) if c['arguments'] else '无显式实参'
   if expr not in exprs:exprs.append(expr)
 callerText='\n'.join(c['callerSource'] or '' for e in entryList for c in e['callers'])
 fields=sorted(set(re.findall(r'\b(?:success|data|message|cancelled|action|awaitingConfirmation|storyId|bookId|pendingAction)\b',callerText)))
 status='前端异步候选＋原生入口已绑定' if entryList else '未观察到前端调用；有原生函数体'
 rows.append(f'| `{name}` | {esc(" / ".join(sources) or "未观察到")} | `{esc(" / ".join(params) or "未恢复前端签名")}` | `{b["nativeEntry"]}` | [原式](source/bridge-contracts/{file}) | {status} |')
 report.append({'method':name,'nativeEntry':b['nativeEntry'],'nativeConstructorSource':b['initializerSource'],'nativeConstructorLine':b['initializerLine'],'nameConstantIndex':b['constantIndexes'][0],'frontendFunctionCandidates':len(entryList),'frontendComponentCandidates':sources,'parameters':params,'argumentExpressions':exprs,'responseFieldCandidates':fields,'contractFile':'source/bridge-contracts/'+file,'status':status,'limitations':['static candidate only; backend signature and full branch behavior are separate','response field names are syntax candidates, not control/dataflow proof']})
doc=['# 09 当前版接口调用合同补证\n',
 '本轮基于当前bundle `7c252056…`，独立重写白名单字面量解释器，不执行原bundle/helper、不导入目标模块。解码表zan共565项，Qan下标减267，旋转278次后原校验式=299251。原式、UTF16/UTF8范围和调用者均保留。\n',
 '## 一、名称存在与实际调用分开\n',
 '扫描得到57个novels_属性访问，覆盖44个名称。其中8个支路的调用参数含未解析`_0x…`占位，另1个非异步脚手架分支暂不算有效调用。其余48个异步候选包含文件选择的guard和调用重复，不能称为48个独立功能。全部46个公开API均通过IDA构造参数绑定到真实原生地址；不代表46个函数体都已读完。\n',
 '当前Qan相关未解析属性为0，不等于整个17MB前端所有动态属性都已解完；其他字符串表和动态对象关系不在该结论范围。\n',
 '| 原接口 | 前端组件/共享函数 | 参数原式 | IDA入口 | 完整调用片段 | 状态 |\n|---|---|---|---|---|---|']
doc+=rows
doc += ['\n## 二、输入与返回原式\n',
 '每项inspection文件保留原桥接函数、真实调用者全文和参数表达式；以下表仅做索引，不把局部变量名擅自解释成后台签名。返回字段是调用者语法候选，需看完整success/cancelled分支。\n',
 '| 接口 | 调用实参原式 | 出现的响应字段候选 |\n|---|---|---|']
for r in report:doc.append(f'| `{r["method"]}` | {esc("<br>".join(r["argumentExpressions"]) or "前端未观察到；原生体另查")} | {esc(", ".join(r["responseFieldCandidates"]) or "未观察到/动态访问")} |')
doc += ['\n## 三、原生构造绑定的可重复证据\n',
 'API模块常量池基址`0x1803aff50`；IDA初始化函数`0x180091e60`调用构造器`0x1803609a0`。从每个构造器的函数指针、名称槽、限定名称槽，按`(槽地址-模块池基址)/8`取已解码常量，得到真实方法名和原生入口。每项保留常量索引、原文件字节范围、初始化C行号，见`source/native-entrypoint-bindings.json`。这比搜索到一个方法字符串多了一层真实函数绑定，仍不证明运行调用成功。\n',
 '会话服务同法绑定13个方法；pipeline_runner暂未匹配此构造器形态，不能推定该模块没有方法。\n',
 '## 四、两个规则接口的结论\n',
 '- `novels_get_book_rules`绑定`0x1800888e0`，IDA函数体490行：加载models.book_rules的load_book_rules，取_state().story_dir(book_id)，成功返回success/data并调用to_dict。原始C行96–150，注释版`source/ida-annotated/0x1800888e0.c`。',
 '- `novels_save_book_rules`绑定`0x180089120`，IDA函数体482行：BookRules.from_dict后传给save_book_rules，并使用作品story_dir。原始C行95–166，注释版同目录。',
 '- 当前已解析前端范围未观察到这两个属性调用。应登记“后端存在、前端调用未知”，不能判为原版假功能或自行新增两个按钮。',
 '- 参数与错误传播的完全恢复仍需结合两个函数体全部异常出口，尚不能仅按这些load名称宣称规则完整生效。\n',
 '## 五、本阶段验收边界\n',
 'G01的名称/前端参数原式/调用者/原生入口对照已经落盘，实际后端签名、全部异常和运行验收不在“静态覆盖通过”中。后续G02/G03继续从send_message和confirm_action追进AgentSessionService；G08再按最终请求核对提示词实际使用。']
(OUT.parent/'09-current-bridge-contracts.md').write_text('\n'.join(doc)+'\n',encoding='utf-8')
(OUT/'bridge-contracts.json').write_text(json.dumps({'sourceSha256':calls['sourceSha256'],'methods':report,'excludedBranches':[r for r in calls['results'] if r['classification']!='async_bridge_candidate']},ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps({'methods':len(report),'withAsyncFrontendCandidate':sum(r['frontendFunctionCandidates']>0 for r in report),'nativeEntries':len(set(r['nativeEntry'] for r in report))}))
