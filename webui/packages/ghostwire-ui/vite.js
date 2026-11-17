// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

// Vite preset for building apps from the Ghostwire UI library sources; plain
// JavaScript, as Node doesn't strip types from modules inside node_modules.

import { createRequire } from 'node:module'
import path from 'node:path'
import { pathToFileURL } from 'node:url'

// Imports the default export of the named build-time dependency. When the
// library is linked from outside the app, such as via "portal:" to a sibling
// checkout, its dependencies are installed in the app's node_modules instead,
// so then resolve them from the app's directory.
const load = async (name) => {
    let mod
    try {
        mod = await import(name)
    } catch (err) {
        if (err?.code !== 'ERR_MODULE_NOT_FOUND') {
            throw err
        }
        const require = createRequire(path.join(process.cwd(), 'package.json'))
        mod = await import(pathToFileURL(require.resolve(name)).href)
    }
    const def = mod.default
    return typeof def === 'function' || typeof def?.default !== 'function' ? def : def.default
}

const [
    mdx, svgr,
    remarkGfm, remarkImages, remarkTextr, remarkGEmoji,
    textrTypoApos, textrTypoQuotes, textrTypoPossPluralsApos, textrTypoEllipses, textrTypoNumberEnDashes,
    rehypeSlug,
] = await Promise.all([
    '@mdx-js/rollup', 'vite-plugin-svgr',
    'remark-gfm', 'remark-images', 'remark-textr', 'remark-gemoji',
    'typographic-apostrophes', 'typographic-quotes', 'typographic-apostrophes-for-possessive-plurals',
    'typographic-ellipses', 'typographic-en-dashes',
    'rehype-slug',
].map(load))

// There's only typographic-em-dashes that covers US typographic style, but no
// need for a full-blown npm module just to get European en dash typography.
const textrTypoEnDashes = (input) => {
    return input
        .replace(/ -- /gim, ' – ')
}

/** @type {Readonly<import('@mdx-js/rollup').Options>} */
export const mdxConfiguration = {
    remarkPlugins: [
        remarkGfm,
        remarkImages,
        remarkGEmoji,
        [remarkTextr, {
            plugins: [
                textrTypoApos,
                textrTypoQuotes,
                textrTypoPossPluralsApos,
                textrTypoEllipses,
                textrTypoNumberEnDashes,
                textrTypoEnDashes,
            ],
            options: {
                locale: 'en-us'
            }
        }],
    ],
    rehypePlugins: [
        rehypeSlug,
    ],
}

// Packages that must be present only once in an app, even when the library
// gets linked from outside the app, such as via "portal:".
const singletons = [
    'react',
    'react-dom',
    'react-router-dom',
    '@mui/material',
    '@mui/system',
    '@mui/icons-material',
    '@emotion/react',
    '@emotion/styled',
    'jotai',
    'notistack',
]

/**
 * Returns the Vite plugins required for building apps using the Ghostwire UI
 * library: compiling the MDX help and about contents, SVG React components,
 * and deduplicating singleton dependencies. Apps additionally need to use
 * the React plugin from @vitejs/plugin-react.
 *
 * @returns {import('vite').PluginOption[]}
 */
export const ghostwireUI = () => [
    {
        enforce: 'pre',
        ...mdx(mdxConfiguration),
    },
    svgr({
        svgrOptions: {
            icon: true,
        },
    }),
    {
        name: 'ghostwire-ui',
        config: () => ({
            resolve: {
                dedupe: singletons,
            },
        }),
    },
]

export default ghostwireUI
