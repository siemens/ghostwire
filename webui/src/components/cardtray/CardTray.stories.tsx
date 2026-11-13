// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Card, CardContent } from '@mui/material'

import { CardTray } from './CardTray'

const meta: Meta<typeof CardTray> = {
    title: 'Universal/CardTray',
    component: CardTray,
    args: {
        children: [1, 2, 3].map(no =>
            <Card key={no}><CardContent>card #{no}</CardContent></Card>),
    },
    argTypes: {
        children: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof CardTray>

export const Standard: Story = {}

export const Animated: Story = {
    args: {
        animate: true,
    },
}
