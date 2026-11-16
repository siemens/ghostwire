// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { ErrorNotice } from './ErrorFallback'

const meta: Meta<typeof ErrorNotice> = {
    title: 'Universal/ErrorNotice',
    component: ErrorNotice,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ErrorNotice>

export const Standard: Story = {
    args: {
        title: "This view couldn't be shown",
        error: new Error('lorem ipsum dolor sit amet'),
        onRetry: () => { },
    },
}

export const WithReload: Story = {
    args: {
        title: "This content couldn't be loaded",
        error: new TypeError('Failed to fetch dynamically imported module'),
        onRetry: () => { },
        reload: true,
    },
}
