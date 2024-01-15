// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react'

import HomeIcon from '@mui/icons-material/Home'

import { DrawerLinkItem } from './DrawerLinkItem'

const meta: Meta<typeof DrawerLinkItem> = {
    title: 'Universal/DrawerLinkItem',
    component: DrawerLinkItem,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof DrawerLinkItem>

export const Basic: Story = {
    args: {
        label: 'Foo',
        path: '/foo',
    },
}

export const Icon: Story = {
    args: {
        label: 'Home',
        icon: <HomeIcon/>,
        path: '/home',
    },
}

export const Avatar: Story = {
    args: {
        label: 'Home',
        icon: <HomeIcon/>,
        path: '/home',
        avatar: true,
    },
}
