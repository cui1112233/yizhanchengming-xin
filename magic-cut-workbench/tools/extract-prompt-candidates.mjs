import { createHash } from 'node:crypto'
import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'

const candidatePattern = /(提示词|不生成字幕|解说旁白|生成约束|system\s*prompt|prompt\s*template)/i
const quotedStringPattern = /(["'`])((?:\\.|(?!\1)[\s\S]){8,}?)\1/g

export function extractCandidates(source, file) {
  const candidates = []
  for (const match of source.matchAll(quotedStringPattern)) {
    const raw = match[2]
    const text = decodeLiteral(raw, match[1])
    if (!candidatePattern.test(text)) continue
    const contentOffset = match.index + 1
    const startByte = Buffer.byteLength(source.slice(0, contentOffset))
    const endByte = startByte + Buffer.byteLength(raw)
    candidates.push({
      file,
      text,
      startByte,
      endByte,
      sha256: createHash('sha256').update(text).digest('hex'),
    })
  }
  return candidates
}

function decodeLiteral(raw, quote) {
  if (quote === '"') {
    try { return JSON.parse(`"${raw}"`) } catch (_) { return raw }
  }
  return raw
    .replaceAll(`\\${quote}`, quote)
    .replaceAll('\\n', '\n')
    .replaceAll('\\r', '\r')
    .replaceAll('\\t', '\t')
}

async function main(paths) {
  const output = []
  for (const path of paths) {
    const source = await readFile(path, 'utf8')
    output.push(...extractCandidates(source, path))
  }
  process.stdout.write(`${JSON.stringify(output, null, 2)}\n`)
}

if (process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]) {
  await main(process.argv.slice(2))
}
