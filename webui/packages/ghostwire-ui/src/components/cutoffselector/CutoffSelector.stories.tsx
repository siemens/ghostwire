// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { atom } from 'jotai'

import { CutoffSelector, NEVER } from './CutoffSelector'

const meta: Meta<typeof CutoffSelector> = {
    title: 'Ghostwire/CutoffSelector',
    component: CutoffSelector,
    argTypes: {
        atom: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof CutoffSelector>

export const Standard: Story = {
    args: {
        atom: atom(3),
        element: 'container',
    },
}

export const IrregularPlural: Story = {
    args: {
        atom: atom(10),
        element: 'process',
        elements: 'processes',
        em: 12,
    },
}

export const NonStandardCutoff: Story = {
    args: {
        atom: atom(7),
        element: 'route',
    },
}

export const ShowAlways: Story = {
    args: {
        atom: atom(NEVER),
        element: 'network interface',
    },
}
