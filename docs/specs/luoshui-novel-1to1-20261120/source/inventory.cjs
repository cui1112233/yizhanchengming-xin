// Parse recovered shipped text only; never import or execute the target.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const req = require('node:module').createRequire(path.resolve('frontend/package.json'));
const parser = req('@babel/parser');
const generate = req('@babel/generator').default;
const root = __dirname;
const dir = path.join(root, 'frontend');
const manifest = JSON.parse(fs.readFileSync(path.join(dir, 'manifest.json'), 'utf8'));
const controls = [], strings = [], functions = [], emits = [], props = [];
for (const component of manifest.components) {
  const file = component.file, source = fs.readFileSync(path.join(dir, file), 'utf8');
  const ast = parser.parse(source, { sourceType: 'unambiguous' });
  const walk = node => {
    if (!node || typeof node !== 'object') return;
    if (node.type === 'ObjectProperty') {
      const key = node.key.name || node.key.value;
      if (key === 'props' && node.value.type === 'ObjectExpression') for(const p of node.value.properties) if(p.type==='ObjectProperty') props.push({component:component.name,file,line:p.loc.start.line,name:p.key.name||p.key.value,definition:generate(p.value).code});
      if (typeof key === 'string' && (/^on[A-Z]|^onUpdate:/.test(key) || ['disabled','placeholder','title','accept','multiple','min','max'].includes(key))) {
        controls.push({ id: 'UI-' + String(controls.length + 1).padStart(4,'0'), component: component.name, file, line: node.loc.start.line, kind: key, expression: generate(node.value).code, status: 'static_expression_only', runtime: 'not_tested' });
      }
      if (key === 'emits' && node.value.type === 'ArrayExpression') for (const e of node.value.elements) if (e?.type === 'StringLiteral') emits.push({ component: component.name, file, line: e.loc.start.line, event: e.value });
    }
    if (node.type === 'FunctionDeclaration') functions.push({component:component.name,file,line:node.loc.start.line,endLine:node.loc.end.line,name:node.id?.name});
    if (node.type === 'StringLiteral' && /[\u3400-\u9fff]/.test(node.value)) strings.push({component:component.name,file,line:node.loc.start.line,text:node.value});
    for (const [key,value] of Object.entries(node)) {
      if (['loc','extra','leadingComments','trailingComments','innerComments'].includes(key)) continue;
      if (Array.isArray(value)) value.forEach(v=>{if(v?.type)walk(v)});
      else if(value?.type)walk(value);
    }
  };
  walk(ast);
}
const output={sourceSha256:manifest.sha256,method:'Babel AST static inventory; event expressions are NOT completed feature contracts',components:manifest.components,controls,emits,functions,strings,props};
fs.writeFileSync(path.join(root,'ui-inventory.json'),JSON.stringify(output,null,2));
const original='F:/软件/洛水2026/洛水2026/modules/novels.pyd';
const identity={frontend: {path:manifest.sourcePath,sha256:manifest.sha256},native:{path:original,sha256:crypto.createHash('sha256').update(fs.readFileSync(original)).digest('hex')},limitations:['Recovered Vue render fragments, not original author source','Static evidence does not establish runtime or pixel equivalence','Old frontend 5278a6d1 is historical only']};
fs.writeFileSync(path.join(root,'identity.json'),JSON.stringify(identity,null,2));
console.log(JSON.stringify({components:manifest.components.length,controls:controls.length,emits:emits.length,functions:functions.length,chineseStrings:strings.length}));
