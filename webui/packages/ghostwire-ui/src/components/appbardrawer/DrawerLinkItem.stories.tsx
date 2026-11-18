// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react'

import { Box } from '@mui/material'
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

/** This story renders two `DrawerLinkItem`s, the second being an `avatar` item. */
export const PlainAndAvatar: Story = {
    render: () => <>
        <Box sx={{ width: '20em', m: 1 }}>
            <DrawerLinkItem label="Home" icon={<HomeIcon />} path="/home" />
        </Box>
        <Box sx={{ width: '20em', m: 1 }}>
            <DrawerLinkItem avatar label="Home" icon={<HomeIcon />} path="/home" />
        </Box>
    </>,
}
