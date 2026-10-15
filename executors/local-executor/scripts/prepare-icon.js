const fs = require('node:fs');
const path = require('node:path');

const sourcePath = path.join(__dirname, '..', '..', 'branding', 'yizhan-icon.png');
const png = fs.readFileSync(sourcePath);
const pngSignature = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
if (png.length < 24 || !png.subarray(0, 8).equals(pngSignature)) {
  throw new Error(`Invalid PNG icon: ${sourcePath}`);
}

const width = png.readUInt32BE(16);
const height = png.readUInt32BE(20);
if (width !== 256 || height !== 256) {
  throw new Error(`Windows icon source must be 256x256, got ${width}x${height}`);
}

const buildDir = path.join(__dirname, '..', 'build');
fs.mkdirSync(buildDir, { recursive: true });
fs.writeFileSync(path.join(buildDir, 'icon.png'), png);

const header = Buffer.alloc(22);
header.writeUInt16LE(0, 0);
header.writeUInt16LE(1, 2);
header.writeUInt16LE(1, 4);
header[6] = 0;
header[7] = 0;
header[8] = 0;
header[9] = 0;
header.writeUInt16LE(1, 10);
header.writeUInt16LE(32, 12);
header.writeUInt32LE(png.length, 14);
header.writeUInt32LE(22, 18);

fs.writeFileSync(path.join(buildDir, 'icon.ico'), Buffer.concat([header, png]));
