// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { StorybookConfig as StorybookViteConfig } from '@storybook/react-vite'
import { mdxConfiguration } from '../src/mdxconfig.ts'
import type { Plugin } from 'vite'

const config: StorybookViteConfig = {
    framework: {
        name: '@storybook/react-vite',
        options: {},
    },

    stories: [
        '../src/**/*.stories.@(ts|tsx)',
        '../src/*.mdx',
    ],

    addons: [
        {
            name: '@storybook/addon-docs',
            options: {
                mdxPluginOptions: {
                    mdxCompileOptions: {
                        ...mdxConfiguration,
                    }
                },
            }
        },
        '@storybook/addon-links',
    ],

    docs: {
        defaultName: 'Description',
    },

    core: {
        disableTelemetry: true,
        disableWhatsNewNotifications: true,
    },

    typescript: {
        check: true,
    },

    async viteFinal(config) {
        // drop the @mdx-js/rollup plugin that we get from the vite
        // configuration, as this otherwise causes problems with the mdx plugin
        // brought in by @storybook/addon-docs.
        config.plugins = config.plugins?.filter(e => (e as Plugin)?.name !== '@mdx-js/rollup')
        return config
    },

}

export default config