// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

// The app configuration module, which the Vite configuration resolves to either
// src/defaultconfig.ts or the module specified in VITE_APP_CONFIG.
declare module 'ghostwire-app-config' {
    import type { GhostwireAppConfig } from 'ghostwire-ui'
    const config: GhostwireAppConfig
    export default config
}
