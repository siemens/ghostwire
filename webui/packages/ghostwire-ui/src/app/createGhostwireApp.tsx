// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import App from './App'
import type { GhostwireAppConfig } from './config'

/**
 * Returns the Ghostwire app component, configured with the specified MUI
 * themes and branding; missing configuration elements default to Ghostwire's
 * defaults.
 *
 * ```tsx
 * const GhostwireApp = createGhostwireApp({
 *     themes: { light: { palette: { primary: { main: '#1565c0' } } } },
 *     branding: { name: 'Foobar' },
 * })
 * createRoot(container).render(<GhostwireApp />)
 * ```
 */
export const createGhostwireApp = (config: GhostwireAppConfig = {}) => {
    const GhostwireApp = () => <App config={config} />
    return GhostwireApp
}
