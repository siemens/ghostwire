// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Options } from '@mdx-js/rollup'
import type { PluginOption } from 'vite'

/** MDX compile options with Ghostwire's remark and rehype plugins. */
export declare const mdxConfiguration: Readonly<Options>

/**
 * Returns the Vite plugins required for building apps using the Ghostwire UI
 * library; apps additionally need the React plugin from @vitejs/plugin-react.
 */
export declare const ghostwireUI: () => PluginOption[]

export default ghostwireUI
