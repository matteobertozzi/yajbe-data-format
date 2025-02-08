
import dts from 'bun-plugin-dts';

await Promise.all([
  Bun.build({
    entrypoints: ['../yajbe.ts'],
    external: [],
    outdir: './dist',
    minify: false,
    plugins: [dts()],
    format: 'esm',
    naming: "[dir]/[name].js",
  }),
])