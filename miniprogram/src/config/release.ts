declare const __MINIPROGRAM_BUILD_SHA__: string;
declare const __MINIPROGRAM_BUILD_DIRTY__: boolean;

export const MINIPROGRAM_VERSION = "1.2.0";
export const MINIPROGRAM_MATERIAL = "liquid-smooth-a";
const compiledBuildSha = typeof __MINIPROGRAM_BUILD_SHA__ === "string" ? __MINIPROGRAM_BUILD_SHA__ : "unknown";
export const MINIPROGRAM_BUILD_SHORT = /^[0-9a-f]{7,40}$/i.test(compiledBuildSha)
  ? compiledBuildSha.slice(0, 7).toLowerCase()
  : "unknown";
export const MINIPROGRAM_BUILD_DIRTY = typeof __MINIPROGRAM_BUILD_DIRTY__ === "boolean"
  ? __MINIPROGRAM_BUILD_DIRTY__
  : true;
export const MINIPROGRAM_VERSION_LABEL = `v${MINIPROGRAM_VERSION} · ${MINIPROGRAM_BUILD_SHORT}${MINIPROGRAM_BUILD_DIRTY ? " · dirty" : ""}`;
