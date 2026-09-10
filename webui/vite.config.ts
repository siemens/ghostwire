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
            '/json': 'http://localhost:5001',
            '/mobyshark': 'http://localhost:5001',
            '/discover/mobyshark': 'http://localhost:5001',
            '/mobydig': 'http://localhost:5001',
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
