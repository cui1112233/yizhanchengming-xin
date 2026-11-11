const fontFamilies = new Set(['Microsoft YaHei', 'SimHei', 'SimSun', 'KaiTi', 'Arial']);
const alignments = new Set(['left', 'center', 'right']);

function finite(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value));
}

function color(value, fallback) {
  const text = String(value || '').trim();
  return /^#[0-9a-f]{6}([0-9a-f]{2})?$/i.test(text) ? text.toLowerCase() : fallback;
}

export const aicutFontOptions = [
  { value: 'Microsoft YaHei', label: '微软雅黑' },
  { value: 'SimHei', label: '黑体' },
  { value: 'SimSun', label: '宋体' },
  { value: 'KaiTi', label: '楷体' },
  { value: 'Arial', label: 'Arial' }
];

const textResourcePresets = Object.freeze({
  'text-basic-caption': { fontSize: 48, textColor: '#ffffff', strokeColor: '#000000', strokeWidth: 2, textAlign: 'center' },
  'text-bold-title': { fontFamily: 'SimHei', fontSize: 82, textColor: '#ffffff', strokeColor: '#111827', strokeWidth: 5, letterSpacing: 2, textAlign: 'center' },
  'text-outline-caption': { fontSize: 64, textColor: '#ffffff', strokeColor: '#000000', strokeWidth: 6, textAlign: 'center' },
  'text-shadow-caption': { fontSize: 58, textColor: '#ffffff', strokeWidth: 0, shadowColor: '#000000cc', shadowOffsetX: 4, shadowOffsetY: 6, shadowBlur: 8, textAlign: 'center' },
  'text-panel-caption': { fontSize: 52, textColor: '#ffffff', strokeWidth: 0, backgroundColor: '#000000b8', textAlign: 'center' },
  'text-word-highlight': { fontSize: 64, textColor: '#ffe239', strokeColor: '#000000', strokeWidth: 4, textAlign: 'center' },
  'text-news-bar': { fontFamily: 'SimHei', fontSize: 52, textColor: '#ffffff', strokeWidth: 0, backgroundColor: '#155eefdd', textAlign: 'left' },
  'text-lower-third': { fontSize: 48, textColor: '#ffffff', strokeWidth: 0, backgroundColor: '#111827d9', textAlign: 'left' },
  'text-comic': { fontFamily: 'SimHei', fontSize: 64, textColor: '#ffe239', strokeColor: '#7c2d12', strokeWidth: 6, letterSpacing: 2, textAlign: 'center' },
  'text-handwriting': { fontFamily: 'KaiTi', fontSize: 58, textColor: '#fff7ed', strokeColor: '#7c2d12', strokeWidth: 2, shadowColor: '#00000099', shadowOffsetX: 3, shadowOffsetY: 4, shadowBlur: 5, textAlign: 'center' },
  'text-vertical': { fontFamily: 'KaiTi', fontSize: 58, textColor: '#ffffff', strokeColor: '#000000', strokeWidth: 3, textAlign: 'center' },
  'text-gradient-title': { fontFamily: 'SimHei', fontSize: 82, textColor: '#b78bff', strokeColor: '#25164f', strokeWidth: 5, letterSpacing: 2, textAlign: 'center' }
});

export function aicutTextResourcePreset(id) {
  const preset = textResourcePresets[String(id || '')];
  return preset ? normalizeAICutTextStyle(preset) : null;
}

function cleanFontAsset(asset) {
  const id = String(asset?.id || '').trim();
  const family = String(asset?.fontFamily || '').trim();
  const url = String(asset?.url || '').trim();
  if (asset?.kind !== 'font' || !id || !family || family.length > 128 || !url) return null;
  return { id, family, url, contentType: String(asset?.contentType || '').toLowerCase() };
}

export function aicutFontOptionsForAssets(assets = []) {
  const custom = assets.map(cleanFontAsset).filter(Boolean).map((font) => ({ value: `asset:${font.id}`, label: `${font.family} · 用户字体` }));
  return [...aicutFontOptions, ...custom];
}

export function aicutFontFaceCSS(assets = []) {
  return assets.map(cleanFontAsset).filter(Boolean).map((font) => {
    const format = font.contentType === 'font/otf' ? 'opentype' : 'truetype';
    return `@font-face{font-family:${JSON.stringify(font.family)};src:url(${JSON.stringify(font.url)}) format('${format}');font-display:swap;}`;
  }).join('\n');
}

