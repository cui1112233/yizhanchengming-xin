const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const req=require('node:module').createRequire(path.resolve('frontend/package.json'));
const postcss=req('postcss');
const source='F:/软件/洛水2026/洛水2026/dist/assets/index-CVcgNnrd.css';
const bytes=fs.readFileSync(source), ast=postcss.parse(bytes.toString('utf8'));
const selected=[];
ast.walkRules(rule=>{
  if(!/novel|imitation|scheduler|hooks-panel|agent-settings|style-panel/.test(rule.selector))return;
  const parents=[];let p=rule.parent;
  while(p&&p.type!=='root'){if(p.type==='atrule')parents.unshift('@'+p.name+' '+p.params);p=p.parent;}
  let css=rule.toString();for(const at of parents.reverse())css=at+'{'+css+'}';
  selected.push({selector:rule.selector,sourceLine:rule.source.start.line,css});
});
fs.writeFileSync(path.join(__dirname,'novel-styles.css'),selected.map(r=>r.css).join('\n'));
fs.writeFileSync(path.join(__dirname,'style-manifest.json'),JSON.stringify({source,sha256:crypto.createHash('sha256').update(bytes).digest('hex'),selection:'selector regex novel|imitation|scheduler|hooks-panel|agent-settings|style-panel',rules:selected.map(({css,...r})=>r),limitations:['shared generic styles and unresolved external class mappings are not fully included','pixel equivalence not verified']},null,2));
console.log(JSON.stringify({selectedRules:selected.length}));
