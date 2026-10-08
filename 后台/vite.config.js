import { defineConfig } from 'vite'

// The Go embed handler mounts this application below /admin. Keeping the
// base explicit prevents production assets from resolving to the user SPA's
// /assets directory when an administrator refreshes a nested route.
export default defineConfig({
  base: '/admin/',
})
