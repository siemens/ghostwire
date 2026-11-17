// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

import { ghostwireUI } from 'ghostwire-ui/vite'

export default defineConfig({
    base: './',
    build: {
        outDir: 'build',
        chunkSizeWarningLimit: 1024,
    },
    resolve: {
        // resolve the dependencies of the library linked via "portal:" from
        // this app's node_modules.
        preserveSymlinks: true,
    },
    plugins: [
        react(),
        ghostwireUI(),
    ],
})
