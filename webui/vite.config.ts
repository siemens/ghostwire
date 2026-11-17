import { defineConfig, loadEnv } from 'vite'
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

export default defineConfig(({ mode }) => {
    // Downstream builds can point VITE_APP_CONFIG to their own app configuration
    // module (default-exporting a GhostwireAppConfig), and VITE_APP_PUBLIC_DIR to
    // their own static assets, such as favicon and app icons. Both are taken
    // from the environment or the .env files, relative to this directory.
    const env = loadEnv(mode, import.meta.dirname, 'VITE_APP_')
    const resolved = (envpath: string | undefined, fallback: string) =>
        path.resolve(import.meta.dirname, envpath || fallback)

    return {
        base: './',
        publicDir: resolved(env.VITE_APP_PUBLIC_DIR, 'public'),
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
            alias: {
                ...Object.fromEntries(
                    srcs.map(d => [d, path.resolve(import.meta.dirname, `src/${d}`)])
                ),
                'ghostwire-app-config': resolved(env.VITE_APP_CONFIG, 'src/app/defaultconfig.ts'),
            },
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
        ],
    }
})