export function normalizeAICutTextStyle(value = {}) {
	const fontAssetId = String(value.fontAssetId || '').trim().slice(0, 96);
	const requestedFamily = String(value.fontFamily || '').trim().slice(0, 128);
	const style = {
	fontFamily: fontAssetId && requestedFamily ? requestedFamily : fontFamilies.has(value.fontFamily) ? value.fontFamily : 'Microsoft YaHei',
    fontSize: clamp(finite(value.fontSize, 54), 12, 240),
	textCurve: clamp(finite(value.textCurve, 0), -180, 180),
    textColor: color(value.textColor, '#ffffff'),
    strokeColor: color(value.strokeColor, '#000000'),
    strokeWidth: clamp(finite(value.strokeWidth, 2), 0, 12),
    backgroundColor: color(value.backgroundColor, '#00000000'),
    shadowColor: color(value.shadowColor, '#000000cc'),
    shadowOffsetX: clamp(finite(value.shadowOffsetX, 0), -40, 40),
    shadowOffsetY: clamp(finite(value.shadowOffsetY, 0), -40, 40),
    shadowBlur: clamp(finite(value.shadowBlur, 0), 0, 20),
    letterSpacing: clamp(finite(value.letterSpacing, 0), -10, 40),
    lineHeight: clamp(finite(value.lineHeight, 1.3), .8, 3),
    textAlign: alignments.has(value.textAlign) ? value.textAlign : 'center'
	};
	if (fontAssetId) style.fontAssetId = fontAssetId;
	const textRuns = normalizeAICutTextRuns(value.textRuns);
	if (textRuns.length) style.textRuns = textRuns;
	return style;
}

export function normalizeAICutTextRuns(value = []) {
	if (!Array.isArray(value)) return [];
	return value.slice(0, 32).map((run) => ({
		start: Math.trunc(clamp(finite(run?.start, 0), 0, 511)),
		end: Math.trunc(clamp(finite(run?.end, 0), 0, 512)),
		textColor: color(run?.textColor, '#ffffff'),
		fontSize: clamp(finite(run?.fontSize, 54), 12, 240),
		bold: Boolean(run?.bold),
		italic: Boolean(run?.italic)
	})).filter((run) => run.end > run.start).sort((left, right) => left.start - right.start || left.end - right.end);
}

export function aicutTextRunStyleAt(value, index) {
	const runs = normalizeAICutTextRuns(value?.textRuns);
	const run = runs.filter((item) => index >= item.start && index < item.end).at(-1);
	if (!run) return undefined;
	return {
		color: run.textColor,
		fontSize: `${run.fontSize}px`,
		fontWeight: run.bold ? 700 : 400,
		fontStyle: run.italic ? 'italic' : 'normal'
	};
}

export function aicutCurvedTextLayout(text, curve, fontSize = 54, letterSpacing = 0) {
	const characters = Array.from(String(text || ''));
	const bend = clamp(finite(curve, 0), -180, 180);
	if (!characters.length) return [];
	const height = Math.abs(bend) / 180 * clamp(finite(fontSize, 54), 12, 240) * 1.8;
	return characters.map((character, index) => {
		const position = characters.length === 1 ? 0 : index / (characters.length - 1) * 2 - 1;
		return {
			character,
			translateY: bend === 0 ? 0 : -Math.sign(bend) * (1 - position * position) * height,
			rotate: bend === 0 ? 0 : bend * position * .5,
			letterSpacing: finite(letterSpacing, 0)
		};
	});
}

export function textAnimationFrame(text, animation, elapsedMs, durationMs) {
  const characters = Array.from(String(text || ''));
  const progress = clamp(finite(elapsedMs, 0) / Math.max(1, finite(durationMs, 1)), 0, 1);
  if (animation === 'typewriter') {
    const count = Math.min(characters.length, Math.floor(characters.length * progress));
    return { visibleText: characters.slice(0, count).join(''), highlightedCount: 0 };
  }
  if (animation === 'word-highlight') {
    const count = Math.min(characters.length, Math.floor(characters.length * progress));
    return { visibleText: characters.join(''), highlightedCount: count };
  }
  return { visibleText: characters.join(''), highlightedCount: 0 };
}

export function aicutTextPreviewStyle(value = {}) {
  const style = normalizeAICutTextStyle(value);
  const hasBackground = style.backgroundColor.length === 7 || style.backgroundColor.slice(7) !== '00';
  const hasShadow = style.shadowOffsetX !== 0 || style.shadowOffsetY !== 0 || style.shadowBlur !== 0;
  return {
    fontFamily: style.fontFamily,
    fontSize: `${style.fontSize}px`,
    color: style.textColor,
    WebkitTextStroke: style.strokeWidth ? `${style.strokeWidth}px ${style.strokeColor}` : undefined,
    backgroundColor: style.backgroundColor,
    textShadow: hasShadow ? `${style.shadowOffsetX}px ${style.shadowOffsetY}px ${style.shadowBlur}px ${style.shadowColor}` : undefined,
    padding: hasBackground ? '0.18em 0.35em' : undefined,
    borderRadius: hasBackground ? '0.12em' : undefined,
    letterSpacing: `${style.letterSpacing}px`,
    lineHeight: style.lineHeight,
    textAlign: style.textAlign
  };
}
