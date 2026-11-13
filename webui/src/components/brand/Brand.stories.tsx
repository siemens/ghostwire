// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { DynVarsProvider } from 'components/dynvars'

import { Brand } from './Brand'

const meta: Meta<typeof Brand> = {
    title: 'Ghostwire/Brand',
    component: Brand,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Brand>

export const Default: Story = {}

export const Rebranded: Story = {
    decorators: [
        (Story) => <DynVarsProvider value={{ brand: 'Edgeshark' }}><Story /></DynVarsProvider>,
    ],
}
