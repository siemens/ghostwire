// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react'

import { ExtLink } from './ExtLink'

const meta: Meta<typeof ExtLink> = {
    title: 'Universal/ExtLink',
    component: ExtLink,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ExtLink>

export const Standard: Story = {
    args: {
        href: 'https://github.com/thediveo/lxkns',
        children: '@thediveo/lxkns',
    },
}

export const After: Story = {
    args: {
        iconposition: 'after',
        href: 'https://github.com/thediveo/lxkns',
        children: '@thediveo/lxkns',
    },
}