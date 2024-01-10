// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react'

import { ChapterSkeleton } from './ChapterSkeleton'

const meta: Meta<typeof ChapterSkeleton> = {
    title: 'Universal/ChapterSkeleton',
    component: ChapterSkeleton,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ChapterSkeleton>

export const Basic: Story = {
    args: {
        sx: { width: "20rem" },
    },
}
