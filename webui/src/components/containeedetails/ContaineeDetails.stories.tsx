// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { initialNetns } from 'models/gw/mock'
import { boundBox, containerBox } from 'views/help/mockdata'

import { ContaineeDetails } from './ContaineeDetails'

const meta: Meta<typeof ContaineeDetails> = {
    title: 'Ghostwire/ContaineeDetails',
    component: ContaineeDetails,
    args: {
        containee: containerBox,
    },
    argTypes: {
        containee: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ContaineeDetails>

export const Container: Story = {}

export const IPv4Only: Story = {
    name: 'IPv4 Only',
    args: {
        families: [AddressFamily.IPv4],
    },
}

export const InitialProcess: Story = {
    args: {
        containee: initialNetns.containers[0],
    },
}

export const BindMount: Story = {
    args: {
        containee: boundBox,
    },
}
