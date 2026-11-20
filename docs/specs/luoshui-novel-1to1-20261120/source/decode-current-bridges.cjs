// Static analysis only: no eval, VM, target imports, or target helper execution.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const req=require('node:module').createRequire(path.resolve('frontend/package.json'));
const parser=req('@babel/parser'),generate=req('@babel/generator').default;
const root=__dirname,sourcePath='F:/软件/洛水2026/洛水2026/dist/assets/index-DgTjggsH.js';
const source=fs.readFileSync(sourcePath,'utf8'),ast=parser.parse(source,{sourceType:'unambiguous',attachComment:false});
const nodes=[],parent=new WeakMap();
function walk(n,p){if(!n?.type)return;nodes.push(n);if(p)parent.set(n,p);for(const[k,v]of Object.entries(n)){if(['loc','extra'].includes(k))continue;if(Array.isArray(v))v.forEach(x=>walk(x,n));else if(v?.type)walk(v,n);}}
walk(ast);
const defs=nodes.filter(n=>n.type==='FunctionDeclaration');
const isScope=n=>n&&(['Program','BlockStatement','CatchClause'].includes(n.type)||/Function/.test(n.type));
function scopeAt(n){while(n&&!isScope(n))n=parent.get(n);return n;}
const bindings=new WeakMap();
function bind(scope,name,node){if(!scope)return;let m=bindings.get(scope);if(!m)bindings.set(scope,m=new Map());if(!m.has(name))m.set(name,node);}
for(const n of nodes){
 if(n.type==='VariableDeclarator'&&n.id.type==='Identifier'){
  let s=scopeAt(parent.get(n));if(parent.get(n)?.kind==='var')while(s&&s.type!=='Program'&&!/Function/.test(s.type))s=scopeAt(parent.get(s));
  bind(s,n.id.name,n);
 }
 if(n.type==='FunctionDeclaration')bind(scopeAt(parent.get(n)),n.id.name,n);
}
const findBinding=(name,at,frames)=>{for(let s=scopeAt(at);s;s=scopeAt(parent.get(s))){if(frames.get(s)?.has(name))return {value:frames.get(s).get(name)};const n=bindings.get(s)?.get(name);if(n)return {node:n};}throw Error('unbound '+name);};
function base64(s){const alphabet='abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/=';let raw='',acc=0,count=0;for(const ch of s){const q=alphabet.indexOf(ch);if(q<0)continue;acc=count%4?64*acc+q:q;if(count++%4)raw+=String.fromCharCode(255&acc>>(-2*count&6));}return decodeURIComponent(Array.from(raw,c=>'%'+c.charCodeAt(0).toString(16).padStart(2,'0')).join(''));}
const decodedCache=new Map();
function rc4(s,key){const cacheKey=s+'\0'+key;if(decodedCache.has(cacheKey))return decodedCache.get(cacheKey);s=base64(s);const a=Array.from({length:256},(_,i)=>i);let j=0;for(let i=0;i<256;i++){j=(j+a[i]+key.charCodeAt(i%key.length))%256;[a[i],a[j]]=[a[j],a[i]];}let i=0;j=0;let out='';for(const ch of s){i=(i+1)%256;j=(j+a[i])%256;[a[i],a[j]]=[a[j],a[i]];out+=String.fromCharCode(ch.charCodeAt(0)^a[(a[i]+a[j])%256]);}decodedCache.set(cacheKey,out);return out;}
const tableFn=defs.find(n=>n.id.name==='zan');
const literalTable=tableFn.body.body[0].declarations[0].init;
if(literalTable.type!=='ArrayExpression'||literalTable.elements.some(e=>e.type!=='StringLiteral'))throw Error('nonliteral table');
const table=literalTable.elements.map(n=>n.value);
function value(n,frames=new Map(),depth=0){
 if(!n||depth>60)throw Error('depth/empty');
 const v=x=>value(x,frames,depth+1);
 if(['StringLiteral','NumericLiteral','BooleanLiteral'].includes(n.type))return n.value;
 if(n.type==='NullLiteral')return null;
 if(n.type==='Identifier'){
  if(n.name==='undefined')return undefined;
  const b=findBinding(n.name,n,frames);if('value'in b)return b.value;
  if(b.node.type==='VariableDeclarator')return v(b.node.init);
  return b.node;
 }
 if(n.type==='UnaryExpression'){const x=v(n.argument);if(n.operator==='-')return -x;if(n.operator==='+')return +x;if(n.operator==='!')return !x;if(n.operator==='void')return undefined;}
 if(n.type==='BinaryExpression'){
  const a=v(n.left),b=v(n.right);
  switch(n.operator){case '+':return a+b;case '-':return a-b;case '*':return a*b;case '/':return a/b;case '%':return a%b;case '===':return a===b;case '==':return a==b;case '!==':return a!==b;case '<':return a<b;case '<=':return a<=b;case '>':return a>b;case '>=':return a>=b;}
 }
 if(n.type==='AssignmentExpression'&&n.operator==='='&&n.left.type==='Identifier'){
  const x=v(n.right);let fn=scopeAt(n);while(fn&&fn.type==='BlockStatement')fn=scopeAt(parent.get(fn));let m=frames.get(fn);if(!m)frames.set(fn,m=new Map());m.set(n.left.name,x);return x;
 }
 if(n.type==='SequenceExpression'){let x;for(const e of n.expressions)x=v(e);return x;}
 if(n.type==='CallExpression'&&n.callee.type==='Identifier'){
  const name=n.callee.name,args=n.arguments.map(v);
  if(name==='parseInt')return parseInt(args[0],args[1]);
  if(name==='Qan'){
   const i=args[0]-267;if(!Number.isInteger(i)||i<0||i>=table.length||typeof args[1]!=='string')throw Error('invalid Qan args');return rc4(table[i],args[1]);
  }
  const b=findBinding(name,n,frames).node;
  const fn=b?.type==='VariableDeclarator'?b.init:b;
  if(fn&&['FunctionDeclaration','FunctionExpression','ArrowFunctionExpression'].includes(fn.type)){
   const statements=fn.body?.body;
   if(Array.isArray(statements)&&statements.length===1&&statements[0].type==='ReturnStatement'){
    const nested=new Map(frames),locals=new Map();fn.params.forEach((p,i)=>{if(p.type!=='Identifier')throw Error('nonliteral param');locals.set(p.name,args[i]);});nested.set(fn,locals);return value(statements[0].argument,nested,depth+1);
   }
  }
 }
 throw Error('not whitelisted '+n.type+' '+(n.name||n.operator||''));
}
const tableCall=nodes.find(n=>n.type==='CallExpression'&&n.callee.name==='zan'&&(()=>{let p=parent.get(n);while(p&&!/Function/.test(p.type))p=parent.get(p);return p?.type==='FunctionExpression'&&source.slice(p.start,p.end).includes('parseInt');})());
let rotationFn=parent.get(tableCall);while(rotationFn&&!/Function/.test(rotationFn.type))rotationFn=parent.get(rotationFn);
const checksumTest=nodes.find(n=>n.start>=rotationFn.start&&n.end<=rotationFn.end&&n.type==='IfStatement'&&n.test.type==='BinaryExpression'&&n.test.operator==='==='&&(n.test.left.type==='NumericLiteral'||n.test.right.type==='NumericLiteral')).test;
let rotation=-1;for(let i=0;i<table.length;i++){try{if(value(checksumTest)){rotation=i;break;}}catch{}table.push(table.shift());}
if(rotation<0)throw Error('current checksum not resolved');
console.log(JSON.stringify({decoder:'Qan',subtract:267,rotation,tableEntries:table.length,checksum:299251}));

