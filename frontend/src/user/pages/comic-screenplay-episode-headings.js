// Only standalone episode headings count; chapter/volume headings remain body text.
const headingPattern = /^[ \t]*(?:#{1,6}[ \t]*)?第[ \t]*([0-9〇零一二三四五六七八九十百千万两]+)[ \t]*集([ \t：:·—\-][^\r\n]*)?\r?$/gm;
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
export function episodeHeadingMatches(content) {
  return [...String(content||'').matchAll(headingPattern)].map(match=>({match,no:episodeNumber(match[1])})).filter(row=>Number.isSafeInteger(row.no)&&row.no>0);
}
export function splitDocumentEpisodes(content) {
  const text=String(content||''), headings=episodeHeadingMatches(text);
  if (!text.trim()) return [];
  if (!headings.length) return [{no:1,content:text.trim()}];
  return headings.map(({match,no},index)=>({no,content:text.slice(index===0?0:match.index,headings[index+1]?.match.index??text.length).trim()}));
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
