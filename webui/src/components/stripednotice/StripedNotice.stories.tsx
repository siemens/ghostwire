// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { StripedNotice } from './StripedNotice'

const meta: Meta<typeof StripedNotice> = {
    title: 'Universal/StripedNotice',
    component: StripedNotice,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof StripedNotice>

export const Standard: Story = {
    args: {
        children: 'Under construction: lorem ipsum dolor sit amet.',
    },
}
