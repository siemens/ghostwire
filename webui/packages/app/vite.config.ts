// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import { defineConfig, loadEnv, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import { readFileSync } from 'fs'
import path from 'path'

import { ghostwireUI } from 'ghostwire-ui/vite'

// The yarn workspace root (webui/), where the static build goes to, so that it
// can get embedded into the Gostwire service binary.
const webuiRoot = path.resolve(import.meta.dirname, '../..')

// Ghostwire service to proxy API calls to; use http://localhost:5001 for Edgeshark.
const backend = process.env.GHOSTWIRE_BACKEND || 'http://localhost:5000'

// Replaces the index.html page shell with the specified file, before Vite
// processes it, so that its scripts and asset references get bundled as usual.
const pageShell = (file: string): Plugin => ({
    name: 'ghostwire-app-page-shell',
    transformIndexHtml: {
        order: 'pre',
        handler: () => readFileSync(file, 'utf8'),
    },
})

export default defineConfig(({ mode }) => {
    // Downstream builds can point VITE_APP_CONFIG to their own app configuration
    // module (default-exporting a GhostwireAppConfig), VITE_APP_PUBLIC_DIR to
    // their own static assets, such as favicon and app icons, and
    // VITE_APP_INDEX_HTML to their own index.html page shell. These are taken
    // from the environment or the .env files, relative to the webui/ directory.
    const env = loadEnv(mode, import.meta.dirname, 'VITE_APP_')
    const resolved = (envpath: string | undefined, fallback: string) =>
        envpath ? path.resolve(webuiRoot, envpath) : path.resolve(import.meta.dirname, fallback)

    return {
        base: './',
        publicDir: resolved(env.VITE_APP_PUBLIC_DIR, 'public'),
        build: {
            outDir: path.join(webuiRoot, 'build'),
            emptyOutDir: true,
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
            alias: {
                'ghostwire-app-config': resolved(env.VITE_APP_CONFIG, 'src/defaultconfig.ts'),
            },
        },
        plugins: [
            react(),
            ghostwireUI(),
            env.VITE_APP_INDEX_HTML ? pageShell(path.resolve(webuiRoot, env.VITE_APP_INDEX_HTML)) : null,
        ],
    }
})
