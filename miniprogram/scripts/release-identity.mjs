import { execFileSync, spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { readFileSync, renameSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const RELEASE_VERSION = '1.2.0';
const RELEASE_MATERIAL = 'liquid-smooth-a';

function git(cwd, arguments_) {
  return execFileSync('git', arguments_, {
    cwd,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  }).trim();
}

export function prepareReleaseIdentity({ cwd, version, material }) {
  try {
    const commit = git(cwd, ['rev-parse', '--verify', 'HEAD']);
    const status = git(cwd, ['status', '--porcelain', '--untracked-files=normal']);
    return {
      version,
      material,
      commit,
      shortCommit: commit.slice(0, 7),
      dirty: status.length > 0,
    };
  } catch {
    throw new Error('A Git repository with at least one commit is required to build a release identity.');
  }
}

function sha256(path) {
  return `sha256:${createHash('sha256').update(readFileSync(path)).digest('hex')}`;
}

export function writeReleaseManifest({ cwd, identity, builtAt = new Date().toISOString() }) {
  const manifest = {
    ...identity,
    builtAt,
    wallpapers: {
      light: sha256(join(cwd, 'src/assets/glass/wallpaper-light.png')),
      dark: sha256(join(cwd, 'src/assets/glass/wallpaper-dark.png')),
    },
  };
  const manifestPath = join(cwd, 'dist/release-manifest.json');
  const temporaryPath = `${manifestPath}.tmp`;
  writeFileSync(temporaryPath, `${JSON.stringify(manifest, null, 2)}\n`);
  renameSync(temporaryPath, manifestPath);
  return manifest;
}

function run(command, arguments_, options) {
  const result = spawnSync(command, arguments_, { stdio: 'inherit', ...options });
  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}

export function buildRelease({ cwd, mode = 'production' }) {
  const identity = prepareReleaseIdentity({
    cwd,
    version: RELEASE_VERSION,
    material: RELEASE_MATERIAL,
  });
  const environment = {
    ...process.env,
    MINIPROGRAM_BUILD_SHA: identity.commit,
    MINIPROGRAM_BUILD_DIRTY: String(identity.dirty),
  };
  run(process.execPath, [join(cwd, 'node_modules/@tarojs/cli/bin/taro'), 'build', '--type', 'weapp', '--mode', mode], {
    cwd,
    env: environment,
  });
  run(process.execPath, [join(cwd, 'scripts/sanitize-weapp-wxss.mjs')], { cwd, env: environment });
  return writeReleaseManifest({ cwd, identity });
}

const invokedPath = process.argv[1] ? fileURLToPath(import.meta.url) === process.argv[1] : false;
if (invokedPath && process.argv[2] === 'build') {
  buildRelease({ cwd: process.cwd(), mode: process.argv[3] || 'production' });
}
