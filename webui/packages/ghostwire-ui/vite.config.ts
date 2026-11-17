// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

// Vite configuration used by Storybook.

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

import { ghostwireUI } from './vite.js'

export default defineConfig({
    plugins: [
        react(),
        ghostwireUI(),
    ],
})
