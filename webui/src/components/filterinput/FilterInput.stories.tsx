// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { FilterInput } from './FilterInput'

const meta: Meta<typeof FilterInput> = {
    title: 'Ghostwire/FilterInput',
    component: FilterInput,
    args: {
        onChange: fn(),
        onEnter: fn(),
    },
    argTypes: {
        focusRef: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof FilterInput>

export const Empty: Story = {}

export const Substring: Story = {
    args: {
        filterPattern: { pattern: 'moby', isCaseSensitive: false, isRegexp: false },
    },
}

export const CaseSensitiveRegexp: Story = {
    args: {
        filterPattern: { pattern: '^Morbid_.*_1$', isCaseSensitive: true, isRegexp: true },
    },
}
