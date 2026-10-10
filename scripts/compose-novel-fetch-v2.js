const fs = require('node:fs');
const path = require('node:path');
const { injectNovelFetchV2Script } = require('../lib/novel-fetch-workshop/v2-page');

const [sourcePath, outputPath] = process.argv.slice(2);
if (!sourcePath || !outputPath) {
  throw new Error('用法：node scripts/compose-novel-fetch-v2.js 源文件 目标文件');
}

const source = fs.readFileSync(path.resolve(sourcePath), 'utf8');
const composed = injectNovelFetchV2Script(source);
fs.mkdirSync(path.dirname(path.resolve(outputPath)), { recursive: true });
fs.writeFileSync(path.resolve(outputPath), composed, 'utf8');
