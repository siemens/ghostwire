// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import intro from 'components/helpviewer/01-intro.mdx'
import foobar from 'components/helpviewer/02-foobar.mdx'

import { GwMarkdown } from './GwMarkdown'

const meta: Meta<typeof GwMarkdown> = {
    title: 'Ghostwire/GwMarkdown',
    component: GwMarkdown,
    argTypes: {
        mdx: { control: false },
        shortcodes: { control: false },
        fallback: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof GwMarkdown>

export const Short: Story = {
    args: {
        mdx: intro,
    },
}

export const Long: Story = {
    args: {
        mdx: foobar,
    },
}
