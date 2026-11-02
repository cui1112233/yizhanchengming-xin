import preact from '@preact/preset-vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [preact()],
  clearScreen: false,
  // fs.allow:测试用 `?raw` 读取 src-tauri 里的注入脚本(accounts.js),默认只允许 frontend 目录
  server: { port: 1477, strictPort: true, fs: { allow: ['.', '../src-tauri/src/doubao/scripts'] } },
  envPrefix: ['VITE_', 'TAURI_'],
});
