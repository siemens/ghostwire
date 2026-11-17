// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Provider } from 'jotai'

import RefreshButton from './RefreshButton'

const meta: Meta<typeof RefreshButton> = {
    title: 'Ghostwire/RefreshButton',
    component: RefreshButton,
    decorators: [
        (Story) => <Provider><Story /></Provider>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof RefreshButton>

export const Standard: Story = {}
