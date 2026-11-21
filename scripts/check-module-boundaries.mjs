import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const SOURCE_EXTENSIONS = new Set(['.go', '.js', '.jsx', '.mjs', '.ts', '.tsx']);

function walk(directory) {
  if (!fs.existsSync(directory)) return [];
  const files = [];
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const full = path.join(directory, entry.name);
    if (entry.isDirectory()) files.push(...walk(full));
    else if (SOURCE_EXTENSIONS.has(path.extname(entry.name))) files.push(full);
  }
  return files;
}

function importsFor(file, source) {
  const imports = [];
  const pattern = path.extname(file) === '.go'
    ? /(?:^|\s)"([^"]+)"/gm
    : /(?:from\s+|import\s*)['"]([^'"]+)['"]/gm;
  for (const match of source.matchAll(pattern)) imports.push(match[1]);
  return imports;
}

function moduleOwner(file, root, moduleRoots) {
  for (const moduleRoot of moduleRoots) {
    const absoluteRoot = path.resolve(root, moduleRoot);
    const relative = path.relative(absoluteRoot, file);
    if (relative === '' || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative)) continue;
    return relative.split(path.sep)[0] || null;
  }
  return null;
}

function resolveImport(file, importPath, root, goModulePrefix) {
  if (importPath.startsWith('.')) return path.resolve(path.dirname(file), importPath);
  if (goModulePrefix && importPath.startsWith(goModulePrefix)) {
    return path.resolve(root, importPath.slice(goModulePrefix.length));
  }
  return null;
}

export function checkModuleBoundaries({ root, config }) {
  const absoluteRoot = path.resolve(root);
  const moduleRoots = config.moduleRoots || [];
  const violations = [];
  for (const declaredRoot of moduleRoots) {
    for (const file of walk(path.resolve(absoluteRoot, declaredRoot))) {
      const owner = moduleOwner(file, absoluteRoot, moduleRoots);
      if (!owner) continue;
      const source = fs.readFileSync(file, 'utf8');
      for (const importPath of importsFor(file, source)) {
        const target = resolveImport(file, importPath, absoluteRoot, config.goModulePrefix || '');
        if (!target) continue;
        const targetOwner = moduleOwner(target, absoluteRoot, moduleRoots);
        if (targetOwner && targetOwner !== owner) {
          violations.push({
            file: path.relative(absoluteRoot, file).replaceAll('\\', '/'),
            owner,
            importPath,
            targetOwner,
          });
        }
      }
    }
  }
  return { violations };
}

const invoked = process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url);
if (invoked) {
  const configPath = process.argv[2];
  if (!configPath) throw new Error('用法：node scripts/check-module-boundaries.mjs <config.json> [repo-root]');
  const root = path.resolve(process.argv[3] || process.cwd());
  const config = JSON.parse(fs.readFileSync(path.resolve(configPath), 'utf8'));
  const result = checkModuleBoundaries({ root, config });
  if (result.violations.length) {
    for (const violation of result.violations) {
      console.error(`${violation.file}: 模块 ${violation.owner} 不得导入模块 ${violation.targetOwner}（${violation.importPath}）`);
    }
    process.exitCode = 1;
  } else {
    console.log('模块边界检查通过');
  }
}
