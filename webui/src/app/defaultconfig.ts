// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { GhostwireAppConfig } from './config'

/**
 * The default app configuration, which is used unless the VITE_APP_CONFIG
 * environment variable points to a different configuration module at build
 * time.
 */
const config: GhostwireAppConfig = {}

export default config