const results=[],unknown=[];
function ownerFunction(n){let p=parent.get(n);while(p&&!/Function/.test(p.type))p=parent.get(p);return p;}
function componentOf(n){for(let p=parent.get(n);p;p=parent.get(p)){if(p.type==='ObjectExpression'){const name=p.properties.find(x=>x.type==='ObjectProperty'&&(x.key.name||x.key.value)==='__name'&&x.value.type==='StringLiteral');if(name)return name.value.value;}}return null;}
const allCalls=nodes.filter(x=>['CallExpression','OptionalCallExpression'].includes(x.type));
for(const n of nodes){
 if(!['MemberExpression','OptionalMemberExpression'].includes(n.type))continue;
 let property;
 try{property=n.computed?value(n.property):n.property.name;}catch(ex){if(source.slice(n.property.start,n.property.end).includes('Gan(')||source.slice(n.property.start,n.property.end).includes('Tan('))unknown.push({offset:n.start,expression:source.slice(n.property.start,n.property.end),reason:ex.message});continue;}
 if(typeof property!=='string'||!property.startsWith('novels_'))continue;
 let fn=parent.get(n);while(fn&&!/Function/.test(fn.type))fn=parent.get(fn);
 const range=x=>({utf16:[x.start,x.end],utf8:[Buffer.byteLength(source.slice(0,x.start)),Buffer.byteLength(source.slice(0,x.end))]});
 const component=componentOf(n);
 const fnText=fn?generate(fn,{comments:false}).code:generate(n).code;
 const fnNodes=fn?nodes.filter(x=>x.start>=fn.start&&x.end<=fn.end):[];
 const callNodes=fnNodes.filter(x=>['CallExpression','OptionalCallExpression'].includes(x.type)&&x.callee.type==='MemberExpression'&&!x.callee.computed&&x.callee.property.name==='call');
 const callArgs=callNodes.map(x=>({range:range(x),arguments:x.arguments.map(a=>generate(a).code),hasUnboundObfuscationPlaceholder:x.arguments.some(a=>/_0x[0-9a-f]+/i.test(generate(a).code))}));
 const invocation=fn&&parent.get(fn)?.type==='CallExpression'&&parent.get(fn).callee===fn?parent.get(fn):null;
 const namedCallers=fn?.id?allCalls.filter(x=>{if(x.callee.type!=='Identifier'||x.callee.name!==fn.id.name||x.start===n.start)return false;try{return findBinding(x.callee.name,x,new Map()).node===fn;}catch{return false;}}):[];
 const callerNodes=invocation?[invocation]:namedCallers;
 const callers=callerNodes.map(x=>{const owner=ownerFunction(x);return {range:range(x),arguments:x.arguments.map(a=>generate(a).code),component:componentOf(x),callerName:owner?.id?.name||null,callerRange:owner?range(owner):null,callerSource:owner?generate(owner,{comments:false}).code:null};});
 results.push({method:property,propertyRange:range(n.property),memberRange:range(n),originalLine:n.loc.start.line,propertyExpression:source.slice(n.property.start,n.property.end),functionRange:fn?range(fn):null,functionName:fn?.id?.name||null,parameters:fn?.params.map(p=>generate(p).code)||[],component,async:!!fn?.async,callArguments:callArgs,callers,classification:callArgs.some(c=>c.hasUnboundObfuscationPlaceholder)?'injected_or_unresolved_branch_not_proven_executable':!fn?.async?'non_async_scaffolding_not_proven_executable':'async_bridge_candidate',source:fnText});
}
const report={sourcePath,sourceSha256:crypto.createHash('sha256').update(source).digest('hex'),method:'handwritten RC4/custom-base64 + whitelist AST arithmetic/scope folding; no target code execution',decoder:{name:'Qan',subtract:267,table:'zan',rotation,tableEntries:table.length,checksum:299251},results,unknown};
fs.writeFileSync(path.join(root,'current-bridge-call-sites.json'),JSON.stringify(report,null,2));
console.log(JSON.stringify({callSites:results.length,methods:[...new Set(results.map(r=>r.method))],unresolvedProperties:unknown.length}));
if(process.argv.includes('--inspect')){
  for(const table of ['zan','Arn','Lln']){
    for(const n of nodes.filter(n=>n.type==='CallExpression'&&n.arguments.some(a=>a.type==='Identifier'&&a.name===table)&&['FunctionExpression','ArrowFunctionExpression'].includes(n.callee.type))){console.log('ROTATION',table,generate(n).code.slice(0,6000));}
  }
  console.log('QAN',generate(defs.find(n=>n.id.name==='Qan')).code.slice(0,350));
  for(const call of nodes.filter(n=>n.type==='CallExpression'&&n.callee.type==='Identifier'&&['zan','Urn','Lln'].includes(n.callee.name))){let p=parent.get(call);while(p&&!['FunctionDeclaration','FunctionExpression','ArrowFunctionExpression'].includes(p.type))p=parent.get(p);if(p&&source.slice(p.start,p.end).includes('parseInt'))console.log('ROTATING_FN',call.callee.name,generate(p).code.slice(0,6000));}
}
