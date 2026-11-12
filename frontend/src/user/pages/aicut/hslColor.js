export const HSL_COLOR_BANDS = Object.freeze([
  { key: 'red', label: '红色', hue: 0, color: '#ef4444' },
  { key: 'orange', label: '橙色', hue: 30, color: '#f58b35' },
  { key: 'yellow', label: '黄色', hue: 60, color: '#e9cb37' },
  { key: 'green', label: '绿色', hue: 120, color: '#45ba70' },
  { key: 'cyan', label: '青色', hue: 180, color: '#35bfc5' },
  { key: 'blue', label: '蓝色', hue: 240, color: '#427ee9' },
  { key: 'purple', label: '紫色', hue: 270, color: '#9758de' },
  { key: 'magenta', label: '洋红', hue: 300, color: '#df58b4' },
].map(Object.freeze));

const clamp = (value, min, max) => Math.min(max, Math.max(min, value));
const adjustment = value => typeof value === 'number' && Number.isFinite(value) ? clamp(value, -100, 100) : 0;

/** Stored contract: eight named bands, each with hue/saturation/luminance in [-100,100]. */
export function normalizeHSLAdjustments(settings) {
  return Object.fromEntries(HSL_COLOR_BANDS.map(({ key }) => [key, {
    hue: adjustment(settings?.[key]?.hue),
    saturation: adjustment(settings?.[key]?.saturation),
    luminance: adjustment(settings?.[key]?.luminance),
  }]));
}

function compile(settings) {
  const normalized = normalizeHSLAdjustments(settings);
  const bands = HSL_COLOR_BANDS.map(({ key }) => normalized[key]);
  return bands.some(band => band.hue || band.saturation || band.luminance) ? bands : null;
}

export function hasHSLAdjustments(settings) { return compile(settings) !== null; }

// LUT builders normalize once, then reuse this allocation-free RGB kernel.
export function createHSLRGBTransformer(settings) {
  const bands = compile(settings);
  return (red, green, blue, output, offset = 0) => {
    if (bands) transformInto(red, green, blue, bands, output, offset);
  };
}

/** 33^3 RGB lattice, blue slices along X and green rows along Y; red varies fastest. */
export function createHSLLUTAtlas(settings) {
  const size = 33, width = size * size, height = size;
  const pixels = new Uint8Array(width * height * 4);
  const bands = compile(settings);
  for (let b = 0; b < size; b++) for (let g = 0; g < size; g++) for (let r = 0; r < size; r++) {
    const offset = (g * width + b * size + r) * 4;
    const red = r * 255 / 32, green = g * 255 / 32, blue = b * 255 / 32;
    pixels[offset] = Math.round(red);
    pixels[offset + 1] = Math.round(green);
    pixels[offset + 2] = Math.round(blue);
    pixels[offset + 3] = 255;
    if (bands) transformInto(red, green, blue, bands, pixels, offset);
  }
  return { size, width, height, pixels };
}

// A signed amount moves a normalized channel toward its upper/lower endpoint.
const shift = (channel, amount) => channel + (amount >= 0 ? (1 - channel) * amount : channel * amount);

function transformInto(red, green, blue, bands, target, offset) {
  const r = red / 255, g = green / 255, b = blue / 255;
  const max = Math.max(r, g, b), min = Math.min(r, g, b), delta = max - min;
  // An achromatic pixel has no hue membership: preserve black, white and gray.
  if (!delta) return;
  const lightness = (max + min) / 2;
  const saturation = delta / (1 - Math.abs(2 * lightness - 1));
  let hue = max === r ? 60 * (((g - b) / delta) % 6) : max === g ? 60 * ((b - r) / delta + 2) : 60 * ((r - g) / delta + 4);
  hue = (hue + 360) % 360;

  let left = 7;
  for (let i = 0; i < 7; i++) {
    if (hue < HSL_COLOR_BANDS[i + 1].hue) { left = i; break; }
  }
  const right = (left + 1) % 8;
  const start = HSL_COLOR_BANDS[left].hue;
  const end = right === 0 ? 360 : HSL_COLOR_BANDS[right].hue;
  // Raised cosine crossfade gives continuous values and slopes at every center,
  // including magenta -> red. All three channels use the ORIGINAL hue weights.
  const weight = (1 - Math.cos(Math.PI * (hue - start) / (end - start))) / 2;
  const a = bands[left], z = bands[right];
  const h = a.hue * (1 - weight) + z.hue * weight;
  const s = a.saturation * (1 - weight) + z.saturation * weight;
  const l = a.luminance * (1 - weight) + z.luminance * weight;
  if (!h && !s && !l) return;
  // Hue slider +/-100 = +/-30 degrees; S and L use the signed endpoint rule.
  const nextHue = (hue + h * .3 + 360) % 360;
  const nextSaturation = shift(saturation, s / 100);
  const nextLightness = shift(lightness, l / 100);
  const chroma = (1 - Math.abs(2 * nextLightness - 1)) * nextSaturation;
  const x = chroma * (1 - Math.abs((nextHue / 60) % 2 - 1));
  const m = nextLightness - chroma / 2;
  const sector = Math.floor(nextHue / 60);
  const rr = sector === 0 || sector === 5 ? chroma : sector === 1 || sector === 4 ? x : 0;
  const gg = sector === 1 || sector === 2 ? chroma : sector === 0 || sector === 3 ? x : 0;
  const bb = sector === 3 || sector === 4 ? chroma : sector === 2 || sector === 5 ? x : 0;
  target[offset] = clamp(Math.round((rr + m) * 255), 0, 255);
  target[offset + 1] = clamp(Math.round((gg + m) * 255), 0, 255);
  target[offset + 2] = clamp(Math.round((bb + m) * 255), 0, 255);
}

/** RGB or unpremultiplied RGBA bytes -> new ordinary array; input and alpha stay intact. */
export function applyHSLToRGB(pixel, settings) {
  if ((!Array.isArray(pixel) && !ArrayBuffer.isView(pixel)) || (pixel.length !== 3 && pixel.length !== 4)
    || !Array.from(pixel).every(value => typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 255)) {
    throw new TypeError('Expected three RGB or four RGBA byte channels');
  }
  const output = Array.from(pixel);
  const bands = compile(settings);
  if (bands && (pixel.length === 3 || pixel[3] !== 0)) transformInto(pixel[0], pixel[1], pixel[2], bands, output, 0);
  return output;
}

/** Pure RGBA buffer transform. Normalize once and allocate no per-pixel arrays. */
export function applyHSLToPixels(pixels, settings) {
  if (!(pixels instanceof Uint8ClampedArray) || pixels.length % 4 !== 0) throw new TypeError('Expected a complete Uint8ClampedArray RGBA buffer');
  const output = new Uint8ClampedArray(pixels);
  const bands = compile(settings);
  if (!bands) return output;
  for (let i = 0; i < pixels.length; i += 4) {
    if (pixels[i + 3]) transformInto(pixels[i], pixels[i + 1], pixels[i + 2], bands, output, i);
  }
  return output;
}
