// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { discovery } from 'models/gw/mock'

import { ContaineeNavigator } from './ContaineeNavigator'

const meta: Meta<typeof ContaineeNavigator> = {
    title: 'Ghostwire/ContaineeNavigator',
    component: ContaineeNavigator,
    args: {
        allnetns: discovery.networkNamespaces,
    },
    argTypes: {
        allnetns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ContaineeNavigator>

export const Standard: Story = {}

export const FilterEmpty: Story = {
    args: {
        filterEmpty: true,
    },
}

export const NoLinks: Story = {
    args: {
        nolink: true,
    },
}
