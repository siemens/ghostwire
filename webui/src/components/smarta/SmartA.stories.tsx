// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react'

import { SmartA } from './SmartA'

const meta: Meta<typeof SmartA> = {
    title: 'Universal/SmartA',
    component: SmartA,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof SmartA>

export const External: Story = {
    args: {
        href: 'https://github.com/thediveo/lxkns',
        children: '@thediveo/lxkns',
    },
}

export const Internal: Story = {
    args: {
        href: '/internal',
        children: 'an app-internal route',
    },
}