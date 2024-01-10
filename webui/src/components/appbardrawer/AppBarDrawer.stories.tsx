// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react'

import { Badge, List, Typography } from '@mui/material'
import HomeIcon from "@mui/icons-material/Home"
import AnnouncementIcon from "@mui/icons-material/Announcement"

import AppBarDrawer from './AppBarDrawer'
import { DrawerLinkItem } from './DrawerLinkItem'

const meta: Meta<typeof AppBarDrawer> = {
    title: 'Universal/AppBarDrawer',
    component: AppBarDrawer,
    argTypes: {
        title: { control: false },
        drawertitle: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof AppBarDrawer>

export const Basic: Story = {
    args: {
        title: <Badge badgeContent={42} color='secondary'>Awfull App</Badge>,
        drawertitle: <>
            <Typography variant="h6" color="textSecondary" component="span">
                AwfullApp
            </Typography>
            <Typography variant="body2" color="textSecondary" component="span">
                &nbsp;0.0.0
            </Typography>
        </>,
        drawer: (closeDrawer) => (
            <List onClick={closeDrawer}>
                <DrawerLinkItem
                    key="home"
                    label="Home"
                    icon={<HomeIcon />}
                    path="/" />
                <DrawerLinkItem
                    key="about"
                    label="About"
                    icon={<AnnouncementIcon />}
                    path="/about" />
            </List>
        ),
    },
}
