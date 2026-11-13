import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import svgr from 'vite-plugin-svgr'
import mdx from '@mdx-js/rollup'
import path from 'path'

import { mdxConfiguration } from './src/mdxconfig.js'

const srcs = [
    'app',
    'components',
    'hooks',
    'icons',
    'models',
    'theming',
    'utils',
    'views',
]

// Ghostwire service to proxy API calls to; use http://localhost:5001 for Edgeshark.
const backend = process.env.GHOSTWIRE_BACKEND || 'http://localhost:5000'

export default defineConfig({
    base: './',
    build: {
        outDir: 'build',
        chunkSizeWarningLimit: 1024,
    },
    server: {
        host: "0.0.0.0",
        port: 3400,
        proxy: {
            '/json': backend,
            '/mobyshark': backend,
            '/discover/mobyshark': backend,
            '/mobydig': { target: backend, ws: true },
        },
    },
    resolve: {
        tsconfigPaths: true,
        alias: Object.fromEntries(
            srcs.map(d => [d, path.resolve(import.meta.dirname, `src/${d}`)])
        )
    },
    plugins: [
        {
            enforce: 'pre',
            ...mdx(mdxConfiguration)
        },
        react(),
        svgr({
            svgrOptions: {
                icon: true,
            }
        }),
    ]
})
