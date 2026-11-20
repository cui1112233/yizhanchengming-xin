// Static analysis only: parse shipped JavaScript as text; never import or execute it.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { createRequire } = require('node:module');
const localRequire = createRequire(path.resolve('frontend/package.json'));
const parser = localRequire('@babel/parser');
const generate = localRequire('@babel/generator').default;
const root = __dirname;
const sourcePath = 'F:/软件/洛水2026/洛水2026/dist/assets/index-DgTjggsH.js';
const output = path.resolve('docs/specs/luoshui-novel-1to1-20261120/source/frontend');
fs.mkdirSync(output, { recursive: true });
const bytes = fs.readFileSync(sourcePath);
const source = bytes.toString('utf8');
const sha256 = crypto.createHash('sha256').update(bytes).digest('hex');
console.log('Parsing current shipped frontend without executing it');
const ast = parser.parse(source, { sourceType: 'unambiguous', attachComment: false, tokens: false });
const manifest = { sourcePath, sha256, method: 'Babel AST selection and formatting', runtimeEquivalence: 'unverified', limitations: ['Generated files are inspection-only fragments with unresolved external and obfuscated bindings.', 'Not original pre-build source; function names and decoder expressions are preserved.', 'Generated line numbers refer to recovered fragments, not original source.'], components: [] };
const stack = [ast];
while (stack.length) {
  const node = stack.pop();
  if (!node || typeof node !== 'object') continue;
  if (node.type === 'ObjectExpression') {
    const prop = node.properties.find(p => p.type === 'ObjectProperty' && (p.key.name || p.key.value) === '__name' && p.value.type === 'StringLiteral');
    const name = prop?.value.value;
    if (name && /^Novel/.test(name)) {
      const relative = name.replace(/[^A-Za-z0-9_-]/g, '_') + '.readable.js';
      const header = '// Inspection only. Original SHA256: ' + sha256 + '\n// Original UTF16 offsets [' + node.start + ', ' + node.end + '); original line ' + node.loc.start.line + '\n// External dependencies and runtime equivalence remain unresolved.\n';
      const text = header + 'const recoveredComponent = ' + generate(node, { compact: false, comments: false }).code + ';\n';
      fs.writeFileSync(path.join(output, relative), text, 'utf8');
      manifest.components.push({ name, file: relative, originalUtf16Range: [node.start, node.end], originalUtf8Range: [Buffer.byteLength(source.slice(0,node.start)), Buffer.byteLength(source.slice(0,node.end))], originalLine: node.loc.start.line, recoveredSha256: crypto.createHash('sha256').update(text).digest('hex') });
    }
  }
  for (const [key, value] of Object.entries(node)) {
    if (key === 'loc' || key === 'extra' || key.endsWith('Comments')) continue;
    if (Array.isArray(value)) { for (const item of value) if (item?.type) stack.push(item); }
    else if (value?.type) stack.push(value);
  }
}
manifest.components.sort((a,b) => a.originalUtf16Range[0] - b.originalUtf16Range[0]);
fs.writeFileSync(path.join(output,'manifest.json'), JSON.stringify(manifest,null,2));
console.log(JSON.stringify({ components: manifest.components.length, names: manifest.components.map(c=>c.name), output }));
