// Only standalone episode headings count; chapter/volume headings remain body text.
const episodePattern = /^[ \t]*(?:#{1,6}[ \t]*)?第[ \t]*([0-9〇零一二三四五六七八九十百千万两]+)[ \t]*集([ \t：:·—\-][^\r\n]*)?\r?$/gm;
const englishEpisodePattern = /^[ \t]*(?:#{1,6}[ \t]*)?Episode[ \t]+([0-9]+)([ \t：:·—\-][^\r\n]*)?\r?$/gim;
// 洛水章节识别：章/回/节/篇/卷/部/集，以及 Chapter/Episode。
const chapterPattern = /^[ \t]*(?:#{1,6}[ \t]*)?第[ \t]*([0-9〇零一二三四五六七八九十百千万两]+)[ \t]*(章|回|节|篇|卷|部|集)([ \t：:·—\-][^\r\n]*)?\r?$/gm;
const englishChapterPattern = /^[ \t]*(?:#{1,6}[ \t]*)?(Chapter|Episode)[ \t]+([0-9]+)([ \t：:·—\-][^\r\n]*)?\r?$/gim;

export function episodeNumber(value) {
  if (/^[0-9]+$/.test(value)) return Number(value);
  const digits = { '〇':0, '零':0, '一':1, '二':2, '两':2, '三':3, '四':4, '五':5, '六':6, '七':7, '八':8, '九':9 };
  if (/^[〇零一二三四五六七八九两]+$/.test(value)) return Number([...value].map(char=>digits[char]).join(''));
  let total=0, section=0, digit=0;
  for (const char of value) {
    if (Object.hasOwn(digits,char)) { digit=digits[char]; continue; }
    const unit=({'十':10,'百':100,'千':1000,'万':10000})[char];
    if (!unit) return NaN;
    if (unit===10000) { total+=(section+digit)*unit; section=0; digit=0; }
    else { section+=(digit||1)*unit; digit=0; }
  }
  return total+section+digit;
}

function mergedMatches(text, patterns) {
  const rows = patterns.flatMap(pattern => [...String(text||'').matchAll(pattern)]);
  return rows.sort((a,b)=>a.index-b.index);
}
function cleanHeading(value) {
  return String(value||'').replace(/^[ \t]*#{1,6}[ \t]*/, '').trim();
}

export function episodeHeadingMatches(content) {
  return mergedMatches(content, [episodePattern, englishEpisodePattern])
    .map(match=>({match,no:episodeNumber(match[1])}))
    .filter(row=>Number.isSafeInteger(row.no)&&row.no>0);
}

// 章节标题用于「项目原文 → 章节列表」的拆分，允许章/回/节/篇/卷/部/集与 Chapter/Episode。
export function chapterHeadingMatches(content) {
  const chinese = [...String(content||'').matchAll(chapterPattern)]
    .map(match=>({match,no:episodeNumber(match[1]),unit:match[2]}));
  const english = [...String(content||'').matchAll(englishChapterPattern)]
    .map(match=>({match,no:Number(match[2]),unit:match[1]}));
  return [...chinese, ...english]
    .filter(row=>Number.isSafeInteger(row.no)&&row.no>0)
    .sort((a,b)=>a.match.index-b.match.index)
    .map(row=>({...row,title:cleanHeading(row.match[0])||`第${row.no}${row.unit==='Chapter'||row.unit==='Episode'?row.unit:'章'}`}));
}

export function splitDocumentEpisodes(content) {
  const text=String(content||''), headings=episodeHeadingMatches(text);
  if (!text.trim()) return [];
  if (!headings.length) return [{no:1,content:text.trim()}];
  return headings.map(({match,no},index)=>({no,content:text.slice(index===0?0:match.index,headings[index+1]?.match.index??text.length).trim()}));
}

// 返回按标题拆分的章节；无标题时整篇一章，前言单独成章，空章保留位置。
export function splitSourceChapters(content) {
  const text=String(content||'');
  if (!text.trim()) return [];
  const headings=chapterHeadingMatches(text);
  if (!headings.length) return [{title:'',sourceText:text.trim()}];
  const rows=[];
  const preamble=text.slice(0,headings[0].match.index).trim();
  if (preamble) rows.push({title:'前言',sourceText:preamble});
  headings.forEach((heading,index)=>{
    const start=heading.match.index+heading.match[0].length;
    const end=index+1<headings.length?headings[index+1].match.index:text.length;
    rows.push({title:heading.title,sourceText:text.slice(start,end).trim()});
  });
  return rows;
}

export function assertUniqueEpisodeNumbers(episodes) {
  const seen=new Set();
  for (const episode of episodes) {
    if (seen.has(episode.no)) throw Error(`第${episode.no}集编号重复，请修改后再生成`);
    seen.add(episode.no);
  }
}

export function stripEpisodeHeading(content) {
  const text=String(content||'').trim(), first=episodeHeadingMatches(text)[0];
  return first?.match.index===0?text.slice(first.match[0].length).trim():text;
}
