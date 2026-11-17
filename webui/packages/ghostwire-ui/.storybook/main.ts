// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { StorybookConfig as StorybookViteConfig } from '@storybook/react-vite'
import { mdxConfiguration } from '../vite.js'
import type { Plugin, PluginOption } from 'vite'

// flattens nested plugin options, such as returned by the ghostwireUI() preset.
const flattened = (plugins: PluginOption[]): PluginOption[] =>
    plugins.flatMap(p => Array.isArray(p) ? flattened(p) : [p])

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

    features: {
        // don't annoy with the "Get started" onboarding checklist in the
        // sidebar.
        sidebarOnboardingChecklist: false,
    },

    typescript: {
        check: true,
    },

    async viteFinal(config) {
        // drop the @mdx-js/rollup plugin that we get from the vite
        // configuration, as this otherwise causes problems with the mdx plugin
        // brought in by @storybook/addon-docs.
        config.plugins = flattened(config.plugins ?? []).filter(e => (e as Plugin)?.name !== '@mdx-js/rollup')
        return config
    },

}

export default config