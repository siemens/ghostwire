// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { BrandIcon } from './BrandIcon'

const meta: Meta<typeof BrandIcon> = {
    title: 'Ghostwire/BrandIcon',
    component: BrandIcon,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof BrandIcon>

export const Standard: Story = {}

export const Large: Story = {
    args: {
        fontSize: 'large',
    },
}
