import { existsSync } from 'node:fs';
import { readdir, rm } from 'node:fs/promises';
import { resolve } from 'node:path';

const dist = resolve(process.cwd(), 'dist');

if (existsSync(dist)) {
  // 保留用户下载的执行器包；其余内容都是 Vite 上一次构建的派生产物。
  // 先由脚本清理、再让 Vite 使用 emptyOutDir=false，避免 Windows 在大量旧文件上清空目录时触发原生崩溃。
  for (const entry of await readdir(dist)) {
    if (entry === 'downloads') continue;
    await rm(resolve(dist, entry), { recursive: true, force: true });
  }
}
