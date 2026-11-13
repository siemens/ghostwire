// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { useRef } from 'react'

import type { Meta, StoryObj } from '@storybook/react-vite'

import { SnackbarProvider } from 'notistack'

import { Box, Button, Card } from '@mui/material'

import { useScreenShooterModal } from './context'
import { ScreenShooter } from './ScreenShooter'

const Shootable = () => {
    const ref = useRef<HTMLDivElement>(null)
    const setModal = useScreenShooterModal()
    return <>
        <Box sx={{ p: 2 }} ref={ref}>
            <Card>
                <Box sx={{ p: 2 }}>FOO<b>BAR</b>!</Box>
            </Card>
        </Box>
        <Button variant="outlined" onClick={() => setModal?.(ref.current)}>
            Capture
        </Button>
    </>
}

const meta: Meta<typeof ScreenShooter> = {
    title: 'Universal/ScreenShooter',
    component: ScreenShooter,
    args: {
        children: <Shootable />,
    },
    argTypes: {
        children: { control: false },
    },
    decorators: [
        (Story) => <SnackbarProvider maxSnack={3}><Story /></SnackbarProvider>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ScreenShooter>

export const Standard: Story = {}

export const CustomBasename: Story = {
    args: {
        basename: 'foobar',
    },
}
